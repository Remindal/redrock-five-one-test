package model

import "time"

type Activity struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string    `gorm:"column:name"`
	Stock       int32     `gorm:"column:stock"`
	RemainStock int32     `gorm:"column:remain_stock"`
	StartTime   time.Time `gorm:"column:start_time"`
	EndTime     time.Time `gorm:"column:end_time"`
	Status      int32     `gorm:"column:status;default:1"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Activity) TableName() string {
	return "activity"
}
