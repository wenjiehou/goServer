/*
	房间基类
*/
package roombase

import (
	"fmt"
	//"math/rand"
	"reflect"
	"sort"
	"strconv"
	"time"

	"PZ_GameServer/common/util"
	al "PZ_GameServer/common/util/arrayList"
	"PZ_GameServer/protocol/def"
	"PZ_GameServer/protocol/pb"
	st "PZ_GameServer/server/game/statement"
	"PZ_GameServer/server/user"

	px "PZ_GameServer/server/game/room/ningbo/paixingLogic"

	proto "github.com/golang/protobuf/proto"
	"github.com/rs/xid"
)

//状态Enum
const (
	None                 = 0 // 无状态, 游戏停止
	Dealt                = 1 // 骰子/发牌
	WaitStartChange3Card = 2 // 等待开始换三张
	WaitChange3Card      = 3 // 等待换3张
	WaitFixMiss          = 4 // 等待定缺
	WaitPut              = 5 // 等待打牌
	WaitTool             = 6 // 等待操作
	Total                = 7 // 结算
	Finished             = 8 // 当前房间销毁
)

//判断规则是否存在 NoLock
func (r *RoomBase) CheckRule(rules []int32, rule_type int32) bool {
	for _, rule := range rules {
		if rule == rule_type {
			return true
		}
	}
	return false
}

//房间创建规则初始化赋值 NoLock
func (r *RoomBase) InitRoomRule(rules []int32, config map[string]string, playerNum int) {
	for _, rule := range rules {
		switch rule {
		case 6:
			r.Rules.MaxRound = 10
			value, _ := strconv.Atoi(config["6"])
			r.Rules.Play_NeedDiamond = value
		case 7:
			r.Rules.MaxRound = 20
			value, _ := strconv.Atoi(config["7"])
			r.Rules.Play_NeedDiamond = value
		case 8:
			r.Rules.MaxRound = 34
			value, _ := strconv.Atoi(config["8"])
			r.Rules.Play_NeedDiamond = value
		case 4:
			r.Rules.MaxRound = 6
			value, _ := strconv.Atoi(config["4"])
			r.Rules.Play_NeedDiamond = value
		case 5:
			r.Rules.MaxRound = 12
			value, _ := strconv.Atoi(config["5"])
			r.Rules.Play_NeedDiamond = value
		case 18:
			r.Rules.MaxRound = 20
			value, _ := strconv.Atoi(config["18"])
			r.Rules.Play_NeedDiamond = value
		case 19:
			r.Rules.MaxTime = 15 //
			value, _ := strconv.Atoi(config["19"])
			r.Rules.Play_NeedDiamond = value
		case 20:
			r.Rules.MaxTime = 20
			value, _ := strconv.Atoi(config["20"])
			r.Rules.Play_NeedDiamond = value
		case 21:
			r.Rules.MaxTime = 30
			value, _ := strconv.Atoi(config["21"])
			r.Rules.Play_NeedDiamond = value
		case 22:
			r.Rules.MaxTime = 60
			value, _ := strconv.Atoi(config["22"])
			r.Rules.Play_NeedDiamond = value
		case 23:
			r.Rules.PayType = 4
		case 24:
			r.Rules.PayType = 1
		case 25:
			r.Rules.MaxTai = 50
		case 26:
			r.Rules.MaxTai = 100
		case 27:
			r.Rules.MaxTai = 0
		}
	}

	if r.Rules.PayType == 4 {
		r.Rules.Play_NeedDiamond = r.Rules.Play_NeedDiamond / playerNum
	} else { //竟然一样多，我日
		r.Rules.Play_NeedDiamond = r.Rules.Play_NeedDiamond / playerNum
	}
}

// 创建房间 NoLock
func (r *RoomBase) Create(roomId int, t int32, userId int, rule *RoomRule) {
	r.RoomId = roomId
	r.Type = t
	r.CreateTime = time.Now().Unix()
	r.CreateUserId = userId
	r.VoteTimeOut = def.VoteTimeOut // 投票超时时间
	r.Votes = make([]int, rule.SeatLen)
	r.VoteStarter = -1
	//r.State <- int(mjgame.StateID_GameState_Normal)
	r.UniqueCode = xid.New().String()
	r.StopTicker = false

	r.Seats = make([]*SeatBase, rule.SeatLen) // 初始化座位
	for i := 0; i < rule.SeatLen; i++ {
		r.Seats[i] = &SeatBase{}
	}
	r.Rules = *rule
}

// 初始化 NoLock
func (r *RoomBase) Init() {
	r.WaitOptTool = &RoomWaitOpts{} //
	r.WaitOptTool.ClearAll()        // 初始化

	for _, v := range r.Seats {
		v.Disband = -1
		v.IsZhuang = false
		v.Ting = 0
		v.Message = nil
		r.CurCard = nil
		v.HuCardIDs = []int{}
		v.PengCardIDs = []int{}
		v.ChowCardIDs = []int{}
		v.LastCardID = -1
		v.IsTransfer = false
		v.IsPutCard = false //是否出过牌 用于地胡

		v.Cards = &UserCard{}
		v.Cards.List = al.New()
		v.Cards.Out = al.New()
		v.Cards.Peng = al.New()
		v.Cards.Kong = al.New()
		v.Cards.Chow = al.New()
		v.Cards.Hu = al.New()
		v.Cards.Hua = al.New()
	}
}

// 通知操作 NoLock
func (r *RoomBase) NotifyTool() {
	for i := 0; i < r.WaitOptTool.Count(); i++ {
		needWait := r.WaitOptTool.NeedWaitTool[i]

		if needWait == nil || r.Seats[needWait.Index].User == nil || needWait.Index < 0 {
			continue //用户断线,或者不在座位上,不需要发送操作消息
		}

		ack := &mjgame.ACK_WaitTool{
			Seat:    int32(needWait.Index),
			Type:    []int32{int32(needWait.CanTools[0]), int32(needWait.CanTools[1]), int32(needWait.CanTools[2]), int32(needWait.CanTools[3])},
			TimeOut: int32(r.WaitToolTimeOut),
		}

		r.Seats[needWait.Index].User.SendMessage(mjgame.MsgID_MSG_ACK_WaitTool, ack)
		r.SaveBattleRecord(needWait.Index, mjgame.MsgID_MSG_ACK_WaitTool, ack)

		r.Seats[needWait.Index].Message = &Message{ // 断线重连的消息
			Id:      mjgame.MsgID_MSG_ACK_WaitTool,
			Content: ack,
		}

		r.MToolChecker.SetTools(
			needWait.Index,
			[]int{
				needWait.CanTools[0],
				needWait.CanTools[1],
				needWait.CanTools[2],
				needWait.CanTools[3],
				needWait.CanTools[4],
				needWait.CanTools[5], -1, -1,
			},
		)
	}

}

func (r *RoomBase) Try(fun func(), handler func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			handler(err)
		}
	}()
	fun()
}

// 添加消息记录
func (r *RoomBase) AddMsgList(args ...interface{}) {
	//m := &mjgame.SitDown{}
	//er := proto.Unmarshal(*(args[0].(*[]byte)), m)
	m := (args[1]).(user.RunParam)
	r.MsgList.AddMessage(&m)
}

