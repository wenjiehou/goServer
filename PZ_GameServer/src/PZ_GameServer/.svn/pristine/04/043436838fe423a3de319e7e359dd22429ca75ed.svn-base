package model

import (
	"time"
)

type Notice struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	Title        string    `json:"account"`
	Content      string    `json:"cargo_id"`
	Status       int       `json:"status"` //0:无效 1：有效
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	AreaIds      string    `json:"area_ids"`
	Type         int       `json:"type" sql:"index"` //1:跑马灯 2：停服公告,
	ShowPosition int       `json:"show_position"`
	ChannelId    int       `json:"channel_id" sql:"index"`
	IsTop        int       `json:"is_top"`
}

type NoticeModel struct {
	CommonModel
}

func GetNoticeModel() *NoticeModel {
	return &NoticeModel{CommonModel{db: commonDb}}
}

func (u *NoticeModel) GetNotices(id uint) ([]*Notice, error) {
	var notices []*Notice

	if err := u.db.Model(&Notice{}).Where("type = 1 and id >? and status = 1 and end_time>? and start_time <?",
		id, time.Now(), time.Now()).Find(&notices).Error; err != nil {
		return nil, err
	}

	return notices, nil
}

func (u *NoticeModel) GetStopGameNotice() (*Notice, error) {

	notice := Notice{}
	if err := u.db.Model(&Notice{}).Where("type = 2 and status = 1 and end_time>? and start_time <?",
		time.Now(), time.Now()).Last(&notice).Error; err != nil {
		return nil, err
	}
	return &notice, nil
}
