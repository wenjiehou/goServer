package model

import "time"

type LogItem struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	UserID int `json:"user_id" sql:"index"`
	ItemId int `json:"item_id" sql:"index"`
	Count  int `json:"count"`
	Type   int `json:"type"`
}

type LogItemModel struct {
	CommonModel
}

func GetLogItemModel() *LogItemModel {
	return &LogItemModel{CommonModel{db: commonDb}}
}