package model

import (
	"time"
)

type Mail struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	FromID  int    `json:"from_id" sql:"index"`
	ToID    int    `json:"to_id" sql:"index"`
	Title   string `json:"title"`
	Goods   string `json:"goods"`
	Context string `json:"context"`
	Watched int    `json:"watched"`
	Used    int    `json:"used"`
	Type    int    `json:"type"` //0:非系统邮件 1:系统邮件
}

type MailModel struct {
	CommonModel
}

func GetMailModel() *MailModel {
	return &MailModel{CommonModel{db: commonDb}}
}
