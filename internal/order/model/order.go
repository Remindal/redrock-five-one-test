package model

import "time"

type Order struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ActivityId int64     `gorm:"column:activity_id"`
	UserId     string    `gorm:"column:user_id"`
	Status     string    `gorm:"column:status;default:PROCESSING"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Order) TableName() string {
	return "order"
}
