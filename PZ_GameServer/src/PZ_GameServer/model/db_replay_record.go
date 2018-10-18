package model

import (
	"fmt"
	"time"
)

type ReplayRecord struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	PlayBack   string `json:"play_back"`   //牌局步骤
	ReviewCode string `json:"review_code"` //回放码
}

type ReplayRecordModel struct {
	CommonModel
}

func GetReplayRecordModel() *ReplayRecordModel {
	return &ReplayRecordModel{CommonModel{db: commonDb}}
}

func (u *ReplayRecordModel) Create(record *ReplayRecord) error {
	if err := u.db.Model(&ReplayRecord{}).Create(record).Error; err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (u *ReplayRecordModel) GetReplayRecordByReviewCode(code string) (*ReplayRecord, error) {
	var record ReplayRecord
	if err := u.db.Model(&ReplayRecord{}).Where("review_code=?", code).Find(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}