// 初始化全部的牌 麻将用
func (r *RoomBase) InitRandAllCard() {
	for i := 0; i < r.Rules.SeatLen; i++ {
		r.Seats[i].Cards = &UserCard{}
		r.Seats[i].Cards.List = al.New()
		r.Seats[i].Cards.Out = al.New()
		r.Seats[i].Cards.Peng = al.New()
		r.Seats[i].Cards.Kong = al.New()
		r.Seats[i].Cards.Chow = al.New()
		r.Seats[i].Cards.Hua = al.New()
		r.Seats[i].Cards.Hu = al.New()
	}

	r.AllCards = make([]Card, r.Rules.AllCardLen)    //所有牌
	cacheAllCard := make([]Card, r.Rules.AllCardLen) //

	Count := 0
	for i := 0; i < 9 && Count < r.Rules.AllCardLen; i++ {
		for l := 0; l < 4; l++ { //万 4*9=36
			c := Card{Type: W, Num: i, Status: -1, TIndex: -1}
			c.ID = st.GetMjIndex(c.Type, c.Num)
			c.MSG = st.GetMjNameForIndex(c.ID)
			cacheAllCard[Count] = c
			Count++
		}
	}

	for i := 0; i < 9 && Count < r.Rules.AllCardLen; i++ {
		for l := 0; l < 4; l++ { //饼 4*9=36
			c := Card{Type: B, Num: i, Status: -1, TIndex: -1}
			c.ID = st.GetMjIndex(c.Type, c.Num)
			c.MSG = st.GetMjNameForIndex(c.ID)
			cacheAllCard[Count] = c
			Count++
		}
	}

	for i := 0; i < 9 && Count < r.Rules.AllCardLen; i++ {
		for l := 0; l < 4; l++ { //条 4*9=36
			c := Card{Type: T, Num: i, Status: -1, TIndex: -1}
			c.ID = st.GetMjIndex(c.Type, c.Num)
			c.MSG = st.GetMjNameForIndex(c.ID)
			cacheAllCard[Count] = c
			Count++
		}
	}

	for i := 0; i < 4 && Count < r.Rules.AllCardLen; i++ {
		for l := 0; l < 4; l++ { //风 4*4=16 东 南 西 北
			c := Card{Type: F, Num: i, Status: -1, TIndex: -1}
			c.ID = st.GetMjIndex(c.Type, c.Num)
			c.MSG = st.GetMjNameForIndex(c.ID)
			cacheAllCard[Count] = c
			Count++
		}
	}

	//for i := 0; i < r.Rules.Card_J; i++ {

	for i := 0; i < 3 && Count < r.Rules.AllCardLen; i++ {
		for l := 0; l < 4; l++ { //箭 4*3=12 中 发 白
			c := Card{Type: J, Num: i, Status: -1, TIndex: -1}
			c.ID = st.GetMjIndex(c.Type, c.Num)
			c.MSG = st.GetMjNameForIndex(c.ID)
			cacheAllCard[Count] = c
			Count++
		}
	}

	//}

	//for i := 0; i < r.Rules.Card_H; i++ { // 花牌
	for i := 0; i < 8 && Count < r.Rules.AllCardLen; i++ {
		c := Card{Type: H, Num: i, Status: -1, TIndex: -1}
		c.ID = st.GetMjIndex(c.Type, c.Num)
		c.MSG = st.GetMjNameForIndex(c.ID)
		cacheAllCard[Count] = c
		Count++
	}
	//}

	// 混淆
	for rr := 0; rr < 300; rr++ {
		rand1 := util.RandInt64(0, int64(r.Rules.AllCardLen))
		rand2 := util.RandInt64(0, int64(r.Rules.AllCardLen))
		c1 := Card{ID: cacheAllCard[rand1].ID, Type: cacheAllCard[rand1].Type, Num: cacheAllCard[rand1].Num, MSG: cacheAllCard[rand1].MSG}
		c2 := Card{ID: cacheAllCard[rand2].ID, Type: cacheAllCard[rand2].Type, Num: cacheAllCard[rand2].Num, MSG: cacheAllCard[rand2].MSG}
		cacheAllCard[rand2] = c1
		cacheAllCard[rand1] = c2
	}

	for c := 0; c < r.Rules.AllCardLen; c++ {
		r.AllCards[c] = cacheAllCard[c]
		//fmt.Println(r.AllCards[c].ID, r.AllCards[c].Type, r.AllCards[c].Num, r.AllCards[c].MSG)
	}

	r.AllCardLength = len(r.AllCards)
	r.EndBlank = 0

	//	// 混淆
	//	r.AllCards = make([]Card, 0)

	//	//	if Debug {
	//	//		r.AllCards = cacheAllCard
	//	//	} else {
	//	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	//	randSlices := seed.Perm(len(cacheAllCard))
	//	for _, v := range randSlices {
	//		r.AllCards = append(r.AllCards, cacheAllCard[v])
	//	}
	//	//	}

}

