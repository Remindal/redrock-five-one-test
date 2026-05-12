package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	gatewayURL  = "http://127.0.0.1:8888"
	activityId  = int64(10)
	concurrency = 5000
	duration    = 10 * time.Second
)

func main() {
	transport := &http.Transport{
		MaxIdleConns:        2000,
		MaxIdleConnsPerHost: 2000,
		MaxConnsPerHost:     2000,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   1 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	// Step 1: 批量登录获取 Token（每个 worker 一个独立用户）
	fmt.Printf("正在批量登录 %d 个用户获取 Token...\n", concurrency)
	tokens := make([]string, concurrency)
	var loginWg sync.WaitGroup
	var loginFail int64
	for i := 0; i < concurrency; i++ {
		loginWg.Add(1)
		go func(idx int) {
			defer loginWg.Done()
			userId := fmt.Sprintf("bench_user_%d_%d", time.Now().UnixNano(), idx)
			body, _ := json.Marshal(map[string]string{"user_id": userId})
			resp, err := client.Post(gatewayURL+"/api/auth/login", "application/json", bytes.NewReader(body))
			if err != nil {
				atomic.AddInt64(&loginFail, 1)
				return
			}
			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)
			resp.Body.Close()
			if code, ok := result["code"].(float64); ok && int(code) == 200 {
				if t, ok := result["token"].(string); ok {
					tokens[idx] = t
				}
			}
		}(i)
	}
	loginWg.Wait()
	validTokens := 0
	for _, t := range tokens {
		if t != "" {
			validTokens++
		}
	}
	fmt.Printf("登录完成: 成功 %d / 目标 %d (失败 %d)\n", validTokens, concurrency, loginFail)
	if validTokens == 0 {
		fmt.Println("无可用 Token，压测终止")
		return
	}

	// Step 2: 压测秒杀接口
	fmt.Printf("开始压测: 并发 %d, 持续 %v, 活动 ID %d\n", validTokens, duration, activityId)
	url := gatewayURL + "/api/seckill/do"
	var wg sync.WaitGroup
	start := time.Now()
	var success, fail, repeat, stockout, ratelimit, busy, unauthorized int64

	for i := 0; i < concurrency; i++ {
		if tokens[i] == "" {
			continue
		}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			token := tokens[idx]
			for time.Since(start) < duration {
				body, _ := json.Marshal(map[string]int64{"activity_id": activityId})
				req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)

				resp, err := client.Do(req)
				if err != nil {
					atomic.AddInt64(&fail, 1)
					continue
				}

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				resp.Body.Close()

				code, _ := result["code"].(float64)
				switch int(code) {
				case 200:
					atomic.AddInt64(&success, 1)
				case 4005:
					atomic.AddInt64(&stockout, 1)
				case 4006:
					atomic.AddInt64(&repeat, 1)
				case 4007:
					atomic.AddInt64(&busy, 1)
				case 4008:
					atomic.AddInt64(&ratelimit, 1)
				case 4009:
					atomic.AddInt64(&unauthorized, 1)
				default:
					atomic.AddInt64(&fail, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	total := success + repeat + stockout + fail + ratelimit + busy + unauthorized
	fmt.Println()
	fmt.Printf("总时间:     %.2fs\n", elapsed.Seconds())
	fmt.Printf("成功:       %d\n", success)
	fmt.Printf("库存不足:   %d\n", stockout)
	fmt.Printf("重复:       %d\n", repeat)
	fmt.Printf("系统繁忙:   %d\n", busy)
	fmt.Printf("限流拒绝:   %d\n", ratelimit)
	fmt.Printf("未授权:     %d\n", unauthorized)
	fmt.Printf("失败/超时:  %d\n", fail)
	fmt.Printf("总请求:     %d\n", total)
	fmt.Printf("QPS:        %.2f\n", float64(total)/elapsed.Seconds())

	// Step 3: 等待 MQ 消费完成
	fmt.Println("\n等待 MQ 消费完成...")
	time.Sleep(30 * time.Second)

	// Step 4: 一致性检查
	fmt.Println("\n========== 一致性检查 ==========")
	checkConsistency(activityId, success)
}

func checkConsistency(activityId int64, successCount int64) {
	ctx := context.Background()

	// 1. Redis 库存
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	stockKey := fmt.Sprintf("seckill:stock:%d", activityId)
	stock, err := rdb.Get(ctx, stockKey).Int64()
	if err == redis.Nil {
		stock = 0
	} else if err != nil {
		fmt.Printf("Redis 查询错误: %v\n", err)
		stock = -1
	}
	fmt.Printf("Redis 剩余库存: %d (期望: 0)\n", stock)

	// 2. MySQL 订单总数
	db, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:3307)/seckill?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		fmt.Printf("MySQL 连接错误: %v\n", err)
		return
	}

	var orderCount int64
	db.Raw("SELECT COUNT(*) FROM `order` WHERE activity_id = ?", activityId).Scan(&orderCount)
	fmt.Printf("数据库订单数: %d (期望: %d)\n", orderCount, successCount)

	// 3. 重复用户检查
	var uniqueCount int64
	db.Raw("SELECT COUNT(DISTINCT user_id) FROM `order` WHERE activity_id = ?", activityId).Scan(&uniqueCount)
	fmt.Printf("独立用户数: %d (重复: %v)\n", uniqueCount, uniqueCount != orderCount)

	// 4. 结论
	fmt.Println("\n========== 检查结论 ==========")
	pass := true
	if stock < 0 {
		fmt.Printf("❌ Redis 库存为负 (实际: %d)，存在超卖\n", stock)
		pass = false
	} else {
		fmt.Printf("✅ Redis 库存正常 (剩余: %d，无超卖)\n", stock)
	}
	if orderCount != successCount {
		fmt.Printf("❌ 数据库订单数与成功数不一致 (订单: %d, 成功: %d)\n", orderCount, successCount)
		pass = false
	} else {
		fmt.Println("✅ 订单数与成功数一致")
	}
	if uniqueCount != orderCount {
		fmt.Printf("❌ 存在重复用户订单 (独立: %d, 总数: %d)\n", uniqueCount, orderCount)
		pass = false
	} else {
		fmt.Println("✅ 无重复用户订单")
	}
	if pass {
		fmt.Println("\n🎉 一致性检查全部通过")
	} else {
		fmt.Println("\n⚠️ 一致性检查未通过")
	}
}
