package model

import (
	"fmt"
	"time"
)

const (
	Lose = iota
	Win
)

type RoomRecord struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	RoomID         int `json:"room_id" sql:"index"`
	RoomType       int `json:"room_type"`
	BattleRecordID int `json:"battle_record_id" sql:"index"`
	UserID         int `json:"user_id" sql:"index"`
	Position       int `json:"position"`
	Win            int `json:"win"`

	Room         Room         `json:"room" gorm:"ForeignKey:RoomID"`
	BattleRecord BattleRecord `json:"battle_record" gorm:"ForeignKey:BattleRecordID"`
}

type Result struct {
	RoomType int
	Count    int
}

type RoomRecordModel struct {
	CommonModel
}

func GetRoomRecordModel() *RoomRecordModel {
	return &RoomRecordModel{CommonModel{db: commonDb}}
}

func (r *RoomRecordModel) Create(record *RoomRecord) error {
	if err := r.db.Model(&RoomRecord{}).Create(record).Error; err != nil {
		return err
	}
	return nil
}

func (r *RoomRecordModel) QueryAll(userId int) (map[int]int, error) {
	var results []Result
	sqlStr := fmt.Sprintf("select room_type,count(*) as count from room_records where user_id = ? and created_at > " +
		"date_sub(now(),interval 7 day) group by room_type")
	if err := r.db.Raw(sqlStr, userId).Scan(&results).Error; err != nil {
		return nil, err
	}

	resultMap := make(map[int]int)
	for _, v := range results {
		if _, ok := resultMap[v.RoomType]; !ok {
			resultMap[v.RoomType] = v.Count
		}
	}
	return resultMap, nil
}

func (r *RoomRecordModel) Query(userId, win int) (map[int]int, error) {
	var results []Result
	sqlStr := fmt.Sprintf("select room_type,count(*) as count from room_records where user_id = ? and win = ? and created_at > " +
		"date_sub(now(),interval 7 day) group by room_type")
	if err := r.db.Raw(sqlStr, userId, win).Scan(&results).Error; err != nil {
		return nil, err
	}

	resultMap := make(map[int]int)
	for _, v := range results {
		if _, ok := resultMap[v.RoomType]; !ok {
			resultMap[v.RoomType] = v.Count
		}
	}
	return resultMap, nil
}

func (r *RoomRecordModel) GetRoomIdsById(id int) ([]int, error) {
	var ids []int
	if err := r.db.Model(&RoomRecord{}).Where("user_id=?", id).Pluck("distinct(room_id)", &ids).Error; err != nil {
		return ids, err
	}
	return ids, nil
}
