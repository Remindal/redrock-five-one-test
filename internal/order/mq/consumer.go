package mq

import (
	"context"
	"encoding/json"
	"log"
	"seckill-system/internal/order/dao"
	"seckill-system/internal/order/service"
	"seckill-system/kitex_gen/order"

	amqp "github.com/rabbitmq/amqp091-go"
)

func StartConsumer() {
	msgs, err := dal.RabbitCh.Consume(
		"seckill_queue",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("rabbitmq consume error: %v", err)
	}

	for d := range msgs {
		processMessage(d)
	}
}

func processMessage(d amqp.Delivery) {
	var msg dal.SeckillMessage
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		log.Printf("msg unmarshal error: %v, body: %s", err, string(d.Body))
		_ = d.Ack(false)
		return
	}

	resp, err := service.NewOrderService(context.Background()).CreateOrder(&order.CreateOrderReq{
		ActivityId: msg.ActivityId,
		UserId:     msg.UserId,
	})
	if err != nil {
		log.Printf("create order service error: %v", err)
		_ = d.Nack(false, true)
		return
	}

	if resp.Code == 200 {
		_ = d.Ack(false)
		log.Printf("order processed success: orderId=%d", resp.GetOrderId())
		return
	}

	// 其他错误，Nack 让消息重新入队重试
	log.Printf("create order failed: code=%d, msg=%s", resp.Code, resp.Msg)
	_ = d.Nack(false, true)
}
