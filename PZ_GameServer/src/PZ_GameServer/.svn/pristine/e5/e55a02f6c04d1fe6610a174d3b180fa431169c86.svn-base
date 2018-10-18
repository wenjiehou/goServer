package model

import (
	"time"
)

type LogLogin struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	AreaID int    `json:"area_id"`
	UserID string `json:"user_id" sql:"index"`
	Ip     string `json:"login_ip"`
}

type LogLoginModel struct {
	CommonModel
}

func GetLogLoginModel() *LogLoginModel {
	return &LogLoginModel{CommonModel{db: commonDb}}
}