// 初始化全部的牌 三人斗地主用
func (r *RoomBase) InitRandPockAllCard() {
	for i := 0; i < r.Rules.SeatLen; i++ {
		r.Seats[i].Cards = &UserCard{}
		r.Seats[i].Cards.List = al.New()
		r.Seats[i].Cards.Out = al.New()
		r.Seats[i].Cards.Peng = al.New()
		r.Seats[i].Cards.Kong = al.New()
		r.Seats[i].Cards.Chow = al.New()
		r.Seats[i].Cards.Hua = al.New()
		r.Seats[i].Cards.Hu = al.New()
		r.Seats[i].Cards.OutStep = make([][]int, 0)
	}

	r.AllCards = make([]Card, r.Rules.AllCardLen) //所有牌

	cacheAllCard := make([]Card, r.Rules.AllCardLen) //

	Count := 0
	for i := 3; i < 17 && Count < r.Rules.AllCardLen; i++ { //梅花
		if i == 15 {
			continue
		} else {
			c := Card{Type: px.POCK_MEI, Num: i, Status: -1, TIndex: -1}
			c.ID = px.GetPockIndex(c.Type, c.Num)
			c.MSG = ""
			cacheAllCard[Count] = c
			Count++
		}
	}

	for i := 3; i < 17 && Count < r.Rules.AllCardLen; i++ { //黑桃
		if i == 15 {
			continue
		} else {
			c := Card{Type: px.POCK_HEI, Num: i, Status: -1, TIndex: -1}
			c.ID = px.GetPockIndex(c.Type, c.Num)
			c.MSG = ""
			cacheAllCard[Count] = c
			Count++
		}
	}

	for i := 3; i < 17 && Count < r.Rules.AllCardLen; i++ { //方块
		if i == 15 {
			continue
		} else {
			c := Card{Type: px.POCK_FANG, Num: i, Status: -1, TIndex: -1}
			c.ID = px.GetPockIndex(c.Type, c.Num)
			c.MSG = ""
			cacheAllCard[Count] = c
			Count++
		}
	}

	for i := 3; i < 17 && Count < r.Rules.AllCardLen; i++ { //红桃
		if i == 15 {
			continue
		} else {
			c := Card{Type: px.POCK_HONG, Num: i, Status: -1, TIndex: -1}
			c.ID = px.GetPockIndex(c.Type, c.Num)
			c.MSG = ""
			cacheAllCard[Count] = c
			Count++
		}
	}

	if r.Rules.GameType != int(mjgame.MsgID_GTYPE_Pinshi) {
		for i := 51; i < 53 && Count < r.Rules.AllCardLen; i++ { //大小王
			c := Card{Type: px.POCK_WANG, Num: i, Status: -1, TIndex: -1}
			c.ID = px.GetPockIndex(c.Type, c.Num)
			c.MSG = ""
			cacheAllCard[Count] = c
			Count++
		}
	}

	if r.Rules.GameType == int(mjgame.MsgID_GTYPE_SirenDizhu) {
		for i := 3; i < 17 && Count < r.Rules.AllCardLen; i++ { //梅花
			if i == 15 {
				continue
			} else {
				c := Card{Type: px.POCK_MEI, Num: i, Status: -1, TIndex: -1}
				c.ID = px.GetPockIndex(c.Type, c.Num)
				c.MSG = ""
				cacheAllCard[Count] = c
				Count++
			}
		}

		for i := 3; i < 17 && Count < r.Rules.AllCardLen; i++ { //黑桃
			if i == 15 {
				continue
			} else {
				c := Card{Type: px.POCK_HEI, Num: i, Status: -1, TIndex: -1}
				c.ID = px.GetPockIndex(c.Type, c.Num)
				c.MSG = ""
				cacheAllCard[Count] = c
				Count++
			}
		}
		for i := 3; i < 17 && Count < r.Rules.AllCardLen; i++ { //方块
			if i == 15 {
				continue
			} else {
				c := Card{Type: px.POCK_FANG, Num: i, Status: -1, TIndex: -1}
				c.ID = px.GetPockIndex(c.Type, c.Num)
				c.MSG = ""
				cacheAllCard[Count] = c
				Count++
			}
		}

		for i := 3; i < 17 && Count < r.Rules.AllCardLen; i++ { //红桃
			if i == 15 {
				continue
			} else {
				c := Card{Type: px.POCK_HONG, Num: i, Status: -1, TIndex: -1}
				c.ID = px.GetPockIndex(c.Type, c.Num)
				c.MSG = ""
				cacheAllCard[Count] = c
				Count++
			}
		}
		for i := 51; i < 53 && Count < r.Rules.AllCardLen; i++ { //大小王
			c := Card{Type: px.POCK_WANG, Num: i, Status: -1, TIndex: -1}
			c.ID = px.GetPockIndex(c.Type, c.Num)
			c.MSG = ""
			cacheAllCard[Count] = c
			Count++
		}

	}

	// 混淆
	for rr := 0; rr < 300; rr++ {
		rand1 := util.RandInt64(0, int64(r.Rules.AllCardLen))
		rand2 := util.RandInt64(0, int64(r.Rules.AllCardLen))
		c1 := Card{ID: cacheAllCard[rand1].ID, Type: cacheAllCard[rand1].Type, Num: cacheAllCard[rand1].Num, MSG: cacheAllCard[rand1].MSG}
		c2 := Card{ID: cacheAllCard[rand2].ID, Type: cacheAllCard[rand2].Type, Num: cacheAllCard[rand2].Num, MSG: cacheAllCard[rand2].MSG}
		cacheAllCard[rand2] = c1
		cacheAllCard[rand1] = c2
	}

	for c := 0; c < r.Rules.AllCardLen; c++ {
		r.AllCards[c] = cacheAllCard[c]
	}

	r.AllCardLength = len(r.AllCards)
	r.EndBlank = 0
}

//进入房间
func (r *RoomBase) IntoUser(user *user.User) {

	user.RoomId = r.RoomId

	index := r.GetSeatIndexById(user.ID)
	if index < 0 {
		wIndex := r.GetWatchSeat(user.ID)
		if wIndex < 0 {
			r.WatchSeats = append(r.WatchSeats, user)
		} else {
			r.WatchSeats[wIndex] = user
		}
	} else {
		r.Seats[index].UID = strconv.Itoa(user.ID)
		r.Seats[index].User = user
	}
	ack := mjgame.ACKBC_Into_Room{
		Name:    user.NickName,
		Uid:     strconv.Itoa(user.ID),
		RoomId:  int32(user.RoomId),
		Ip:      user.GetIP(),
		Index:   -1,
		Icon:    user.Icon,
		Coin:    int32(user.Coin),
		Type:    int32(r.Rules.GameType),
		Diamond: int32(user.Diamond),
		Level:   0,
		Robot:   int32(user.IsRobot),
		GPS_LNG: user.GPS_LNG,
		GPS_LAT: user.GPS_LAT,
		Rule:    r.Rules.Rules,
	}

	r.BCMessage(mjgame.MsgID_MSG_ACK_RoomInfo, r.GetRoomInfo()) // 房间信息

	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Into_Room, &ack)

	r.BCMessage(mjgame.MsgID_MSG_ACK_Room_User, r.GetRoomUser()) // 用户信息

	if r.IsRun {
		r.SendGameInfo(user, false)
		r.SendSeatCard(user.ID)
	} else {
		if index > 0 {
			if r.Seats[index].State == int(mjgame.StateID_GameState_Total) {
				r.SendGameInfo(user, false)
				r.SendSeatCard(user.ID)
			}
		}
		// TODO 围观用户
	}
}

//得到空余的位置 或者以前的位置
func (r *RoomBase) GetBlankOrSeatIndex(uid int) int {
	seatLen := len(r.Seats)
	for i := 0; i < seatLen; i++ {
		if r.Seats == nil {
			continue
		}
		if r.Seats[i] != nil && r.Seats[i].UID == string(uid) {
			return i
		}
	}
	for i := 0; i < seatLen; i++ {
		if r.Seats[i] == nil {
			return i
		}
		if r.Seats[i].User == nil {
			return i
		}
	}
	return -1
}

//退出房间
func (r *RoomBase) ExitUser(user *user.User) {

}

//是否可以开始游戏(判断玩家金币, 时间)
func (r *RoomBase) CheckCanStart() {

}

//开始游戏
func (r *RoomBase) Start(user *user.User) {
}

//重新开始游戏
func (r *RoomBase) Restart() {

}

//游戏过程
func (r *RoomBase) Process() {
}

//房间全部信息(断线重连, 新玩家进入)
func (r *RoomBase) SendGameInfo(a *user.User, needRecord bool) {
	var iRun int //是否游戏运行中
	if r.IsRun {
		iRun = 1
	}

	curCardId := -1 //最后打出的牌
	if r.CurCard != nil {
		curCardId = r.CurCard.ID
	}

	iAllCardLen := len(r.AllCards)

	iCardLeft := iAllCardLen - r.CurMJIndex - r.EndBlank

	if r.Type == int32(mjgame.MsgID_GTYPE_ZheJiang_XiangShan) || r.Type == int32(mjgame.MsgID_GTYPE_ZheJiang_XiZhou) { //象山和西周需要删掉指定的16张
		iCardLeft = iCardLeft - def.XiangShanDrawCount
	}

	ack := &mjgame.ACKBC_Card_Init{ //  ------ 牌面信息
		Dict:         int32(r.Dict),            // 筛子
		CardCount:    int32(iAllCardLen),       // 总共有多少张牌
		CardFirst:    int32(r.CurMJIndex),      // 拿牌的位置
		EndBlank:     int32(r.EndBlank),        // 结尾空余的牌
		CardLeft:     int32(iCardLeft),         // 还剩下多少张牌
		CurCardId:    int32(curCardId),         //
		CurIndex:     int32(r.CurIndex),        //
		CurTime:      int32(r.WaitToolTimeOut), //
		ZhuangIndex:  int32(r.BankerIndex),     // UID
		StartGame:    int32(iRun),              //
		Defeat:       []int32{0, 0, 0, 0},      //
		LastPutIndex: int32(r.LastPutIndex),    //
		Show:         r.Show,                   //
	}

	if a == nil {
		r.BCMessage(mjgame.MsgID_MSG_ACKBC_Card_Init, ack)
	} else {
		a.SendMessage(mjgame.MsgID_MSG_ACKBC_Card_Init, ack)
	}

	//记录战绩
	if needRecord {
		r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Card_Init, ack)
	}

}

