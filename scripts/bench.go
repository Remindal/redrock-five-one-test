package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	gatewayURL  = "http://127.0.0.1:8888"
	activityId  = int64(7)
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
}
