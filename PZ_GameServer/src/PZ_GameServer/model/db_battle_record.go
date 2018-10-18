package model

import (
	"fmt"
	"time"
)

type BattleRecord struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	RoomID     int    `json:"room_id" sql:"index"`
	Round      int    `json:"round"`
	Result     IntKv  `json:"result"`      //user_id,diamond
	PlayBack   string `json:"play_back"`   //牌局步骤
	ReviewCode string `json:"review_code"` //回放码

	Room Room `json:"room" gorm:"ForeignKey:RoomID"`
}

type BattleRecordModel struct {
	CommonModel
}

func GetBattleRecordModel() *BattleRecordModel {
	return &BattleRecordModel{CommonModel{db: commonDb}}
}

func (u *BattleRecordModel) Create(record *BattleRecord) error {
	if err := u.db.Model(&BattleRecord{}).Create(record).Error; err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (u *BattleRecordModel) GetBattleRecordById(id int) (*BattleRecord, error) {
	var record BattleRecord
	if err := u.db.Model(&BattleRecord{}).Where("id=?", id).Find(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (u *BattleRecordModel) GetBattleRecordByReviewCode(code string) (*BattleRecord, error) {
	var record BattleRecord
	if err := u.db.Model(&BattleRecord{}).Where("review_code=?", code).Find(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (u *BattleRecordModel) GetBattleRecordByRoomId(id int32) ([]BattleRecord, error) {
	var records []BattleRecord
	if err := u.db.Model(&BattleRecord{}).Where("room_id=?", id).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}