//转到下一个等待操作的玩家
func (r *RoomBase) TurnNextWaitTool() {
	//	//所有玩家操作完毕
	//	if r.NeedWaitTool.Count == 0 || r.CurToolIndex >= r.NeedWaitTool.Count {
	//		r.TurnNextPlayer(true, true)
	//		return
	//	}

	//	//等待操作
	//	waitParam := (*r.NeedWaitTool.Index(r.CurToolIndex)).(*NeedWait)

	//	if waitParam.User != nil {
	//		index := r.GetSeatIndexById(waitParam.User.ID) //当前需要操作的玩家index

	//		ack := mjgame.ACK_WaitTool{
	//			Seat:    int32(index),
	//			Type:    []int32{int32(waitParam.Win), int32(waitParam.Kong), int32(waitParam.Peng), int32(waitParam.Chow)},
	//			TimeOut: int32(r.WaitToolTimeOut),
	//		}

	//		waitParam.User.SendMessage(mjgame.MsgID_MSG_ACK_WaitTool, &ack)

	//		r.WaitTimeCount = r.WaitToolTimeOut
	//	} else {
	//		//断线的不操作
	//		r.WaitTimeCount = 0
	//	}

	//	//回调
	//	r.TimeOutCB = reflect.ValueOf(r.TurnNextWaitTool)
	//	r.TimeOutCBParam = nil

	//	r.CurToolIndex++
}

//转到下一个出牌的玩家
func (r *RoomBase) TurnNextPlayer(bGetCard bool, bForward bool) {
}

//开始等待操作
func (r *RoomBase) StartWaitTool(card *Card) {
}

//等待开始
func (r *RoomBase) WaitStart() {
}

//等待操作
func (r *RoomBase) WaitTool(timeout int) {
	r.WaitTimeCount--

	if r.CurToolIndex >= 0 && r.Seats[r.CurToolIndex].User != nil {
		r.AutoTool()
	}

	if r.WaitTimeCount <= 0 {
		r.WaitTimeCount = 0
		r.TimeOutCB.Call(r.TimeOutCBParam)
	}
}

//等待出牌
func (r *RoomBase) WaitPut(timeout int) {
	r.NotifyPutCard(r.CurIndex, WaitPut, r.WaitPutTimeOut) // 通知当前玩家出牌
	r.WaitTimeCount = timeout

	if r.Seats[r.CurIndex].User != nil {
		r.WaitTimeCount = 2 //机器自动打牌
	}

	//设置
	r.TimeOutCB = reflect.ValueOf(r.waitTimeOut)
	r.TimeOutCBParam = []reflect.Value{reflect.ValueOf(user.Param{User: r.Seats[r.CurIndex].User, Index: r.CurIndex})}
}

//等待超時
func (r *RoomBase) waitTimeOut(args Param) {

}

//判断是否可以胡牌
func (r *RoomBase) CheckWin(uIndex int, card *Card) int {
	cardsList := r.Seats[uIndex].Cards.List
	if cardsList == nil {
		return 0
	}

	length := cardsList.Count
	if card != nil {
		length = cardsList.Count + 1
	}
	if length == 1 || length%3 == 0 {
		return 0
	}

	mjs := make([]int, 42)

	for i := 0; i < cardsList.Count; i++ {
		if *cardsList.Index(i) != nil {
			c := (*cardsList.Index(i)).(*Card)
			mjs[c.ID]++ // 类型 w=0 b=1 t=2
		}

	}

	if card != nil {
		mjs[card.ID]++
	}

	str := ""
	for i := 0; i < len(mjs); i++ {
		str += strconv.Itoa(mjs[i]) + ","
	}

	return IsWin(mjs)
}

//判断是否可以杠牌
func (r *RoomBase) CheckKong(cid int, self bool, index int) int {
	var kongType, count, flag int

	turns := 3
	if self {
		turns = 1
	}

	if index != r.CurIndex { //碰别人
		flag = 1
	}
	for i := 0; i < turns; i++ {
		count++
		if !self {
			index++
		}
		index = index % 4

		if r.Seats[index].User == nil || r.CurIndex == i {
			continue //玩家断线则不操作,跳过
		}

		cards := r.Seats[index].Cards

		cardCount := r.GetUserCardCount(index, cid)
		if !self {
			if cardCount >= def.DoTypePengNumber { //可以明杠
				kongType = def.KongTypeMing
			}
		}

		if self {
			if flag > 0 {
				if cardCount >= def.DoTypePengNumber { //可以明杠
					kongType = def.KongTypeMing
				}
			} else {
				//检查碰后杠( 必须要自摸才能杠)
				t, n := st.GetMjTypeNum(cid)
				kCount := r.GetCardCount(cards.Peng, t, n)
				if kCount >= def.DoTypePengNumber { // 在碰的牌中找
					kongType = def.KongTypePeng
				}
				if cardCount >= def.DoTypeKongNumber { //可以暗杠
					kongType = def.KongTypeAn
				}
			}
		}

		if kongType > 0 {
			r.AddToolUser(index, 0, kongType, 0, 0, 0, 1)
			r.Seats[index].IsCanKong = true
			return kongType
		}

		if count != turns {
			kongType = 0
		}
	}

	//自己手牌杠判断
	if self {
		cards := r.Seats[index].Cards
		//检查碰后杠(在手牌中检查)
		listCards := r.GetListArray(cards.List)
		pengCards := r.GetListArray(cards.Peng)

	L:
		for i, v := range pengCards {
			if i%3 == 0 {
				for _, card := range listCards {
					if v.Cid == card.Cid {
						kongType = def.KongTypePeng
						break L
					}
				}
			}
		}
		if kongType == 0 {
			//暗杠
			for _, card := range listCards {
				cardCount := r.GetUserCardCount(index, int(card.Cid))
				if cardCount == def.DoTypeKongNumber { //可以暗杠
					kongType = def.KongTypeAn
					break
				}
			}
		}

		if kongType > 0 {
			r.AddToolUser(index, 0, kongType, 0, 0, 0, 1)
			r.Seats[index].IsCanKong = true
		}
	}

	return kongType
}

//判断是否可以碰牌
func (r *RoomBase) CheckPeng(pCard *Card) {
	index := r.CurIndex

	for i := 0; i < 3; i++ {
		index++ //从下家开始判断
		index = index % 4
		if r.Seats[index].User == nil {
			continue
		}
		cardCount := r.GetUserCardCount(index, pCard.ID)
		if cardCount >= 2 {
			r.AddToolUser(index, 0, 0, 1, 0, 0, 1)
			r.Seats[index].IsCanPeng = true

		}
	}
}

