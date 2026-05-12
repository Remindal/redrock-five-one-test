package dal

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	RabbitConn *amqp.Connection
	RabbitCh   *amqp.Channel
)

type SeckillMessage struct {
	ActivityId int64  `json:"activity_id"`
	UserId     string `json:"user_id"`
}

func InitRabbitMQ(addr string) {
	var err error
	RabbitConn, err = amqp.Dial(addr)
	if err != nil {
		log.Fatalf("rabbitmq connect error: %v", err)
	}

	RabbitCh, err = RabbitConn.Channel()
	if err != nil {
		log.Fatalf("rabbitmq channel error: %v", err)
	}

	_, err = RabbitCh.QueueDeclare(
		"seckill_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("rabbitmq queue declare error: %v", err)
	}

	log.Println("rabbitmq connected")
}

func SendSeckillMessage(ctx context.Context, activityId int64, userId string) error {
	ch, err := RabbitConn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	msg := SeckillMessage{
		ActivityId: activityId,
		UserId:     userId,
	}
	body, _ := json.Marshal(msg)

	return ch.PublishWithContext(ctx,
		"",
		"seckill_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: 2, // 2=持久化消息（重启不丢）
		},
	)
}
