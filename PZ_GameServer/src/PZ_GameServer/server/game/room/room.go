package room

import (
	"PZ_GameServer/common/random_name"
	"PZ_GameServer/common/util"
	"PZ_GameServer/model"
	"PZ_GameServer/protocol/pb"
	"PZ_GameServer/sdk"
	"PZ_GameServer/server/game/room/ningbo"
	"PZ_GameServer/server/game/room/pinshi"
	"PZ_GameServer/server/game/room/srddz"
	"PZ_GameServer/server/game/room/xiangshan"
	"PZ_GameServer/server/game/room/xizhou"
	"PZ_GameServer/server/game/roombase"
	"PZ_GameServer/server/user"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

// 房间结构
type RoomHandle struct {
	RoomId     int         //
	Room       interface{} // 房间实例
	CreateTime int64       // 创建时间
	GameType   int         // 房间类型
}

type MatchItemList struct {
	List []*MatchItem
}

type MatchItem struct {
	Match *mjgame.Match_Room
	User  *user.User
}

var (
	RoomList    = make(map[int]*RoomHandle) // 房间
	MatchList   = make(map[string]*MatchItemList)
	RoomDefInfo = make(map[int32]*roombase.RoomRule) // 房间默认规则
	mutex       sync.RWMutex
	Config      = make(map[string]map[string]string)
	matchMutex  sync.Mutex
)

const (
	XiZhou    = int32(mjgame.MsgID_GTYPE_ZheJiang_XiZhou)
	XiangShan = int32(mjgame.MsgID_GTYPE_ZheJiang_XiangShan)
	NingBo    = int32(mjgame.MsgID_GTYPE_SanDizhu)
	ZhenHai   = int32(mjgame.MsgID_GTYPE_ZheJiang_ZhenHai)
	BeiLun    = int32(mjgame.MsgID_GTYPE_ZheJiang_Beilun)
	Srddz     = int32(mjgame.MsgID_GTYPE_SirenDizhu)
	Pinshi    = int32(mjgame.MsgID_GTYPE_Pinshi)
)

func LoadConfig(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	//UTF-8 text string with a Byte Order Mark (BOM). The BOM identifies that the text is UTF-8 encoded,
	// but it should be removed before decoding
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	err = json.Unmarshal(data, &Config)
	if err != nil {
		return err
	}
	return nil
}

// 初始化注册游戏类型
func InitGames() {
	RoomDefInfo[XiZhou] = &xizhou.XiZhou_RoomRule
	RoomDefInfo[XiangShan] = &xiangshan.XiangShan_RoomRule
	RoomDefInfo[NingBo] = &ningbo.NingBo_RoomRule
	RoomDefInfo[Srddz] = &srddz.Srddz_RoomRule
	RoomDefInfo[Pinshi] = &pinshi.Pinshi_RoomRule
	XueLiu_Init() // 血流
	go WatchRoom()
}

// 创建房间
func CreateRoom(roomId int, croom *mjgame.Create_Room, user *user.User) *RoomHandle {
	_, ok := RoomDefInfo[croom.Type]
	if !ok {
		return nil // 没有这种房间类型
	}

	// 创建房间
	room := GetNewRoom(roomId, croom.Type, user, croom.Rule)

	mutex.Lock()
	defer mutex.Unlock()
	RoomList[roomId] = &RoomHandle{
		RoomId:     roomId,
		Room:       room,
		CreateTime: time.Now().Unix(),
		GameType:   int(croom.Type),
	}

	return RoomList[roomId]
}

//匹配玩家
func MatchRoom(param *mjgame.Match_Room, muser *user.User) (*RoomHandle, []*user.User) {
	matchMutex.Lock()
	defer matchMutex.Unlock()

	var t = param.Type
	mt := ""
	ruleArr := append(param.Rule, t)
	for _, v := range ruleArr {
		mt += strconv.Itoa(int(v))
	}
	fmt.Println("mt:", mt)

	if MatchList[mt] == nil {
		MatchList[mt] = &MatchItemList{
			List: make([]*MatchItem, 0),
		}
	}

	MatchList[mt].List = append(MatchList[mt].List, &MatchItem{
		Match: param,
		User:  muser,
	})

	var matchLen int
	switch mjgame.MsgID(t) {
	case mjgame.MsgID_GTYPE_ZheJiang_XiZhou: // 西周
		matchLen = 4
	case mjgame.MsgID_GTYPE_SiChuan_XueLiu: // 血流
		matchLen = 4
	case mjgame.MsgID_GTYPE_ZheJiang_XiangShan: // 象山
		matchLen = 4
	case mjgame.MsgID_GTYPE_SanDizhu:
		matchLen = 3
	case mjgame.MsgID_GTYPE_SirenDizhu:
		matchLen = 4
	case mjgame.MsgID_GTYPE_Pinshi:
		matchLen = 2 //拼十两个人就可以玩哈
	}

	fmt.Println("len(MatchList[t].List):", len(MatchList[mt].List))
	fmt.Println("matchLen:", matchLen)

	if len(MatchList[mt].List) >= matchLen { //可以匹配
		tempList := append(MatchList[mt].List[:matchLen])
		MatchList[mt].List = append(MatchList[mt].List[matchLen:])

		creatRoom := &mjgame.Create_Room{
			SID:  tempList[0].Match.SID,
			Type: tempList[0].Match.Type,
			City: tempList[0].Match.City,
			PWD:  tempList[0].Match.PWD,
			Rule: tempList[0].Match.Rule,
		}

		//创建房间
		roomHandle, roomId := CreateLockRoom(creatRoom, tempList[0].User, -1)
		if roomId > 0 {
			users := make([]*user.User, 0)
			for _, v := range tempList {
				users = append(users, v.User)
			}
			//直接这几个玩家请求加入这个房间即可
			return roomHandle, users
		}
	} else if mjgame.MsgID(t) == mjgame.MsgID_GTYPE_SanDizhu { //三人斗地主目前直接给机器人

		tempList := append(MatchList[mt].List[:1])
		//给这个家伙弄两个ai

		for i := 0; i < 2; i++ {
			user := user.GetUser(nil)
			unionId := util.GetSID()
			openid := util.GetSID()
			usAi := CreateAiUser(23, 32, "255.255.0.255", nil, unionId, openid, i)
			if usAi.Sid == "" {
				usAi.Sid = usAi.OpenID
			}
			user.User = usAi
			tempList = append(tempList, &MatchItem{
				Match: tempList[0].Match,
				User:  user,
			})

		}

		MatchList[mt].List = append(MatchList[mt].List[1:])

		creatRoom := &mjgame.Create_Room{
			SID:  tempList[0].Match.SID,
			Type: tempList[0].Match.Type,
			City: tempList[0].Match.City,
			PWD:  tempList[0].Match.PWD,
			Rule: tempList[0].Match.Rule,
		}

		//创建房间
		roomHandle, roomId := CreateLockRoom(creatRoom, tempList[0].User, -1)
		if roomId > 0 {
			users := make([]*user.User, 0)
			for _, v := range tempList {
				users = append(users, v.User)
			}
			//直接这几个玩家请求加入这个房间即可
			return roomHandle, users
		}
	}
	return nil, nil
}

//创建新的用户																																																																																																																				```
func CreateAiUser(GPS_LNG, GPS_LAT float32, ip string, userInfo *sdk.UserInfo, unionID string, openid string, add int) *model.User {
	var user model.User

	user.Sid = util.GetSID()
	user.LastIp = ip
	user.Coin = 50000
	user.IsRobot = 1
	user.Diamond = 1000
	user.GPS_LAT = GPS_LNG
	user.GPS_LNG = GPS_LAT
	if userInfo != nil {
		user.NickName = userInfo.Nickname
		user.Sex = userInfo.Sex
		user.Province = userInfo.Province
		user.City = userInfo.City
		user.Country = userInfo.Country
		user.Icon = userInfo.Headimgurl
	} else {
		user.NickName = random_name.GetRandomName()
		//user.Icon = "icon_" + strconv.Itoa(util.RandInt(0, 5))
		user.Icon = ""
		user.Sex = 1
	}
	user.OpenID = openid
	user.UnionID = unionID
	fmt.Println("chuangjianxinyonghu meiwenti")

	//这个就不需要了，机器人嘛
	rand.Seed(time.Now().Unix())
	user.ID = rand.Intn(100000) + add

	//	err := model.GetUserModel().Create(&user)
	//	if err != nil {
	//		fmt.Println("charushibai::", err.Error())
	//	}

	return &user
}

//查找当前玩家的匹配信息
func GetMatchInfo(muser *user.User) *mjgame.Match_Room {
	matchMutex.Lock()
	defer matchMutex.Unlock()

	for _, v := range MatchList {
		for _, v1 := range v.List {
			if v1.User.ID == muser.ID {
				v1.User = muser
				return v1.Match
			}
		}
	}
	return nil
}

//将玩家从匹配列表中删除
func DeleMatchFromList(muser *user.User) bool {
	matchMutex.Lock()
	defer matchMutex.Unlock()

	for _, v := range MatchList {
		for idx, v1 := range v.List {
			if v1.User.ID == muser.ID {
				v.List = append(v.List[:idx], v.List[idx+1:]...)
				return true
			}
		}
	}
	return false
}

// 得到房间
func GetNewRoom(rid int, t int32, user *user.User, rules []int32) interface{} {
	var v interface{} = int64(0)

	key := strconv.Itoa(int(t))

	switch mjgame.MsgID(t) {
	case mjgame.MsgID_GTYPE_ZheJiang_XiZhou: // 西周
		room := xizhou.RoomXiZhou{}
		xizhou.XiZhou_RoomRule.Rules = rules
		room.Rules = xizhou.XiZhou_RoomRule
		room.RoomBase.InitRoomRule(rules, Config[key], 4)
		room.Create(rid, t, user, &room.Rules)
		room.CostType = rules[4]
		return &room
	case mjgame.MsgID_GTYPE_SiChuan_XueLiu: // 血流
		room := RoomXueLiu{}
		room.Rules = XueLiu_RoomRule
		room.Create(rid, t, user, &room.Rules)
		room.CostType = rules[4]
		return &room
	case mjgame.MsgID_GTYPE_ZheJiang_XiangShan: // 象山
		room := xiangshan.RoomXiangshan{}
		xiangshan.XiangShan_RoomRule.Rules = rules
		room.Rules = xiangshan.XiangShan_RoomRule
		room.RoomBase.InitRoomRule(rules, Config[key], 4)
		room.Create(rid, t, user, &room.Rules)
		room.CostType = rules[4]
		return &room
	case mjgame.MsgID_GTYPE_SanDizhu:
		room := ningbo.RoomNingBo{}
		ningbo.NingBo_RoomRule.Rules = rules
		room.Rules = ningbo.NingBo_RoomRule
		room.RoomBase.InitRoomRule(rules, Config[key], 3)
		room.Create(rid, t, user, &room.Rules)
		room.SubType = rules[3]
		room.CostType = rules[4]
		return &room
	case mjgame.MsgID_GTYPE_SirenDizhu:
		room := srddz.RoomSrddz{}
		srddz.Srddz_RoomRule.Rules = rules
		room.Rules = srddz.Srddz_RoomRule
		room.RoomBase.InitRoomRule(rules, Config[key], 4)
		room.Create(rid, t, user, &room.Rules)
		room.SubType = rules[3]
		room.CostType = rules[4]
		return &room
	case mjgame.MsgID_GTYPE_Pinshi:
		room := pinshi.RoomPinshi{}
		pinshi.Pinshi_RoomRule.Rules = rules
		room.Rules = pinshi.Pinshi_RoomRule
		room.RoomBase.InitRoomRule(rules, Config[key], 10)
		room.Create(rid, t, user, &room.Rules)
		room.SubType = rules[3]
		room.CostType = rules[4]
		return &room
	}
	return v
}

// 查询规则存在
func CheckRule(rule int, pRule *[]int) bool {
	for _, v := range *pRule {
		if v == rule {
			return true
		}
	}
	return false
}

//
func WatchRoom() {
	for {
		select {
		case roomId := <-roombase.ChanRoom:
			go func() {
				fmt.Println("shanchufangjian...0")
				time.Sleep(1 * time.Second)
				mutex.Lock()
				fmt.Println("shanchufangjian...1")
				delete(RoomList, roomId)
				fmt.Println("shanchufangjian...2")
				roombase.Redis_RemovePlayingUser(roomId)
				roombase.Redis_RemoveRoom(roomId)
				fmt.Println("Room delete ", roomId, RoomList[roomId])
				mutex.Unlock()
			}()
			//			mutex.Lock()
			//			delete(RoomList, roomId)
			//			roombase.Redis_RemovePlayingUser(roomId)
			//			roombase.Redis_RemoveRoom(roomId)
			//			fmt.Println("Room delete ", roomId, RoomList[roomId])
			//			mutex.Unlock()

		}
	}
}

//@andy新房间ID
func NewRoomId() int {
	mutex.Lock()
	defer mutex.Unlock()

	roomId := 0
	total := len(RoomList)
	var tempTime int64
	if total > 0 {
		for _, v := range RoomList {
			if tempTime < v.CreateTime {
				tempTime = v.CreateTime
				roomId = v.RoomId
			}
		}
	}

	return roomId
}

//@andy获取房间
func GetLockRoomHandle(roomId int) (*RoomHandle, bool) {
	mutex.Lock()
	defer mutex.Unlock()

	room, ok := RoomList[roomId]
	return room, ok
}

//@andy创建房间
func CreateLockRoom(croom *mjgame.Create_Room, user *user.User, rid int) (*RoomHandle, int) {
	mutex.Lock()
	defer mutex.Unlock()

	var roomId = rid
	if rid <= 0 {
		for {
			roomId = GetNewRoomID()
			_, haved := RoomList[roomId]
			if !haved {
				break
			}
		}
	}
	_, ok := RoomDefInfo[croom.Type]
	if !ok {
		fmt.Println("ddddddddddddddddddddddcaonima")
		return nil, -1 // 没有这种房间类型
	}

	// 创建房间
	room := GetNewRoom(roomId, croom.Type, user, croom.Rule)

	RoomList[roomId] = &RoomHandle{
		RoomId:     roomId,
		Room:       room,
		CreateTime: time.Now().Unix(),
		GameType:   int(croom.Type),
	}

	return RoomList[roomId], roomId
}

//获取新房间号
func GetNewRoomID() int {
	temp := strconv.Itoa(util.RandInt(1, 999999))
	mlen := 6 - len(temp)

	for i := 0; i < mlen; i++ {
		temp += "0"
	}
	roomId, _ := strconv.Atoi(temp)
	return roomId
}