//判断是否可以吃牌
func (r *RoomBase) CheckChow(pCard *Card) {
	if pCard.Type == F || pCard.Type == H || pCard.Type == J {
		return
	}

	index := (r.CurIndex + 1) % 4

	if r.Seats[index].User == nil { //玩家断线则不操作,跳过
		return
	}

	list := r.Seats[index].Cards.List
	var n1, n2, n3, n4 bool

	for i := 0; i < list.Count; i++ {
		if *list.Index(i) != nil {
			card := (*list.Index(i)).(*Card)
			if card.Type == pCard.Type {
				switch card.Num {
				case pCard.Num - 2:
					n1 = true
				case pCard.Num - 1:
					n2 = true
				case pCard.Num + 1:
					n3 = true
				case pCard.Num + 2:
					n4 = true
				}
			}
		}

	}

	if n1 && n2 || n2 && n3 || n3 && n4 {
		r.AddToolUser(index, 0, 0, 0, 1, 0, 1)
		r.Seats[index].IsCanChow = true
	}
}

// 检查并更新操作检查器
func (r *RoomBase) UpdateToolChecker() {

	// type NeedWait struct {
	//	User *user.User
	//	Win  int
	//	Kong int
	//	Peng int
	//	Chow int
	// }
	r.MToolChecker.SetAllUserTool(-1) // 禁止全部操作

	for i := 0; i < r.WaitOptTool.Count(); i++ {
		//fmt.Println(i, "waitUser.User", waitUser.Index, waitUser.CanTools)
		//if waitUser.User != nil {
		// 胡0 杠1 碰2 吃3 出4 过5 摸6 等7
		cantools := r.WaitOptTool.NeedWaitTool[i].CanTools
		r.MToolChecker.SetTools(r.WaitOptTool.NeedWaitTool[i].Index, []int{cantools[0], cantools[1], cantools[2], cantools[3], -1, 0, -1, -1})
	}

	r.MToolChecker.ShowAllTools()

}

//判断是否可以听牌
func (r *RoomBase) CheckTing(uIndex int, pcard *Card) {
}

//判断是否可以过
func (r *RoomBase) CheckPass() {
}

//是否可以摸牌
func (r *RoomBase) CheckCanGet(index int, cid int) bool {

	return true
}

// 是否可以出牌
func (r *RoomBase) CheckCanPut(index int, cid int) bool {

	if cid < 0 || cid > 33 { // 花不能出
		return false // 牌不对
	}

	ccount := r.Seats[index].Cards.List.Count //手牌数量不对
	if ccount < 1 || ccount == 3 || ccount == 4 || ccount == 6 || ccount == 7 || ccount == 9 || ccount == 10 || ccount == 12 || ccount == 13 {
		return false
	}

	c := r.GetUserCardCount(index, cid)
	if c > 0 {
		return true
	} else {
		return false
	}
}

//是否可以杠牌
//参数 	: int用户索引,    int牌ID(-1=判断其他牌的暗杠),     bool是否是只检查自己(暗杠, 碰杠)false=检查其他3家(直杠)
//返回值 	: int可以杠牌的用户(-1没有)   int杠牌类型(明杠 暗杠 碰杠)    cid牌的
func (r *RoomBase) CheckCanKong(index int, cid int, self bool) (int, int) {
	if index < 0 || index > 3 || cid > 33 {
		return -1, -1 // 牌不对
	}
	strCardName := st.GetMjNameForIndex(cid)

	if self {

		//if cid >= 0 {
		// 只检查自己(暗杠, 碰杠)
		ccount := r.Seats[index].Cards.List.Count

		kongSelf := false // 杠自己
		countmod := ccount % 3
		if countmod == 2 { //手牌数量不对
			kongSelf = true
		} else if countmod == 1 {
			kongSelf = false
		} else if countmod == 0 {
			fmt.Println("手牌数量不对 ", ccount, ccount%3)
			return -1, -1
		}

		c := r.GetUserCardCount(index, cid)
		if !kongSelf {
			if c >= 3 && r.CurCard.ID == cid {
				r.AddToolUser(index, 0, 1, 0, 0, 1, 1)
				r.Seats[index].IsCanPeng = true
				r.RoomRecord += "检测(" + r.Seats[index].User.NickName + ") 碰杠" + strCardName + "\r\n"
				return index, def.KongTypeMing //明杠
			} else {
				return -1, -1
			}
		}

		if c >= 4 {
			r.AddToolUser(index, 0, 1, 0, 0, 1, 1)
			r.Seats[index].IsCanPeng = true
			r.RoomRecord += "检测(" + r.Seats[index].User.NickName + ") 暗杠" + strCardName + "\r\n"
			return index, def.KongTypeAn //暗杠
		}
		t, n := st.GetMjTypeNum(cid)
		c2 := r.GetCardCount(r.Seats[index].Cards.Peng, t, n)
		if c >= 1 && c2 >= 3 {
			r.AddToolUser(index, 0, 1, 0, 0, 1, 1)
			r.RoomRecord += "检测(" + r.Seats[index].User.NickName + ") 碰杠" + strCardName + "\r\n"
			r.Seats[index].IsCanPeng = true
			return index, def.KongTypePeng //碰杠
		}
		//} else {
		// 判断其他牌的暗杠
		mjs := make([]int, 42)
		for i := 0; i < r.Seats[index].Cards.List.Count; i++ {
			if *r.Seats[index].Cards.List.Index(i) != nil {
				card := (*r.Seats[index].Cards.List.Index(i)).(*Card)
				mjs[card.ID]++
				if mjs[card.ID] >= 4 {
					r.AddToolUser(index, 0, 1, 0, 0, 1, 1)
					r.Seats[index].IsCanPeng = true
					r.RoomRecord += "检测(" + r.Seats[index].User.NickName + ") 暗杠" + strCardName + "\r\n"
					return index, def.KongTypeAn
				}
			}

		}

		// 判断单张的暗杠
		pengmjs := make([]int, 42)
		for i := 0; i < r.Seats[index].Cards.Peng.Count; i++ {
			if *r.Seats[index].Cards.Peng.Index(i) != nil {
				card := (*r.Seats[index].Cards.Peng.Index(i)).(*Card)
				pengmjs[card.ID]++
				if mjs[card.ID] > 0 { //碰杠
					r.AddToolUser(index, 0, 1, 0, 0, 1, 1)
					r.Seats[index].IsCanPeng = true
					r.RoomRecord += "检测(" + r.Seats[index].User.NickName + ") 碰杠" + strCardName + "\r\n"
					return index, def.KongTypePeng
				}
			}

		}

		//}

	} else {
		for i := 0; i < 3; i++ {
			index++ //从下家开始判断
			index = index % 4
			cardCount := r.GetUserCardCount(index, cid)
			if cardCount >= 3 {
				r.AddToolUser(index, 0, 1, 0, 0, 0, 1)
				//r.RoomRecord += "检测(" + r.Seats[index].User.NickName + ") 杠" + strCardName + "\r\n"
				r.Seats[index].IsCanKong = true
			}
		}
	}

	return -1, -1
}

//碰
func (r *RoomBase) CheckCanPeng(index int, cid int, card *Card) bool {

	if card == nil || cid != card.ID { // 是不是当前操作的牌
		return false
	}

	if cid < 0 || cid > 33 {
		return false // 牌不对
	}

	ccount := r.Seats[index].Cards.List.Count //手牌数量不对
	if ccount < 2 || ccount == 3 || ccount == 5 || ccount == 6 || ccount == 8 || ccount == 9 || ccount == 11 || ccount == 12 {
		return false
	}

	c := r.GetUserCardCount(index, cid) // 判断手牌中是否存在
	if c >= 2 {
		return true
	} else {
		return false
	}
}

