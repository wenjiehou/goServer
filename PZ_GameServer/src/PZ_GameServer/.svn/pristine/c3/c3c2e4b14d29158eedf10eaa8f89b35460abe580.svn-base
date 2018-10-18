package model

import "time"

const (
	DiamondCreateRoomOnePay  = iota + 1 //一人付创建房间
	DiamondCreateRoomFourPay            //四人付创建房间
	DiamondCreateRoomReturn             //创建房间返还
)

type LogDiamond struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	UserID  int `json:"user_id" sql:"index"`
	Diamond int `json:"diamond"`
	Type    int `json:"type"`
}

type LogDiamondModel struct {
	CommonModel
}

func GetLogDiamondModel() *LogDiamondModel {
	return &LogDiamondModel{CommonModel{db: commonDb}}
}

func (l *LogDiamondModel) Create(logDiamond *LogDiamond) error {
	if err := l.db.Model(&LogDiamond{}).Create(logDiamond).Error; err != nil {
		return err
	}
	return nil
}
