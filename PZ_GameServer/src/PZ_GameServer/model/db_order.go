package model

import (
	"time"
)

type Order struct {
	ID        uint      `gorm:"primary_key"`
	CreatedAt time.Time `sql:"index"`
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	UserID    int    `json:"user_id" sql:"index"`
	ChannelID int    `json:"channel_id" sql:"index"`
	AccountID int    `json:"account_id"`
	OrderNo   string `json:"order_no"`
	CargoId   int    `json:"cargo_id"` //
	RMB       int    `json:"rmb"`
	Diamond   int    `json:"diamond"`
	Status    int    `json:"status" sql:"index"`
	PayType   int    `json:"pay_type"`
}

type OrderModel struct {
	CommonModel
}

func GetOrderModel() *OrderModel {
	return &OrderModel{CommonModel{db: commonDb}}
}
