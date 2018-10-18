package model

import (
	"time"
)

type Account struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	ChannelID int    `json:"channel"`
	Account   string `json:"account"`
	PassWord  string `json:"pass_word"`
}

type AccountModel struct {
	CommonModel
}

func GetAccountModel() *AccountModel {
	return &AccountModel{CommonModel{db: commonDb}}
}