//吃
func (r *RoomBase) CheckCanChow(index int, cids []int, card *Card) bool {

	if card == nil || cids[0] != card.ID && cids[1] != card.ID && cids[2] != card.ID { // 是不是当前操作的牌
		return false
	}

	if cids[0] < 0 || cids[0] > 26 || cids[1] < 0 || cids[1] > 26 || cids[2] < 0 || cids[2] > 26 {
		return false // 牌不对
	}

	ccount := r.Seats[index].Cards.List.Count //手牌数量不对

	if ccount < 2 || ccount%3 == 0 {
		return false
	}

	c1 := r.GetUserCardCount(index, cids[0])
	c2 := r.GetUserCardCount(index, cids[1])
	c3 := r.GetUserCardCount(index, cids[2])
	if c1 > 1 {
		c1 = 1
	}
	if c2 > 1 {
		c2 = 1
	}
	if c3 > 1 {
		c3 = 1
	}
	cc := c1 + c2 + c3

	if cc == 2 || cc == 3 {
	} else {
		return false
	}
	InsertSort(cids)
	if cids[2]-cids[1] == cids[1]-cids[0] && cids[1]-cids[0] == 1 {
		return true
	} else {
		return false
	}
}

//region 直接插入排序
func InsertSort(list []int) {
	var temp int
	var i int
	var j int
	// 第1个数肯定是有序的，从第2个数开始遍历，依次插入有序序列
	for i = 1; i < len(list); i++ {
		temp = list[i] // 取出第i个数，和前i-1个数比较后，插入合适位置
		// 因为前i-1个数都是从小到大的有序序列，所以只要当前比较的数(list[j])比temp大，就把这个数后移一位
		for j = i - 1; j >= 0 && temp < list[j]; j-- {
			list[j+1] = list[j]
		}
		list[j+1] = temp
	}
}

func CalcAbs(a int) (ret int) {
	ret = (a ^ a>>31) - a>>31
	return
}

//查找牌
func (r *RoomBase) FindCard(index int, cids []int) bool {
	seat := r.Seats[index]
	AllFind := true
	for c := 0; c < len(cids); c++ {
		fd := false
		for i := 0; i < seat.Cards.List.Count; i++ {
			if *seat.Cards.List.Index(i) != nil {
				card := (*seat.Cards.List.Index(i)).(*Card)
				if card.ID == cids[c] {
					fd = true
					break
				}
			}

		}

		if !fd {
			AllFind = false
			break
		}
	}
	return AllFind
}

//自动出牌
func (r *RoomBase) AutoPutCard() {
}

//自动操作
func (r *RoomBase) AutoTool() {
	//	seat := r.Seats[r.CurToolIndex]
	//	for i := 0; i < r.NeedWaitTool.Count; i++ {
	//		nw := (*r.NeedWaitTool.Index(i)).(*NeedWait)
	//		if string(nw.User.ID) == seat.UID {
	//			if nw.Win > 0 {
	//				// 胡牌
	//				r.WaitTimeCount = 0
	//				r.WinCard(seat.User, r.CurCard.ID)
	//				break
	//			} else if nw.Kong > 0 {
	//				// 杠牌
	//				r.WaitTimeCount = 0
	//				r.KongCard(seat.User, r.CurCard.ID)
	//				break
	//			} else if nw.Peng > 0 {
	//				// 碰牌
	//				card := r.GetUserCard(seat.Cards.List, r.CurCard.Type, r.CurCard.Num)
	//				r.WaitTimeCount = 0
	//				r.PengCard(nw.User, card.ID)
	//				break
	//			} else if nw.Chow > 0 {
	//				// 吃牌
	//				r.WaitTimeCount = 0
	//				break
	//			}
	//		}
	//	}
}

//得到下一张牌
func (r *RoomBase) GetNextCard(forward bool) (*Card, mjgame.ACKBC_GetCard) {
	var fromLast bool
	var count, cid int
	seat := r.Seats[r.CurIndex]
	fromLast = forward

	bcGetCard := &mjgame.ACKBC_GetCard{}

	// TODO :  EndBlank 需要计算
	for {
		card := r.GetNewCard(fromLast, r.CurIndex)
		if card != nil {
			if count > 0 {
				forward = false
			}
			count = 0
			if card.Type == H { //是否花牌，花牌重新抓取
				seat.Cards.Hua.Add(card)
				count++
				cid, fromLast = card.ID, false
			} else {
				seat.Cards.List.Add(card)
				cid = -1
			}

			if forward {
				(r.StlCtrl).(st.IStatement).AddTool(st.T_Mo, r.CurIndex, -1, []int{card.ID})
			} else {
				(r.StlCtrl).(st.IStatement).AddTool(st.T_MoBack, r.CurIndex, -1, []int{card.ID})
			}

			bcGetCard.Index = int32(r.CurIndex)
			bcGetCard.Cid = int32(cid)
			bcGetCard.FromLast = forward
			bcGetCard.Tool = []int32{0, 0, 0, 0}

			if card.Type == H {
				r.MToolChecker.SetTools(r.CurIndex, []int{-1, -1, -1, -1, -1, -1, -1, -1}) // 一个用户只能做一次操作, 操作后不管是否成功都禁止其他任何操作
				r.BCMessage(mjgame.MsgID_MSG_ACKBC_GetCard, bcGetCard)

				rec := [3]interface{}{r.CurIndex, cid, forward}
				r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_GetCard, rec)
			} else { //围观用户
				for _, v := range r.WatchSeats {
					v.SendMessage(mjgame.MsgID_MSG_ACKBC_GetCard, bcGetCard)
				}
				return card, *bcGetCard
			}
		} else {
			break
		}
	}

	return nil, *bcGetCard
}

//玩家准备
func (r *RoomBase) Ready(user *user.User) {
	var readyCount int32
	for _, seat := range r.Seats {
		if seat.State == int(mjgame.StateID_UserState_Ready) {
			readyCount++
		}
	}

	if readyCount == int32(r.Rules.SeatLen) {
		startUser, index := r.GetFirstSitSeatInfo()
		if startUser != nil {
			notifyStartGame := &mjgame.NotifyStartGame{
				Uid: strconv.Itoa(startUser.ID),
			}
			startUser.SendMessage(mjgame.MsgID_MSG_NOTIFY_START_GAME, notifyStartGame)
			r.Seats[index].Message = &Message{
				Id:      mjgame.MsgID_MSG_NOTIFY_START_GAME,
				Content: notifyStartGame,
			}
		}
	}
	ackReady := &mjgame.ACKBC_Ready{
		ReadyCount: readyCount,
		UID:        strconv.Itoa(user.ID),
		MSG:        "",
	}
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Ready, ackReady)

	index := r.GetSeatIndexById(user.ID)
	if index >= 0 {
		r.Seats[index].Message = &Message{
			Id:      mjgame.MsgID_MSG_ACKBC_Ready,
			Content: ackReady,
		}
	}
}

//出牌
func (r *RoomBase) PutCard(uIndex int, pcard *Card) {

}

//胡牌
func (r *RoomBase) WinCard(user *user.User, cid int) {
}

//杠牌
func (r *RoomBase) KongCard(user *user.User, cid int) {

}

//碰牌
func (r *RoomBase) PengCard(a *user.User, cid int) {

}

//吃牌
func (r *RoomBase) ChowCard() {

}

//听牌
func (r *RoomBase) TingCard() {

}

