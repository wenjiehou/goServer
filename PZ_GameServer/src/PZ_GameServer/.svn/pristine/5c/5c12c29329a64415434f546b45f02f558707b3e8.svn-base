package model

import (
	"time"
)

type UserDeny struct {
	ID        int `json:"id" gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	UserID      int       `json:"user_id" sql:"index"`
	DenyType    int       `json:"deny_type"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      int       `json:"status"`
	DenyTime    time.Time `json:"deny_time"`
	DenyAdminID int       `json:"deny_admin_id"`
	AllowTime   time.Time `json:"allow_time"`
	AllAdminID  int       `json:"all_admin_id"`
}

type UserDenyModel struct {
	CommonModel
}

func GetUserDenyModel() *UserDenyModel {
	return &UserDenyModel{CommonModel{db: commonDb}}
}