package model

import (
	"time"
)

type UserItem struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	UserID int `json:"user_id" sql:"index"`
	ItemID int `json:"item_id"`
	Count  int `json:"count"`
}

type UserItemModel struct {
	CommonModel
}

func GetUserItemModel() *UserItemModel {
	return &UserItemModel{CommonModel{db: commonDb}}
}