//过
func (r *RoomBase) Pass() {
}

//流局
func (r *RoomBase) Draw() {

}

//每次操作后的结算
func (r *RoomBase) ToolTotal() {

}

//一局结束最后结算
func (r *RoomBase) RoundTotal() {

}

//全部局结束后的结算
func (r *RoomBase) EndGameTotal() {

}

//下一局(一局结束)
func (r *RoomBase) NextRound() {

}

//下一大圈
func (r *RoomBase) NextBigRound() {

}

//解散房间
func (r *RoomBase) Disband() {

}

//更新用户分数
func (r *RoomBase) UpdateScore(scores []int32) {
	for k := 0; k < len(scores); k++ {
		if r.Seats[k] != nil && r.Seats[k].User != nil && r.Seats[k].Accumulation != nil {
			r.Seats[k].Accumulation.Score += scores[k]
		}
	}

}

//得到房间用户信息
func (r *RoomBase) GetRoomUser() *mjgame.ACK_Room_User {
	userList := make([]*mjgame.ACK_User_Info, 0)
	for _, v := range r.Seats {
		var user mjgame.ACK_User_Info
		if v.User != nil {
			if v.State == int(mjgame.StateID_UserState_Normal) {
				continue
			}
			var ip string
			if v.User.Conn != nil {
				ip = v.User.Conn.RemoteAddr().String()
			}

			user.Uid = strconv.Itoa(v.User.ID)
			user.Index = int32(v.Index)
			user.Ip = ip
			user.Name = v.User.NickName
			user.Icon = v.User.Icon
			user.Robot = int32(v.User.IsRobot)
			user.Coin = int32(v.User.Coin)
			user.GPS_LAT = v.User.GPS_LAT
			user.GPS_LNG = v.User.GPS_LNG
			user.Diamond = int32(v.User.Diamond)
			user.RoomId = int32(r.RoomId)
			user.State = int32(v.User.State)
			user.Sex = int32(v.User.Sex)
			if v.State == int(mjgame.StateID_UserState_Ready) {
				user.Ready = true
			}
			if v.Accumulation != nil {
				user.Score = int32(v.Accumulation.Score)
			}
			if v.User.State == def.Offline {
				if v.OfflineTime.Unix() > 0 {
					if v.OfflineTime.Add(def.KickTimeDuration * time.Second).Before(time.Now()) {
						user.CanKick = true
					}

					user.OfflineTime = int32(v.OfflineTime.Add(def.KickTimeDuration*time.Second).Unix() - time.Now().Unix())
					if user.OfflineTime < 0 {
						user.OfflineTime = 0
					}
				}
			}
			userList = append(userList, &user)
		}
	}

	for _, v := range r.WatchSeats {
		var user mjgame.ACK_User_Info
		if v != nil {
			var ip string
			if v.Conn != nil {
				ip = v.Conn.RemoteAddr().String()
			}

			user.Uid = strconv.Itoa(v.User.ID)
			user.Index = -1
			user.Ip = ip
			user.Name = v.User.NickName
			user.Icon = v.User.Icon
			user.Robot = int32(v.User.IsRobot)
			user.Coin = int32(v.User.Coin)
			user.GPS_LAT = v.User.GPS_LAT
			user.GPS_LNG = v.User.GPS_LNG
			user.Diamond = int32(v.User.Diamond)
			user.RoomId = int32(r.RoomId)
			user.State = int32(v.User.State)
			user.Sex = int32(v.User.Sex)
			userList = append(userList, &user)
		}
	}

	roomUser := mjgame.ACK_Room_User{
		RID:   int32(r.RoomId),
		Users: userList,
	}

	return &roomUser
}

//得到房间信息
func (r *RoomBase) GetRoomInfo() *mjgame.ACK_Room_Info {

	ack := mjgame.ACK_Room_Info{
		RoomId:     int32(r.RoomId),
		Type:       int32(0),
		City:       int32(0),
		Level:      int32(0),
		Rules:      r.Rules.Rules,
		SeatCount:  int32(len(r.Seats)),
		Starting:   r.IsRun,
		RoundCount: int32(r.RoundCount),
	}
	return &ack
}

//广播消息
func (r *RoomBase) BCMessage(msgId mjgame.MsgID, pb proto.Message) {
	b, err := proto.Marshal(pb)
	if err != nil {
		fmt.Println("marshaling error: ", msgId, err)
	}

	m := mjgame.Message{ID: int32(msgId), MSG: b}
	data, err := proto.Marshal(&m)
	if err != nil {
		fmt.Println("marshaling message error: ", err)
	}

	for _, v := range r.Seats {
		if v != nil && v.User != nil && v.User.Conn != nil {
			v.User.Conn.WriteMsg(data)
		}
	}

	for _, v := range r.WatchSeats {
		if v != nil && v.User != nil && v.Conn != nil {
			v.Conn.WriteMsg(data)
		}
	}
}

//////////////////////////////////////
// 发送玩家牌消息
func (r *RoomBase) SendSeatCard(userId int) {
	allUserCards := make([]*mjgame.SeatCard, 0)

	//fmt.Println(r.Seats, "seats")
	var index int = -1
	for i, v := range r.Seats { // TODO : 这里可能报错

		if v.Cards == nil { //零时Fix
			continue
		}

		ack_card := &mjgame.SeatCard{
			Seat:    int32(v.Index),
			ListLen: int32(v.Cards.List.Count),
			Chow:    r.GetListArray(v.Cards.Chow),
			Out:     r.GetListArray(v.Cards.Out),
			Hua:     r.GetListArray(v.Cards.Hua),
			Hu:      r.GetListArray(v.Cards.Hu),
		}

		pengCards := r.GetListArray(v.Cards.Peng)
		for i, v := range pengCards {
			if i%3 == 0 {
				ack_card.Peng = append(ack_card.Peng, v)
			}
		}

		kongCards := r.GetListArray(v.Cards.Kong)
		for i, v := range kongCards {
			if i%4 == 0 {
				ack_card.Kong = append(ack_card.Kong, v)
			}
		}

		if v.User != nil && v.User.ID == userId {
			ack_card.List = r.GetListArray(v.Cards.List)
			index = i
			ack_card.LastCardId = int32(v.LastCardID)
			ack_card.Type = int32(v.LastToolType)
		}
		allUserCards = append(allUserCards, ack_card)
	}

	ack := mjgame.ACK_User_SeatCard{
		Cards: allUserCards,
	}

	if index >= 0 {
		r.Seats[index].User.SendMessage(mjgame.MsgID_MSG_ACK_User_SeatCard, &ack)
	} else {
		user.BCMessage(mjgame.MsgID_MSG_ACK_User_SeatCard, &ack, r.WatchSeats)
	}
}

