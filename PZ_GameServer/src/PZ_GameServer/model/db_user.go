package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type User struct {
	ID        int `json:"id" gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	AccountID    int     `json:"account_id"`
	UnionID      string  `json:"union_id" sql:"unique"`
	OpenID       string  `json:"open_id"`
	NickName     string  `json:"nick_name"`
	Sex          int     `json:"sex"`
	Province     string  `json:"province"`
	City         string  `json:"city"`
	Country      string  `json:"country"`
	Sid          string  `json:"sid"`
	ChannelId    int     `json:"channel_id"`
	Type         int     `json:"type"`  //玩家类型
	State        int     `json:"state"` //玩家状态
	Coin         int     `json:"coin"`
	Card         int     `json:"card"`
	Diamond      int     `json:"diamond"`
	Email        string  `json:"email"`
	WinWeek      int     `json:"win_week"`
	WeekCount    int     `json:"week_count"`
	UsualArea    int     `json:"usual_area"`
	BenefitTime  string  `json:"benefit_time"`
	BenefitCount int     `json:"benefit_count"`
	LastIp       string  `json:"last_ip"`
	GPS_LNG      float32 `json:"gps_lng"`
	GPS_LAT      float32 `json:"gps_lat"`
	IsRobot      int     `json:"is_robot"`
	RoomId       int     `json:"room_id"`
	Icon         string  `json:"icon"`
	Level        int     `json:"level"`
	ParentUid    int     `json:"parent_uid"`
}

type UserModel struct {
	CommonModel
}

func GetUserModel() *UserModel {
	return &UserModel{CommonModel{db: commonDb}}
}

func (u *UserModel) Create(user *User) error {
	if err := u.db.Model(&User{}).Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (u *UserModel) GetUserById(id int) (*User, error) {
	var user User
	if err := u.db.Model(&User{}).Where("id=?", id).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *UserModel) GetUserByOpenId(openid string) (*User, error) {
	var user User

	if err := u.db.Model(&User{}).Where("open_id=?", openid).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *UserModel) Save(user *User) error {
	if user == nil {
		return nil
	}

	if err := u.db.Model(&User{}).Save(user).Error; err != nil {
		return err
	}

	return nil
}

func (u *UserModel) GetMaxId() (int, error) {
	var user User
	if err := u.db.Model(&User{}).Select("id").Order("id desc").First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}

	return user.ID, nil
}

func (u *UserModel) SetUserRoomIdToZero() error {
	if err := u.db.Exec("update users set room_id = 0, sid = '' where id >0").Error; err != nil {
		return err
	}

	return nil
}

func (u *UserModel) GetUsersByIds(ids []int) (map[int]User, error) {
	var users []User
	if err := u.db.Model(&User{}).Where("id in (?)", ids).Find(&users).Error; err != nil {
		return nil, err
	}

	var mapUser = make(map[int]User)
	for _, v := range users {
		mapUser[v.ID] = v
	}

	return mapUser, nil
}

func (u *UserModel) UpdateRoomID(user *User) error {
	if user == nil {
		return nil
	}

	if err := u.db.Model(user).UpdateColumns(User{RoomId: user.RoomId, State: user.State}).Error; err != nil {
		return err
	}

	return nil
}