func (r *RoomBase) RecordSeatCard() {

	allUserCards := make([]*mjgame.SeatCard, 0)

	//fmt.Println(r.Seats, "seats")
	for _, v := range r.Seats { // TODO : 这里可能报错

		if v.Cards == nil { //零时Fix
			continue
		}

		ack_card := &mjgame.SeatCard{
			Seat:    int32(v.Index),
			ListLen: int32(v.Cards.List.Count),
			Chow:    r.GetListArray(v.Cards.Chow),
			Out:     r.GetListArray(v.Cards.Out),
			Hua:     r.GetListArray(v.Cards.Hua),
			Hu:      r.GetListArray(v.Cards.Hu),
		}

		pengCards := r.GetListArray(v.Cards.Peng)
		for i, v := range pengCards {
			if i%3 == 0 {
				ack_card.Peng = append(ack_card.Peng, v)
			}
		}

		kongCards := r.GetListArray(v.Cards.Kong)
		for i, v := range kongCards {
			if i%4 == 0 {
				ack_card.Kong = append(ack_card.Kong, v)
			}
		}

		ack_card.List = r.GetListArray(v.Cards.List)
		ack_card.LastCardId = int32(v.LastCardID)
		ack_card.Type = int32(v.LastToolType)
		allUserCards = append(allUserCards, ack_card)
	}

	ack := mjgame.ACK_User_SeatCard{
		Cards: allUserCards,
	}

	//fmt.Println("jilulefapai.................")

	r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACK_User_SeatCard, ack)
}

//得到座位位置
func (r *RoomBase) GetSeatIndexById(userId int) int {
	for i, seat := range r.Seats {
		if seat != nil && seat.UID == strconv.Itoa(userId) {
			return i
		}
	}
	return -1
}

//获取当前局，最早坐下玩家信息
func (r *RoomBase) GetFirstSitSeatInfo() (*user.User, int) {
	var list []int
	var mapUserTime = make(map[int]int)

	for k, v := range r.Seats {
		if v.User != nil {
			key := int(v.CreateTimeStamp)
			list = append(list, int(v.CreateTimeStamp))
			if _, ok := mapUserTime[key]; !ok {
				mapUserTime[key] = k
			}
		}
	}

	if len(list) > 0 {
		sort.Ints(list)
	}

	for _, v := range list {
		if v > 0 {
			if index, ok := mapUserTime[v]; ok {
				return r.Seats[index].User, index
			}
		}
	}

	return nil, -1
}

//通知用户操作
func (r *RoomBase) NotifyPutCard(index int, t int, itime int) {
	if r.Seats[index].User == nil {
		return
	}

	tools := []int32{0, 0} // 自摸 暗杠
	/*if r.CheckWin(index, nil) > 0 { // 自摸
		//fmt.Println("自摸 ", r.Seat[index].agent.Name)
		tools[0] = 1
	}*/

	/*kongType := r.CheckAnKong(index, 0)
	if kongType > 0 { // 暗杠 // 碰杠
		tools[1] = int32(kongType)
	}*/

	ack := mjgame.ACKBC_CurPlayer{
		Seat:      int32(index),
		Type:      int32(t),
		Tool:      tools,
		RoundTime: int32(itime),
	}
	r.Seats[index].User.SendMessage(mjgame.MsgID_MSG_ACKBC_CurPlayer, &ack)
}

//检查是否暗杠
func (r *RoomBase) CheckAnKong(uIndex int, missType int) int {
	user := r.Seats[uIndex].Cards
	kongtype := -1
	//万, 筒, 条 w=0 b=1 t=2
	var types [27]int
	for i := 0; i < user.List.Count; i++ {
		if *user.List.Index(i) != nil {
			pc := (*user.List.Index(i)).(*Card)

			if pc.Type == missType {
				continue
			}

			types[pc.ID]++
			if types[pc.ID] >= 4 {
				kongtype = st.T_AnKong
				break
			}
		}
	}

	for i := 0; i < user.Peng.Count; i++ {
		if *user.Peng.Index(i) != nil {
			pc := (*user.Peng.Index(i)).(*Card)

			/*if pc.Type == missType {
				continue
			}*/

			types[pc.ID]++
			if types[pc.ID] >= 4 {
				kongtype = st.T_PengKong
				break
			}
		}

	}
	return kongtype
}

//得到座位位置
func (r *RoomBase) GetWatchSeat(userId int) int {
	for i, seat := range r.WatchSeats {
		if seat != nil && seat.ID == userId {
			return i
		}
	}

	return -1
}

// 获取当前所有玩家手牌
func (r *RoomBase) GetAllSeatCards() []*mjgame.SeatCard {
	allUserCards := make([]*mjgame.SeatCard, 0)

	for _, v := range r.Seats {
		//由于可能座位牌信息为空，做个判断
		if v == nil || v.User == nil || v.Cards == nil {
			continue
		}
		if r.Type == int32(mjgame.MsgID_GTYPE_Pinshi) {
			if v.JoinPlay == false {
				continue
			}
		}
		ack_card := &mjgame.SeatCard{
			Seat:    int32(v.Index),
			ListLen: int32(v.Cards.List.Count),
			Chow:    r.GetListArray(v.Cards.Chow),
			Peng:    r.GetListArray(v.Cards.Peng),
			Kong:    r.GetListArray(v.Cards.Kong),
			Out:     r.GetListArray(v.Cards.Out),
			Hua:     r.GetListArray(v.Cards.Hua),
			Hu:      r.GetListArray(v.Cards.Hu),
			List:    r.GetListArray(v.Cards.List),
		}
		allUserCards = append(allUserCards, ack_card)
	}

	return allUserCards
}

// 获取当前所有玩家手牌
func (r *RoomBase) GetRecordAllSeatCards() []*mjgame.SeatCard {
	allUserCards := make([]*mjgame.SeatCard, 0)

	for _, v := range r.Seats {
		//由于可能座位牌信息为空，做个判断
		if v == nil || v.Cards == nil {
			continue
		}

		ack_card := &mjgame.SeatCard{
			Seat:    int32(v.Index),
			ListLen: int32(v.Cards.List.Count),
			Hu:      r.GetListArray(v.Cards.Hu),
			Hua:     r.GetListArray(v.Cards.Hua),
			List:    r.GetListArray(v.Cards.List),
			Chow:    r.GetListArray(v.Cards.Chow),
		}

		pengCards := r.GetListArray(v.Cards.Peng)
		for i, v := range pengCards {
			if i%3 == 0 {
				ack_card.Peng = append(ack_card.Peng, v)
			}
		}

		kongCards := r.GetListArray(v.Cards.Kong)
		for i, v := range kongCards {
			if i%4 == 0 {
				ack_card.Kong = append(ack_card.Kong, v)
			}
		}

		allUserCards = append(allUserCards, ack_card)
	}

	return allUserCards
}

//判断当前房间是否还有人
func (r *RoomBase) IsEmpty() bool {
	for _, v := range r.Seats {
		if v.User != nil {
			return false
		}
	}

	for _, v := range r.WatchSeats {
		if v != nil {
			return false
		}
	}

	return true
}

//判断玩家是否是已经坐下
func (r *RoomBase) IsSitDownUser(uid string) bool {
	for _, v := range r.Seats {
		if v.UID == uid {
			return true
		}
	}
	return false
}

//获取当前房间roomrecord用于测试
func (r *RoomBase) GetRoomRecord(user *user.User) {
	message := r.RoomRecord
	user.SendMessage(mjgame.MsgID_MSG_ACK_MSG, &mjgame.ACK_MSG{Type: 1, MSG: message})
}

func (r *RoomBase) NewBattleRecord() {
	r.BattleRecord = &RecordMgs{
		Msgs: make([]*PerRecords, 0),
	}
}

func (r *RoomBase) SaveBattleRecord(pos int, msgId mjgame.MsgID, body interface{}) {

	content := make(map[mjgame.MsgID]interface{})
	content[msgId] = body

	r.BattleRecord.Msgs = append(r.BattleRecord.Msgs, &PerRecords{
		P: pos,
		C: &content,
	})
}
