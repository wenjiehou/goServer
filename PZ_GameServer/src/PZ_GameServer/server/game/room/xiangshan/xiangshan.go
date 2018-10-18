// 象山麻将 GameType : 3010
package xiangshan

import (
	"fmt"
	"math"
	"math/rand"

	"reflect"
	"strconv"
	"time"

	"PZ_GameServer/common/util"
	al "PZ_GameServer/common/util/arrayList"
	"PZ_GameServer/model"
	"PZ_GameServer/protocol/def"
	"PZ_GameServer/protocol/pb"
	"PZ_GameServer/server/game/common"
	"PZ_GameServer/server/game/error"

	rb "PZ_GameServer/server/game/roombase"
	st "PZ_GameServer/server/game/statement"

	"PZ_GameServer/server/user"
	"encoding/json"
	"sync"
)

//开局筛子判断常量
var (
	FrontPoints      = []int{2, 6, 10}
	BackPoints       = []int{4, 8, 12}
	FacePoints       = []int{3, 7, 11}
	SelfPoints       = []int{5, 9}
	IFCXiangShanType = reflect.TypeOf(&RoomXiangshan{})
	Qin12Hun8        = 128 //=(清12混8)
	Qin10Hun6        = 106 //=(清10混6)
)

//投票常量
const (
	Agree = iota + 1
	DisApprove
)

//开局类型
const (
	CalcRound = iota + 1 //按局数
	CalcTime             //按时间
)

//解散状态
const (
	DisbandSuccess = iota + 1 //解散成功
	DisbandFail               //解散失败
)

//最少操作数
const (
	Least = iota + 1
)

//用户操作状态
const (
	Valid   = iota //有效
	InValid        //无效
)

//玩家初始化手牌数
const (
	InitUserCardsNumber = 13
)

// 西周规则
var XiangShan_RoomRule = rb.RoomRule{
	GameType:           3010,            //
	Create_NeedDiamond: 100,             // 创建房间需要的钻石
	SeatLen:            4,               // 座位数量  2, 3, 4
	DefCardLen:         13,              // 默认手牌数量 13
	AllCardLen:         144,             // 牌的数量
	Card_W:             1,               // 默认带万 1代表一次
	Card_B:             1,               // 默认带饼
	Card_T:             1,               // 默认带条
	Card_F:             1,               // 默认带风
	Card_J:             1,               // 默认带箭 (中发白)
	Card_H:             1,               // 默认带花
	Card_Else:          []int{},         // 特殊牌
	CanLaiZi:           0,               // 赖子数量
	CanPeng:            1,               // 可以碰
	CanChow:            1,               // 可以吃
	CanKong:            1,               // 可以直杠
	CanAnKong:          1,               // 可以暗杠
	CanPengKong:        1,               // 可以碰杠
	CanTing:            1,               // 可以听
	CanWin:             1,               // 可以直胡
	CanZiMo:            1,               // 可以自摸胡
	MaxWinCount:        1,               // 最大胡牌数量<0为不限次数, 大众麻将为1
	MaxTime:            0,               // 最大时间(<=0 为不限时间)
	MaxRound:           0,               // 最大局数(<=0 为不限局数)
	BaseScore:          1,               // 基本分
	MaxTai:             0,               // 封顶台数(<=0 为不限)
	WinNeedTai:         0,               // 胡牌最小台数 (<=0 为不限)
	Rules:              []int32{},       // 全部规则(包含特殊规则)
	Temp:               RoomXiangshan{}, // 模板
}

type RoomXiangshan struct {
	rb.RoomBase                          //
	ChengBao         []ChengBaoSeat      //
	FengQuan         int                 // 风圈(0-3)东南西北
	Status           int                 // 状态
	Bankers          []int               // 风圈变了,清空
	Votes            []int               // 投票  -1未操作   0反对   1同意
	VoteTimeCount    int                 // 投票时间
	IsKongHu         bool                // 是否拉杠胡
	KongHuCardID     int                 // 拉杠胡cardid
	IsContinueBanker bool                // 是否连庄
	RoundResult      *mjgame.ACKBC_Total // 记录每一局信息，做短线重连使用,后一局覆盖前一局
	KickUsers        []*KickInfo         // 被踢掉的玩家信息
	QuitSitUsers     map[int]rb.SeatBase // 退出参与玩家信息
	FanType          int                 // 番型   1=(清12混8) 2=(清10混6)
	Mux              sync.RWMutex        // map 读写锁
}

// 承包结构变了
type ChengBaoSeat struct {
	Index int   // seat index
	Seat  []int // 承包
}

// 踢人
type KickInfo struct {
	UserID   int
	Position int
}

// 座位
type Seat_XiangShan struct {
	rb.SeatBase
}

// 初始化
func (r *RoomXiangshan) Init() {
	r.RoomBase.Init()

	r.State = 0
	r.IsDraw = false
	r.IsRun = true
	r.WinUserCount = 0
	r.IsKongHu = false
	r.RoundResult = nil
	r.RoomBase.MToolChecker = rb.ToolChecker{}
	r.RoomBase.MToolChecker.Init(4)
	r.RoomBase.RoomRecord = ""
	r.RoomBase.RoundTotaled = false
	r.Show = false
	r.Votes = []int{0, 0, 0, 0}
	r.RoomBase.VoteStarter = -1
	r.KickUsers = []*KickInfo{}

	if r.RoundCount == 0 {
		for _, v := range r.Seats {
			v.Accumulation = &rb.Accumulation{}
		}
		r.StartTime = int(time.Now().Unix())
	}

	r.ChengBao = make([]ChengBaoSeat, r.Rules.SeatLen)
	for i := 0; i < r.Rules.SeatLen; i++ {
		r.ChengBao[i].Seat = make([]int, r.Rules.SeatLen)
	}
}

//定时器
func (r *RoomXiangshan) TimeTicker() {
	var flag bool

	for {

		if r.StopTicker {
			flag = true
			break
		}

		time.Sleep(1 * time.Second)

		//select {
		//case <-time.After(1 * time.Second):

		//房间到时间未开始解散
		if r.RoundCount == 0 && !r.IsRun && r.StartTime == 0 {
			leftTime := def.WaitStartTime - (time.Now().Unix() - r.CreateTime)
			if leftTime <= 0 {
				r.ClearRoomUserRoomId()
				r.BCMessage(mjgame.MsgID_MSG_NOTIFY_DESTORY_ROOM, &mjgame.NotifyDestoryRoom{RoomId: int32(r.RoomId)})
				flag = true
				r.StopTicker = true
				fmt.Println("房间到时间未开始解散")
				r.Mux.Lock()
				rb.ChanRoom <- r.RoomId //销毁房间
				r.Mux.Unlock()
				break
			}
		}

		//投票解散
		if r.VoteStarter >= 0 {
			r.VoteTimeCount++
			if r.VoteTimeCount >= r.VoteTimeOut {
				fmt.Println("解散")
				r.VoteStarter = -1
				r.IsRun = false
				disApproveCount := r.GetDisApproveCount()
				if disApproveCount < r.Rules.SeatLen/2 {
					fmt.Println("投票解散")
					r.BCMessage(mjgame.MsgID_MSG_NOTIFY_DISBAND, &mjgame.NotifyDisband{RoomId: int32(r.RoomId), Result: DisbandSuccess})
					flag = true
					r.StopTicker = true
					r.DestoryRoom()
					r.Mux.Lock()
					rb.ChanRoom <- r.RoomId //销毁房间
					r.Mux.Unlock()

				} else {
					r.BCMessage(mjgame.MsgID_MSG_NOTIFY_DISBAND, &mjgame.NotifyDisband{RoomId: int32(r.RoomId), Result: DisbandFail})
					r.VoteTimeCount = 0
				}
			}
		}

		//时间到了，自动解散
		if r.Rules.MaxTime > 0 && r.StartTime > 0 {

			leftTime := (r.Rules.MaxTime * 60) - (int(time.Now().Unix()) - r.StartTime)

			if !r.IsRun { //当前处于俩局之间,时间到了自动结算
				if leftTime <= 0 {
					fmt.Println("当前处于俩局之间,时间到了自动结算 ", r.RoomId)
					r.BCMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrCurRoundHasOver)
					list := r.GetSummaryList()
					r.BCMessage(mjgame.MsgID_MSG_NOTIFY_SUMMARY, &list)
					r.ClearRoomUserRoomId()
					flag = true
					r.StopTicker = true
					r.Mux.Lock()
					rb.ChanRoom <- r.RoomId //销毁房间
					r.Mux.Unlock()
					break
				}

			} else {
				//当前正在进行中，有玩家掉线超过180s, 且游戏时间到了游戏流局，房间结束
				if r.RoundCanFinish() && leftTime <= 0 {
					fmt.Println("有玩家掉线超过180s,游戏流局，房间结束 ", r.RoomId)
					flag = true
					r.StopTicker = true
					r.DestoryRoom()
					r.Mux.Lock()
					rb.ChanRoom <- r.RoomId //销毁房间
					r.Mux.Unlock()
				}
			}

		}

		//踢人
		kickIndexs := r.GetKickIndex()
		if len(kickIndexs) > 0 {
			data := &mjgame.NotifyKick{}
			data.Indexs = kickIndexs
			r.BCMessage(mjgame.MsgID_MSG_NOTIFY_KICK, data)
		}

		//}
		if flag {
			fmt.Println("timetick flag break")
			break
		}
	}
}

// 检查是否可以开始
func (r *RoomXiangshan) CheckCanStart() (bool, *mjgame.ErrorItem) {
	for i := 0; i < len(r.Seats); i++ {
		if r.Seats[i] == nil || r.Seats[i].User == nil {
			return false, error.ErrSomePeopleNotReady
		}
	}

	if r.Rules.MaxRound > 0 {
		if r.RoundCount >= r.Rules.MaxRound {
			return false, error.ErrCurRoundHasOver
		}
	}

	// 判断有没有超时
	if r.Rules.MaxTime > 0 {
		if r.StartTime > 0 {
			if (r.Rules.MaxTime*60)-(int(time.Now().Unix())-r.StartTime) < 0 {
				return false, error.ErrCurRoundHasOver
			}
		}
	}

	return true, nil
}

// 添加承包关系
func (r *RoomXiangshan) AddChengBao(index int, tindex int) {
	if index == tindex { //自己不能承包自己
		return
	}
	r.ChengBao[index].Seat[tindex]++
	if r.ChengBao[index].Seat[tindex] == 3 || r.ChengBao[index].Seat[tindex] == 6 { //形成承包关系
		r.ChengBao[tindex].Seat[index] += 3
	}

	// TODO 一旦形成承包则在这里需要广播
}

//初始化每局信息
func (r *RoomXiangshan) InitRound() {
	var rollIndex, bankerIndex, points, state int
	var users = make([]*mjgame.ACK_User_Info, 0)

	if len(r.Bankers) == 0 { // fix 最后一个连庄  一圈的第一局
		rollIndex, bankerIndex, points = r.GetStartInfo()

		if r.RoundCount == 0 { //  fix 第二局后不换位置
			users = r.ChangeUsersPosition()
			r.FirstZhuangIndex = bankerIndex
		} else {
			points = 0
			bankerIndex = r.FirstZhuangIndex
		}

		r.BankerIndex = bankerIndex
		r.Seats[bankerIndex].IsZhuang = true

		r.AddTool(st.T_Dice, rollIndex, -1, []int{})
	} else {
		if r.IsContinueBanker || r.IsDraw {
			//连庄
		} else {
			//不是连庄
			r.BankerIndex++
			r.BankerIndex = r.BankerIndex % 4
		}

		bankerIndex, state = r.BankerIndex, 1
		points, rollIndex = 0, 0

	}
	var index int = bankerIndex
	for i := 0; i < len(r.Seats); i++ {
		r.Seats[index].Direct = i
		index++
		index = index % r.Rules.SeatLen
	}

	r.AddTool(st.T_Zhuang, bankerIndex, -1, []int{})
	r.RoomBase.RoomRecord += "定庄(" + r.Seats[bankerIndex].User.NickName + ")\r\n"

	ack := mjgame.ACKBC_Start{
		RoomId:          int32(r.RoomId),
		Points:          int32(points), //
		State:           int32(state),
		RollIndex:       int32(r.Seats[rollIndex].Index),
		ZhuangIndex:     int32(r.Seats[r.BankerIndex].Index),
		Direction:       int32(r.FengQuan),
		Users:           users,
		RoundCount:      int32(r.RoundCount),
		TotalRoundCount: int32(r.Rules.MaxRound),
	}

	if r.Rules.MaxRound > 0 {
		ack.Type = CalcRound
	}

	if r.Rules.MaxTime > 0 {
		ack.Type = CalcTime
		ack.LeftTime = int64((r.Rules.MaxTime * 60) - (int(time.Now().Unix()) - r.StartTime))
		if ack.LeftTime < 0 {
			ack.LeftTime = 0
		}
	}

	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Start, &ack) // 广播游戏开始消息

	recordAck := mjgame.ACKBC_Start{
		Points:      ack.Points, //
		ZhuangIndex: ack.ZhuangIndex,
		Direction:   ack.Direction,
		RoundCount:  ack.RoundCount,
		LeftTime:    ack.LeftTime,
	}

	for _, v := range ack.Users {
		recordAck.Users = append(recordAck.Users, &mjgame.ACK_User_Info{
			Uid:   v.Uid,
			Index: v.Index,
		})
	}

	//战斗记录 开始
	r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Start, recordAck)
}

// 一局结算
func (r *RoomXiangshan) RoundTotal() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	//fmt.Println("to roundtotal ", r.RoomId)
	if r.RoomBase.RoundTotaled {
		//fmt.Println("已经结算过.", r.RoomId)
		return
	}
	r.RoomBase.RoundTotaled = true
	r.RoomBase.RoomRecord += "开始结算\r\n"

	huCard := -1
	var huIndexes []int
	var lastTool *st.OnceRecord

	if r.StlCtrl != nil && (r.StlCtrl).(*XiangShan_Statement) != nil {
		lastTool = (r.StlCtrl).(*XiangShan_Statement).Get(0) //TODO 这里居然会崩溃
	}

	if lastTool != nil && lastTool.Tool.ToolType == st.T_Hu_ZiMo {
		huIndex := lastTool.Tool.Index
		huCard = lastTool.Tool.Val[0]
		if huIndex < 0 || huCard < 0 {
			fmt.Println("huCard ", huCard)
			return
		}
		huIndexes = append(huIndexes, huIndex)
	}
	if lastTool != nil && lastTool.Tool.ToolType == st.T_Hu {
		huIndexes, huCard = r.GetMultiWinInfo()
	}

	if r.RoundCount == 0 {
		room := &model.Room{
			Type:         int(mjgame.MsgID_GTYPE_ZheJiang_XiangShan),
			UserID:       r.CreateUserId,
			Rules:        util.IntArrToString(r.Rules.Rules),
			ServerRoomID: r.RoomId,
			UniqueCode:   r.UniqueCode,
		}
		err := model.GetRoomModel().Create(room)
		if err == nil {
			r.DbRoomId = room.ID
		}
	}

	r.RoundCount++
	flag, _ := r.CheckCanStart()

	//更新用户状态(断线重连)
	for _, seat := range r.Seats {
		seat.State = int(mjgame.StateID_GameState_Total)
	}

	var scores = make([]int32, 4)
	var ackTotal *mjgame.ACKBC_Total
	if lastTool != nil && lastTool.Tool.ToolType == st.T_Draw {
		var rewards []*mjgame.Reward
		for i := 0; i < r.Rules.SeatLen; i++ {
			rewards = append(rewards, &mjgame.Reward{})
		}
		ackTotal = &mjgame.ACKBC_Total{
			Finished:   !flag,
			RoundCount: int64(r.RoundCount),
			Reward:     rewards,
		}
	} else {
		var total = st.TotalResult{
			TotalScore: make([]int32, 4),
			TotalMsg:   make([]string, 4),
		}
		totalTai := make([]int32, 4)

		r.IsContinueBanker = false
		for _, v := range huIndexes {
			result := r.FanCalc(v, huCard)
			totalTai[v] = result.TotalTai
			for i, v := range result.TotalScore {
				total.TotalScore[i] += v
			}
			for i, v := range result.TotalMsg {
				if len(v) > 0 {
					total.TotalMsg[i] = v
				}
			}
			if r.BankerIndex == v {
				r.IsContinueBanker = true
			}
		}

		r.UpdateScore(total.TotalScore) // 更新分数

		maxScoreIndex := r.GetMaxIndex(total.TotalScore)

		var rewards = make([]*mjgame.Reward, 0)
		for k, v := range total.TotalScore {
			if r.Seats[k] != nil && r.Seats[k].Accumulation != nil {
				if r.CostType == rb.Jinbi {
					r.Seats[k].User.Coin += int(v)

					if r.Seats[k].User.Coin <= 0 {
						r.Seats[k].User.Coin = rb.Jiuji_coin
						r.Seats[k].User.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrBuchongCoin)
					}
					model.GetUserModel().Save(r.Seats[k].User.User)

					o := &mjgame.Reward{
						Score:      v,
						TotalScore: int32(r.Seats[k].User.Coin),
					}
					rewards = append(rewards, o)

				} else {
					o := &mjgame.Reward{
						Score:      v,
						TotalScore: r.Seats[k].Accumulation.Score,
					}
					rewards = append(rewards, o)
				}

			}

		}

		ackTotal = &mjgame.ACKBC_Total{
			WinSeat:    int32(maxScoreIndex),
			WinCard:    int32(huCard), // 这里要读取记录
			Tai:        totalTai,
			Msg:        total.TotalMsg,
			Reward:     rewards,
			Finished:   !flag,
			RoundCount: int64(r.RoundCount),
		}
		scores = total.TotalScore
		r.RoomBase.RoomRecord += "结算:" + r.Seats[maxScoreIndex].User.NickName + " " + total.TotalMsg[maxScoreIndex] + "\r\n"
	}

	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Total, ackTotal)
	r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Total, ackTotal)
	r.RoundResult = ackTotal
	r.InsertRoomRecord(scores)

	//房间结束
	if !flag {

		r.StopTicker = true

		list := r.GetSummaryList()
		r.BCMessage(mjgame.MsgID_MSG_NOTIFY_SUMMARY, &list)
		r.ClearRoomUserRoomId()
		r.Mux.Lock()
		rb.ChanRoom <- r.RoomId //销毁房间
		r.Mux.Unlock()
		return
	}

	if r.IsDraw {
		r.IsContinueBanker = true
	}

	//	if r.IsDraw { //  象山麻将如果庄家有明杠或暗杠，则庄家轮到下一家，反之则继续坐庄
	//		if r.Seats[r.BankerIndex].Cards.Kong.Count > 0 {
	//			r.IsContinueBanker = false
	//		} else {
	//			r.IsContinueBanker = true
	//		}
	//	}

	if !r.IsContinueBanker { // 如果庄胡 或者 流局 则连庄
		r.CalcDirection()
	}
}

// 获得开始信息
func (r *RoomXiangshan) GetStartInfo() (int, int, int) {
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	rollIndex := 0
	bankerIndex := 0
	points := 0

	if r.RoundCount == 0 { // fix 只有第一局需要掷骰子
		points = util.RandInt(2, 12)
		rollIndex = seed.Intn(r.Rules.SeatLen)
		bankerIndex = seed.Intn(r.Rules.SeatLen)
	}

	absValue := int(math.Abs(float64(rollIndex - bankerIndex)))

	if util.GetIndex(SelfPoints, points) >= 0 {
		absValue = 0
	}
	switch absValue {
	case 0:
		bankerIndex = rollIndex
		index := rand.Intn(len(SelfPoints))
		points = SelfPoints[index]
	case 1:
		if rollIndex > bankerIndex {
			index := rand.Intn(len(BackPoints))
			points = BackPoints[index]
		} else {
			index := rand.Intn(len(FrontPoints))
			points = FrontPoints[index]
		}
	case 2:
		index := rand.Intn(len(FacePoints))
		points = FacePoints[index]
	case 3:
		if rollIndex > bankerIndex {
			index := rand.Intn(len(FrontPoints))
			points = FrontPoints[index]
		} else {
			index := rand.Intn(len(BackPoints))
			points = BackPoints[index]
		}
	default:
	}

	return rollIndex, bankerIndex, points
}

// 改变用户座位
func (r *RoomXiangshan) ChangeUsersPosition() []*mjgame.ACK_User_Info {
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	slices := seed.Perm(r.Rules.SeatLen)

	var cloneUsers = make([]user.User, 0)
	for _, v := range r.Seats {
		cloneUsers = append(cloneUsers, *v.User)
	}
	var users = make([]*mjgame.ACK_User_Info, 0)
	for i, _ := range r.Seats {
		index := slices[i]
		//r.Seats[index].Index = index

		r.Seats[i].UID = strconv.Itoa(cloneUsers[index].User.ID)
		r.Seats[i].User = &cloneUsers[index]

		//users = append(users, common.BuildSeatBaseToAckUserInfo(r.Seats[index]))
		users = append(users, common.BuildSeatBaseToAckUserInfo(r.Seats[i]))
	}
	return users
}

// 算番
func (r *RoomXiangshan) FanCalc(index int, cardid int) st.TotalResult {
	arrMj := r.Seats[index].Cards.List
	mjs := make([]int, 42)
	for i := 0; i < arrMj.Count; i++ {
		if *arrMj.Index(i) != nil {
			if *arrMj.Index(i) != nil {
				c := (*arrMj.Index(i)).(*rb.Card)
				mjs[c.ID]++
			}
		}
	}
	mjs[cardid]++

	arg := make([]interface{}, 0)
	(r.StlCtrl).(*XiangShan_Statement).BaseCtl.SiChuan = false
	return (r.StlCtrl).(*XiangShan_Statement).FanCalc(index, arg)
}

//检查是否胡牌
func (r *RoomXiangshan) CheckHu(index int, cardId int) int {
	arrMj := r.Seats[index].Cards.List

	mjs := make([]int, 42)
	for i := 0; i < arrMj.Count; i++ {
		if *arrMj.Index(i) != nil {
			c := (*arrMj.Index(i)).(*rb.Card)
			mjs[c.ID]++
		}
	}
	if cardId >= 0 { // fix 自摸判断时候, list已经包含这张牌了.  自摸时候传-1
		mjs[cardId]++
	}
	result := (r.StlCtrl).(*XiangShan_Statement).CheckHu(mjs)
	if result > 0 {
		isWin := (r.StlCtrl).(*XiangShan_Statement).CheckCanWin(index)
		if !isWin {
			result = 0
		}
	}

	return result
}

//添加操作
func (r *RoomXiangshan) AddTool(toolType int, index int, tindex int, val []int) {
	(r.StlCtrl).(*XiangShan_Statement).AddTool(toolType, index, tindex, val)
}

//
func (r *RoomXiangshan) AddListCard(index int, listcard *al.ArrayList) {

	arr := make([]int, listcard.Count)
	for i := 0; i < listcard.Count; i++ {
		if *listcard.Index(i) != nil {
			arr[i] = (*listcard.Index(i)).(*rb.Card).ID
		}
	}

	(r.StlCtrl).(st.IStatement).AddTool(
		st.T_Deal,
		index,
		-1,
		arr,
	)
}

//判断是否可以吃牌
func (r *RoomXiangshan) CheckCanChow(index int, tIndex int, cards []*rb.Card, card *rb.Card) bool {
	if cards[0] == nil || cards[1] == nil || cards[2] == nil {
		return false
	}
	ccount := r.Seats[index].Cards.List.Count
	if ccount < 2 || ccount%3 == 0 || ccount%3 == 2 {
		return false
	}

	//	if !r.RoomBase.CheckCanChow(index, []int{cards[0].ID, cards[1].ID, cards[2].ID}, card) {
	//		return false
	//	}

	if len(r.Seats[index].ChowCardIDs) > 0 && r.ChengBao[index].Seat[tIndex] < 3 {
		for _, v := range r.Seats[index].ChowCardIDs {
			if r.CurCard.ID == v {
				fmt.Println("不能吃 ChowCardIDs ", r.CurCard)
				return false
			}
		}
	}

	i1, i2, i3 := float64(cards[0].Num), float64(cards[1].Num), float64(cards[2].Num)
	if i1 < 0 || i2 < 0 || i3 < 0 {
		return false
	}
	if cards[0].Type != cards[1].Type || cards[1].Type != cards[2].Type {
		return false
	}
	if (math.Abs(i2-i1) == math.Abs(i3-i1)) ||
		(math.Abs(i1-i2) == math.Abs(i3-i2)) ||
		(math.Abs(i1-i3) == math.Abs(i2-i3)) {
		return true
	}

	return false
}

//判断是否可以碰牌
func (r *RoomXiangshan) CheckPeng(pCard *rb.Card) {
	index := r.CurIndex
	for i := 0; i < 3; i++ {
		index++ //从下家开始判断
		index = index % 4

		if r.Seats[index].User == nil {
			continue
		}
		cardCount := r.GetUserCardCount(index, pCard.ID)
		if cardCount >= 2 {
			var flag bool //过手碰或者过手胡都不能碰
			for _, v := range r.Seats[index].PengCardIDs {
				if v == pCard.ID {
					flag = true
					break
				}
			}

			if flag {
				tip := &mjgame.Tip{
					Tip: "当前玩家处于过手碰",
				}
				r.Seats[index].User.SendMessage(mjgame.MsgID_MSG_NOTIFY_TIP, tip)
				continue
			}
			r.AddToolUser(index, 0, 0, 1, 0, 0, 1)
			r.RoomBase.RoomRecord += "检测(" + r.Seats[index].User.NickName + ") 碰" + pCard.MSG + "\r\n"
			r.Seats[index].IsCanPeng = true
			r.Seats[index].PengCardIDs = append(r.Seats[index].PengCardIDs, pCard.ID)

		}
	}
}

//通知所有有操作玩家
func (r *RoomXiangshan) NotifyTool() {
	r.RoomBase.NotifyTool()
}

// 转到下一个出牌的玩家
// get = 摸牌
// forward = 从后面摸牌
// pass = 是否过掉本次操作
func (r *RoomXiangshan) TurnNextPlayer(get bool, forward bool, pass bool) {
	r.CurIndex++
	r.CurIndex = r.CurIndex % 4

	if !r.IsRun {
		return
	}

	r.WaitOptTool.ClearAll()
	r.WaitOptTool.IsSelf = true
	r.Seats[r.CurIndex].IsCanWin = false
	r.Seats[r.CurIndex].IsCanKong = false
	r.Seats[r.CurIndex].IsCanPeng = false
	r.Seats[r.CurIndex].IsCanChow = false

	ack := mjgame.ACKBC_CurPlayer{Seat: int32(r.Seats[r.CurIndex].Index), Type: int32(rb.WaitPut), RoundTime: int32(r.WaitPutTimeOut)}
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_CurPlayer, &ack) // 广播当前出牌的玩家

	haveHua := false

	if !r.Seats[r.CurIndex].IsTransfer {
		if r.CurCard != nil {
			haveHua = r.CheckHaveHuaByIdx(r.CurIndex)
		}

		r.BoardCastTransferHua()
		r.Seats[r.CurIndex].IsTransfer = true
	}

	canHu := 0
	kongType := 0
	if get && r.Seats[r.CurIndex].Cards.List.Count%3 == 2 { //如果再摸就是相公牌了
		get = false
	}

	//如果换花以后可以杠和胡，就需要这段代码
	if get == true && haveHua == true && r.Seats[r.CurIndex].Cards.List.Count%3 == 2 && r.CurCard != nil {
		//fmt.Println("我有13张牌，哈哈哈")
		//这个时候玩家是可以胡的，追加代码
		canHu = r.CheckHu(r.CurIndex, -1)
		if canHu > 0 {
			r.AddToolUser(r.CurIndex, 1, 0, 0, 0, 1, 1)
			//fmt.Println("添加了胡操作 。。。。")
		}
		//判断是否可以胡
		r.CheckCanKong(r.CurIndex, -1, true)

	} else if haveHua == true && get == true && r.CurCard != nil { //有花并且少一张牌，这家伙需要摸牌

		if r.CheckHu(r.CurIndex, r.CurCard.ID) > 0 {
			r.AddToolUser(r.CurIndex, 1, 0, 0, 0, 1, 1)
		}
		r.CheckChowAfterHuanhua(r.CurIndex, r.CurCard)

		cardNum := r.GetUserCardCount(r.CurIndex, r.CurCard.ID)

		if cardNum >= 2 { //可以碰
			r.AddToolUser(r.CurIndex, 0, 0, 1, 0, 1, 1)
		}

		if cardNum >= 3 {
			r.AddToolUser(r.CurIndex, 0, 1, 0, 0, 1, 1)
		}

		if r.WaitOptTool.Count() > 0 { //当前玩家换完花，对上家出的牌进行判断有操作
			r.CurIndex = (r.CurIndex + 3) % r.Rules.SeatLen
			r.WaitOptTool.IsSelf = false
			get = false
		}
	}

	if get {
		// 摸牌
		var ackbc_GetCard mjgame.ACKBC_GetCard
		r.CurCard, ackbc_GetCard = r.GetNextCard(forward) // 摸牌并发送数据
		r.Show = true
		r.Seats[r.CurIndex].HuCardIDs = []int{} // 过手胡清除

		if r.CurCard != nil {

			//fmt.Println("GetCard----------------------------------------------> ", r.Seats[r.CurIndex].User.NickName, r.CurIndex, r.Seats[r.CurIndex].Cards.List.Count)
			if r.Seats[r.CurIndex].Cards.List.Count != 13 && r.Seats[r.CurIndex].Cards.List.Count%3 == 1 {
				fmt.Println("手牌数量错误. ", r.CurIndex, r.Seats[r.CurIndex].User.NickName, r.CurCard)
			}
			//			strHu := ""
			//			strKong := ""
			canHu = r.CheckHu(r.CurIndex, -1) //检查是否能胡(天胡)
			if !pass {
				_, kongType = r.CheckCanKong(r.CurIndex, r.CurCard.ID, true) //检查是否可以杠牌
			}
			if canHu > 0 {
				r.AddToolUser(r.CurIndex, 1, 0, 0, 0, 1, 1)
				//				strHu = "胡"
			}

			if kongType <= 0 {
				kongType = 0
			} else {
				//				strHu = "杠"
			}
			//r.RoomBase.RoomRecord += "转到(" + r.Seats[r.CurIndex].User.NickName + ") 摸牌:" + r.CurCard.MSG + " " + strHu + strKong + "\r\n"
			ackbc_GetCard.Tool[0] = int32(canHu)
			ackbc_GetCard.Tool[1] = int32(kongType)
			r.RoomBase.MToolChecker.SetCptTool(r.CurIndex, int(mjgame.MsgID_MSG_ACKBC_GetCard), []int{r.CurCard.ID}, r.Seats[r.CurIndex].User.NickName)
			r.RoomBase.MToolChecker.ShowAllTools()
			for _, v := range r.Seats {
				if v.Index == r.CurIndex {
					ackbc_GetCard.Cid = int32(r.CurCard.ID)
				} else {
					ackbc_GetCard.Cid = -1
					ackbc_GetCard.Tool[0] = 0
					ackbc_GetCard.Tool[1] = 0
				}
				v.User.SendMessage(mjgame.MsgID_MSG_ACKBC_GetCard, &ackbc_GetCard)
			}

			rec := [5]interface{}{ackbc_GetCard.Index, r.CurCard.ID, ackbc_GetCard.FromLast, canHu, kongType}

			r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_GetCard, &rec)

		} else {
			// 没有牌摸的情况下,  流局
			//if r.CheckDraw() {
			r.Draw()
			//}
		}
	} else {
		// 不摸牌
		//fmt.Println("NoGetCard----------------------------------------------> ", r.Seats[r.CurIndex].User.NickName, r.CurIndex, r.Seats[r.CurIndex].Cards.List.Count)

		r.RoomBase.MToolChecker.SetAllUserTool(-1)
		lastopt := r.WaitOptTool.GetOpt(r.CurIndex)

		canKong := 0
		strKong := ""
		if lastopt != nil && !pass {
			_, kongType = r.CheckCanKong(r.CurIndex, -1, true)
			if kongType > 0 {
				canKong = 1
				strKong = "杠"
			}
		}

		r.RoomBase.RoomRecord += "转到(" + r.Seats[r.CurIndex].User.NickName + ") 不摸牌 " + strKong + "\r\n"
		r.RoomBase.MToolChecker.SetTools(r.CurIndex, []int{-1, canKong - 1, -1, -1, 0, -1, -1, -1}) // 胡0 杠1 碰2 吃3 出4 过5 摸6 等7
		r.RoomBase.MToolChecker.ShowAllTools()
	}

	if r.IsRun {
		//		fmt.Println("duludulu！！！！！！！")
		r.CurToolIndex = r.CurIndex
		r.WaitTimeCount = r.WaitToolTimeOut
		if r.WaitOptTool.Count() > 0 {
			r.NotifyTool()
		} else {
			//r.State = rb.WaitPut
			//@andy0920
			//r.NewProcess()
			r.WaitPut(r.WaitPutTimeOut)
		}

	}

}

//返回玩家手里是否有花
func (r *RoomXiangshan) CheckHaveHuaByIdx(tIdx int) bool {
	seat := r.Seats[tIdx]
	if seat == nil {
		return false
	}
	listLength := seat.Cards.List.Count

	haveHua := false
	for i := 0; i < listLength; i++ {
		if *seat.Cards.List.Index(i) != nil {
			card := (*seat.Cards.List.Index(i)).(*rb.Card)
			if card.Type == rb.H {
				haveHua = true
				break
			}
		}
	}
	return haveHua
}

//等待操作,自摸判断操作
func (r *RoomXiangshan) WaitPut(timeout int) {

}

//开始等待操作
func (r *RoomXiangshan) StartWaitTool(card *rb.Card) {
	r.WaitOptTool.ClearAll()
	r.WaitOptTool.IsSelf = false

	r.CheckWin(card.ID)
	r.CheckCanKong(r.CurIndex, card.ID, false)
	r.CheckPeng(card)
	r.CheckChow(card)

	//r.RoomBase.RoomRecord += "检查操作(" + r.Seats[r.CurIndex].User.NickName + ") " + strconv.Itoa(r.WaitOptTool.Count()) + "\r\n"

	r.RoomBase.UpdateToolChecker()

	if r.WaitOptTool.Count() > 0 { //等待操作
		r.CurToolIndex = 0 // 从0 开始
		r.WaitTimeCount = r.WaitToolTimeOut
		r.NotifyTool()
		r.WaitPutTool()

		r.CheckNextHuaPass()

	} else {
		//是否流局
		if r.CheckDraw() {
			r.Draw()

		} else {
			//没有要操作的
			r.WaitTimeCount = 0
			r.CurToolIndex = -1
			r.Status = rb.WaitPut
			r.TurnNextPlayer(true, true, false)
		}
	}
}

//下家有花就点过
func (r *RoomXiangshan) CheckNextHuaPass() bool {
	tIdx := (r.CurIndex + 1) % r.Rules.SeatLen
	seat := r.Seats[tIdx]
	if seat == nil {
		return false
	}
	listLength := seat.Cards.List.Count

	haveHua := false
	for i := 0; i < listLength; i++ {
		if *seat.Cards.List.Index(i) != nil {
			card := (*seat.Cards.List.Index(i)).(*rb.Card)
			if card.Type == rb.H {
				haveHua = true
				break
			}
		}
	}

	if haveHua == true { //如果有花就点过
		r.Pass(seat.User)
	}

	return haveHua
}

//检查是否可以胡
func (r *RoomXiangshan) CheckWin(cid int) {
	index := r.CurIndex

	for i := 0; i < 3; i++ {
		index++
		index = index % 4
		r.Seats[index].IsCanWin = false //复位
		result := r.CheckHu(index, cid)
		if result > 0 {
			var flag bool
			if len(r.Seats[index].HuCardIDs) > 0 {
				flag = true
			}

			if flag {
				fmt.Println("当前玩家处于过手胡 ")
				tip := &mjgame.Tip{
					Tip: "当前玩家处于过手胡",
				}
				r.Seats[index].User.SendMessage(mjgame.MsgID_MSG_NOTIFY_TIP, tip)
				continue
			}

			r.Seats[index].HuCardIDs = append(r.Seats[index].HuCardIDs, cid)
			r.AddToolUser(index, 1, 0, 0, 0, 0, 1)
			r.Seats[index].IsCanWin = true
			r.RoomBase.RoomRecord += "检测(" + r.Seats[index].User.NickName + ") 胡" + st.GetMjNameForIndex(cid) + "\r\n"
		}
	}
}

//是否流局
func (r *RoomXiangshan) CheckDraw() bool {

	result := r.AllCardLength - r.EndBlank - r.CurMJIndex - def.XiangShanDrawCount
	if result <= 0 {
		r.AddTool(st.T_Draw, -1, -1, []int{})
		return true
	}
	return false
}

// 流局
func (r *RoomXiangshan) Draw() {

	r.IsDraw = true
	r.RoundToatlFinish = false
	r.IsRun = false
	r.RoomBase.RoomRecord += "流局\r\n"
	allSeatCards := r.GetAllSeatCards()
	ackDraw := &mjgame.ACKBC_Draw{
		RoomId: int32(r.RoomBase.RoomId),
		Cards:  allSeatCards,
	}
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Draw, ackDraw)

	recordDraw := &mjgame.ACKBC_Draw{
		Cards: r.GetRecordAllSeatCards(),
	}
	r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Draw, recordDraw)

	r.RoundTotal()
	r.RoundToatlFinish = true

}

//计算风圈
func (r *RoomXiangshan) CalcDirection() {
	var flag bool
	for _, v := range r.Bankers {
		if v == r.BankerIndex {
			flag = true
			break
		}
	}

	if !flag {
		r.Bankers = append(r.Bankers, r.BankerIndex)
	}

	if len(r.Bankers) == r.Rules.SeatLen {
		r.FengQuan++
		r.FengQuan = r.FengQuan % 4
		//清空圈风
		r.Bankers = []int{}
	}
	return
}

type ByAccumulation []*rb.SeatBase

func (s ByAccumulation) Len() int      { return len(s) }
func (s ByAccumulation) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByAccumulation) Less(i, j int) bool {
	accumulationOne, accumulationTwo := s[i].Accumulation, s[j].Accumulation
	if accumulationOne == nil || accumulationTwo == nil {
		return false
	}
	if accumulationOne.Score > accumulationTwo.Score {
		return true
	} else if accumulationOne.Score == accumulationTwo.Score {
		if accumulationOne.WinCount > accumulationTwo.WinCount {
			return true
		} else if accumulationOne.WinCount == accumulationTwo.WinCount {
			if accumulationOne.FireCount < accumulationTwo.FireCount {
				return true
			}
		}
	}
	return false
}

//统计
func (r *RoomXiangshan) GetSummaryList() mjgame.TotalSummary {
	var totalSummary mjgame.TotalSummary

	/*var sortSeats = r.Seats
	sort.Sort(ByAccumulation(sortSeats))
	for i, v := range sortSeats {
		summary := &mjgame.Summary{
			Id:         strconv.Itoa(v.User.ID),
			Name:       v.User.NickName,
			Icon:       v.User.Icon,
			RoundCount: int32(r.RoundCount),
			Rank:       int32(i + 1),
			FireCount:  v.Accumulation.FireCount,
			WinCount:   v.Accumulation.WinCount,
			PayCount:   v.Accumulation.PayCount,
			Score:      v.Accumulation.Score,
		}
		totalSummary.List = append(totalSummary.List, summary)
	}*/
	return totalSummary
}

//是否可以解散 返回: 0未完成投票,  1解散成功    2解散失败
func (r *RoomXiangshan) IsDisbanding() int {
	var agreeCount int
	var noagreeCount int
	var result = 0
	for _, v := range r.Votes {
		if v == Agree {
			agreeCount++
		} else if v == 2 {
			noagreeCount++
		}
	}

	if agreeCount > r.Rules.SeatLen/2 {
		result = 1
	} else if noagreeCount >= r.Rules.SeatLen/2 {
		result = 2
	}
	return result
}

// 销毁房间
func (r *RoomXiangshan) DestoryRoom() {

	r.StopTicker = true
	//发送流局消息
	allSeatCards := r.GetAllSeatCards()
	ackDraw := &mjgame.ACKBC_Draw{
		RoomId: int32(r.RoomBase.RoomId),
		Cards:  allSeatCards,
	}
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Draw, ackDraw)

	recordDraw := &mjgame.ACKBC_Draw{
		Cards: r.GetRecordAllSeatCards(),
	}
	r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Draw, recordDraw)

	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Total, &mjgame.ACKBC_Total{
		Finished:   true,
		RoundCount: int64(r.RoundCount),
	})

	r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Total, &mjgame.ACKBC_Total{
		Finished:   true,
		RoundCount: int64(r.RoundCount),
	})

	//插入流局记录
	if r.RoundCount == 0 {
		room := &model.Room{
			Type:         int(mjgame.MsgID_GTYPE_ZheJiang_XiangShan),
			UserID:       r.CreateUserId,
			Rules:        util.IntArrToString(r.Rules.Rules),
			ServerRoomID: r.RoomId,
			UniqueCode:   r.UniqueCode,
		}
		err := model.GetRoomModel().Create(room)
		if err == nil {
			r.DbRoomId = room.ID
		}
	}
	var scores = make([]int32, 4)
	r.InsertRoomRecord(scores)

	//发送大结算
	list := r.GetSummaryList()
	r.BCMessage(mjgame.MsgID_MSG_NOTIFY_SUMMARY, &list)
	r.Votes = []int{0, 0, 0, 0}
	r.VoteStarter = -1
	for _, v := range r.Seats {
		if v.User != nil {
			v.User.RoomId = 0
			v.User.GameType = nil
		}
		v.Message = nil
	}
	for _, v := range r.WatchSeats {
		v.User.RoomId = 0
		v.GameType = nil
	}

	r.Mux.Lock()
	rb.ChanRoom <- r.RoomId //销毁房间
	r.Mux.Unlock()
}

// 插入房间记录
func (r *RoomXiangshan) InsertRoomRecord(scores []int32) {

	r.RoomBase.Try(func() {
		result := make(model.IntKv)
		for k, v := range scores {
			var userId int
			for _, user := range r.KickUsers {
				if k == user.Position {
					userId = user.UserID
					break
				}
			}
			if userId == 0 {
				if r.Seats[k].User != nil {
					userId = r.Seats[k].User.ID
				}
			}
			if _, ok := result[userId]; !ok {
				result[userId] = int(v)
			}
		}

		m := model.BeginCommit()

		mon := time.Now().Month()
		day := time.Now().Day()
		h := time.Now().Hour()
		min := time.Now().Minute()
		sec := time.Now().Second()

		tempStr := strconv.Itoa(int(mon)) + strconv.Itoa(day) + strconv.Itoa(h) + strconv.Itoa(min) + strconv.Itoa(sec) + strconv.Itoa(r.RoomId) + strconv.Itoa(r.RoundCount)

		battleRecord := &model.BattleRecord{
			RoomID:     int(r.DbRoomId),
			Round:      r.RoundCount,
			Result:     result,
			ReviewCode: tempStr,
		}

		if err := m.Create(battleRecord).Error; err != nil {
			m.Rollback()
			return
		}

		for k, v := range result {
			roomRecord := &model.RoomRecord{
				RoomID:         int(r.DbRoomId),
				UserID:         k,
				BattleRecordID: int(battleRecord.ID),
				RoomType:       int(mjgame.MsgID_GTYPE_ZheJiang_XiangShan),
			}
			if v > 0 {
				roomRecord.Win = model.Win
			}
			if err := m.Create(roomRecord).Error; err != nil {
				m.Rollback()
				return
			}
		}

		m.Commit()
		//以上是原来的
		return
		m = model.BeginCommit()
		replayRecord := &model.ReplayRecord{
			ReviewCode: tempStr,
		}

		PlayBack, e := json.Marshal(r.BattleRecord)
		if e == nil {
			replayRecord.PlayBack = string(PlayBack)
			//fmt.Println("json:: shuju :: " + battleRecord.PlayBack)
		}

		if err := m.Create(replayRecord).Error; err != nil {
			m.Rollback()
			return
		}

		m.Commit()

	}, func(e interface{}) {
		fmt.Println("xiangshan  InsertRoomRecord ", e)
	})
}

//进入房间
func (r *RoomXiangshan) IntoRoom(user *user.User) {

	fmt.Println("user state::" + strconv.Itoa(user.State))

	var isWatch bool
	user.RoomId = r.RoomId

	index := r.GetSeatIndexById(user.ID)
	if index < 0 {
		wIndex := r.GetWatchSeat(user.ID)
		if wIndex < 0 {
			isWatch = true
			r.WatchSeats = append(r.WatchSeats, user)
		} else {
			r.WatchSeats[wIndex] = user
		}
	} else {
		r.Seats[index].UID = strconv.Itoa(user.ID)
		r.Seats[index].User = user
	}

	ack := &mjgame.ACKBC_Into_Room{
		Name:    user.NickName,
		Uid:     strconv.Itoa(user.ID),
		RoomId:  int32(user.RoomId),
		Ip:      user.GetIP(),
		Index:   -1,
		Icon:    user.Icon,
		Coin:    int32(user.Coin),
		Type:    int32(mjgame.MsgID_GTYPE_ZheJiang_XiangShan),
		Diamond: int32(user.Diamond),
		Level:   0,
		Robot:   int32(user.IsRobot),
		GPS_LNG: user.GPS_LNG,
		GPS_LAT: user.GPS_LAT,
		Rule:    r.Rules.Rules,
	}

	roomInfo := r.GetRoomInfo() //房间信息
	user.SendMessage(mjgame.MsgID_MSG_ACK_RoomInfo, roomInfo)

	if isWatch {
		r.BCMessage(mjgame.MsgID_MSG_ACKBC_Into_Room, ack)
	} else {
		user.SendMessage(mjgame.MsgID_MSG_ACKBC_Into_Room, ack)
	}

	user.SendMessage(mjgame.MsgID_MSG_ACK_Room_User, r.GetRoomUser())

	if r.IsRun {
		r.SendGameInfo(user, false)
		r.SendSeatCard(user.ID)

		if r.VoteStarter >= 0 { //解散状态
			r.DisbandRoom(user, nil)
		}

	} else {
		if index >= 0 {
			if r.Seats[index].State == int(mjgame.StateID_GameState_Total) {
				r.SendGameInfo(user, false)
				r.SendSeatCard(user.ID)
			}
		}
	}
}

//得到房间信息
func (r *RoomXiangshan) GetRoomInfo() *mjgame.ACK_Room_Info {
	ack := mjgame.ACK_Room_Info{
		RoomId:     int32(r.RoomId),
		Type:       int32(mjgame.MsgID_GTYPE_ZheJiang_XiangShan),
		City:       int32(0),
		Level:      int32(0),
		Rules:      r.Rules.Rules,
		SeatCount:  int32(len(r.Seats)),
		Starting:   r.IsRun,
		RoundCount: int32(r.RoundCount),
		Direction:  int32(r.FengQuan),
		UniqueCode: r.UniqueCode,
	}

	ack.DisbandLeftTime = int64(def.WaitStartTime - (time.Now().Unix() - r.CreateTime))
	if ack.DisbandLeftTime < 0 {
		ack.DisbandLeftTime = 0
	}

	if r.IsRun || r.RoundCount > 0 {
		ack.LeftTime = int64((r.Rules.MaxTime * 60) - (int(time.Now().Unix()) - r.StartTime))
		if ack.LeftTime < 0 {
			ack.LeftTime = 0
		}
	}

	startUser, _ := r.GetFirstSitSeatInfo()
	if startUser != nil {
		ack.NickName = startUser.NickName
	}

	return &ack
}

func (r *RoomXiangshan) GetRecordRoomInfo() interface{} {
	ack := &struct {
		RID        int32
		Type       int32
		Rule       []int32
		RoundCount int32
		direction  int32
		leftTime   int64
	}{
		RID:        int32(r.RoomId),
		Type:       int32(mjgame.MsgID_GTYPE_ZheJiang_XiangShan),
		Rule:       r.Rules.Rules,
		RoundCount: int32(r.RoundCount),
		direction:  int32(r.FengQuan),
	}

	if r.IsRun || r.RoundCount > 0 {
		ack.leftTime = int64((r.Rules.MaxTime * 60) - (int(time.Now().Unix()) - r.StartTime))
		if ack.leftTime < 0 {
			ack.leftTime = 0
		}
	}

	return &ack
}

// 等待用户操作
// 判断是否都操作完毕, 或者最高优先级的操作
func (r *RoomXiangshan) WaitPutTool() {
	if !r.IsRun {
		fmt.Println("fangjianjieshule caozuoshibai .....")
		return
	}
	huList, index, opType, ok := r.WaitOptTool.CheckGetCpt()
	if !ok {
		//fmt.Println("qitawanjiayoucaozuo caozuoshibai .....")
		return
	}

	if index >= 0 {
		r.RoomBase.RoomRecord += "判断结果(" + r.Seats[index].User.NickName + ") " + strconv.Itoa(opType) + "\r\n"
	}

	var u *user.User
	var toolUser *rb.NeedWait
	var winUsers []*user.User
	if index >= 0 {
		u = r.Seats[index].User
		toolUser = r.WaitOptTool.GetOpt(index)
		if toolUser == nil {
			fmt.Println("zhaobudaowanjia caozuoshibai .....")
			return
		}

		if opType == 0 && len(huList) > 0 { // 胡牌 一炮多响
			for i := 0; i < len(huList); i++ {
				user := r.Seats[huList[i]].User
				r.RoomBase.MToolChecker.SetTool(r.GetSeatIndexById(user.ID), 0, 0)
				winUsers = append(winUsers, user)
			}
		}
	}
	// 胡0 杠1 碰2 吃3 出4 过5 摸6 等7
	switch opType {
	case rb.Chow:
		chowArgs := &mjgame.Chow{
			Cid1: int32(toolUser.Param[0]),
			Cid2: int32(toolUser.Param[1]),
			Cid3: int32(toolUser.Param[2]),
		}
		r.RoomBase.MToolChecker.SetTool(r.GetSeatIndexById(u.ID), 3, 0)
		r.ChowCard(u, chowArgs)
	case rb.Peng:
		r.RoomBase.MToolChecker.SetTool(r.GetSeatIndexById(u.ID), 2, 0)
		r.PengCard(u, toolUser.Param[0])
	case rb.Kong:
		r.RoomBase.MToolChecker.SetTool(r.GetSeatIndexById(u.ID), 1, 0)
		r.KongCard(u, toolUser.Param[0])
	case rb.Hu:
		r.RoomBase.MToolChecker.SetTool(r.GetSeatIndexById(u.ID), 0, 0)
		r.WinCard(winUsers, toolUser.Param[0])
	case rb.Pass:
		getCard := true           // 摸牌
		if r.WaitOptTool.IsSelf { // 过了自己的暗杠不摸牌
			if r.CurIndex <= 0 {
				r.CurIndex = r.CurIndex + 4 - 1
			} else {
				r.CurIndex--
			}

			getCard = false
		}
		if r.IsKongHu { //处理拉杠胡
			index := r.CurIndex
			r.Show = false
			if index <= 0 {
				r.CurIndex = index - 1 + 4
			} else {
				r.CurIndex = index - 1
			}
			r.Status = rb.WaitPut
			//r.TurnNextPlayer(getCard, false, getCard)
			r.Seats[index].Message = nil
			r.IsKongHu = false
		}
		r.TurnNextPlayer(getCard, true, false)
	}
}

//获得多人胡的信息
func (r *RoomXiangshan) GetMultiWinInfo() ([]int, int) {
	//四人麻将最多三个人胡
	var huIndexes []int
	var huCardId = -1

	for i := 0; i < 3; i++ {
		lastTool := (r.StlCtrl).(*XiangShan_Statement).Get(i)
		if lastTool != nil && lastTool.Tool.ToolType == st.T_Hu {
			huIndex := lastTool.Tool.Index
			huCardId = lastTool.Tool.Val[0]
			if huIndex < 0 || huCardId < 0 {
				return []int{}, huCardId
			}
			huIndexes = append(huIndexes, huIndex)
		}
	}

	return huIndexes, huCardId
}

//得到最大索引
func (r *RoomXiangshan) GetMaxIndex(arr []int32) int {
	var maxIndex int
	var score int32

	for k, v := range arr {
		if v > score {
			score = v
			maxIndex = k
		}
	}

	return maxIndex
}

// 当前局可结束
func (r *RoomXiangshan) RoundCanFinish() bool {
	var count int

	for _, v := range r.Seats {
		if v == nil || v.User == nil {
			count++
			continue
		} else if v.User.State == def.Offline {
			if v.OfflineTime.Add(def.MaxOfflineTime * time.Minute).Before(time.Now()) {
				count++
				break
			}
		}
	}

	if count > 0 {
		return true
	}

	return false
}

// 得到踢人Index
func (r *RoomXiangshan) GetKickIndex() []int32 {
	var indexs []int32
	for k, v := range r.Seats {
		if v.User != nil {
			if v.User.State == def.Offline {
				if v.OfflineTime.Add(def.KickTimeDuration*time.Second).Unix() == time.Now().Unix() {
					indexs = append(indexs, int32(k))
				}
			}
		}
	}

	return indexs
}

// 杠牌操作
func (r *RoomXiangshan) MoveKongList(uIndex int, tIndex int, cid, kongType int) {
	seat := r.Seats[uIndex]
	t, n := st.GetMjTypeNum(cid)

	// 移动第一张牌
	if kongType == 1 { //明杠
		card := []*rb.Card{&rb.Card{ID: cid, TIndex: tIndex, Status: kongType - 1, Type: t, Num: n}}
		r.MoveToList(r.Seats[tIndex].Cards.Out, card, seat.Cards.Kong)
	} else if kongType == 2 || kongType == 3 { //暗杠, 补杠
		card := []*rb.Card{&rb.Card{ID: cid, TIndex: tIndex, Status: kongType - 1, Type: t, Num: n}}
		r.MoveToList(seat.Cards.List, card, seat.Cards.Kong)
	}

	// 移动3张牌
	for i := 0; i < 3; i++ {
		if kongType == 1 || kongType == 2 { //明杠,移动list牌
			card := []*rb.Card{&rb.Card{ID: cid, Status: kongType - 1, TIndex: tIndex, Type: t, Num: n}}
			r.MoveToList(seat.Cards.List, card, seat.Cards.Kong)
		} else if kongType == 3 { //判断是否碰上杠
			card := []*rb.Card{&rb.Card{ID: cid, TIndex: tIndex, Status: kongType - 1, Type: t, Num: n}}
			r.MoveToList(seat.Cards.Peng, card, seat.Cards.Kong)
		}
	}

}

// 吃牌操作
func (r *RoomXiangshan) MoveChowList(index int, cards []*rb.Card) {
	seat := r.Seats[index]
	r.MoveToList(r.Seats[r.CurIndex].Cards.Out, []*rb.Card{cards[0]}, seat.Cards.Chow)
	r.MoveToList(seat.Cards.List, []*rb.Card{cards[1], cards[2]}, seat.Cards.Chow)
}

//
func (r *RoomXiangshan) DealKongHu(cardID int) {

	r.CheckWin(cardID)

	if r.WaitOptTool.Count() > 0 {
		r.CurToolIndex = 0 // 从0 开始
		r.Status = rb.WaitTool
		r.WaitTimeCount = r.WaitToolTimeOut
		r.NotifyTool()
		r.IsKongHu = true
		r.KongHuCardID = cardID
	}
}

//清除房间用户ID
func (r *RoomXiangshan) ClearRoomUserRoomId() {
	for _, v := range r.Seats {
		if v.User != nil {
			v.User.RoomId = 0
			v.User.GameType = nil
		}
	}
	for _, v := range r.WatchSeats {
		v.User.RoomId = 0
		v.GameType = nil
	}
}

//设置玩家手中牌 用于测试环境
func (r *RoomXiangshan) SetInitCards(uid string, cids []string) {
	for _, v := range r.Seats {
		if v.UID == uid && r.CurMJIndex <= 61 && r.CurMJIndex > 0 {
			v.Cards.List = al.New()
			for k, cid := range cids {
				CardId, err := strconv.Atoi(cid)
				if err != nil {
					fmt.Println("*err is ", err)
					continue
				}
				if CardId >= 0 {
					v.Cards.List.Add(rb.GetCardById(CardId))
				}
				if k == 13 {
					r.CurCard = rb.GetCardById(CardId)
					r.WaitPut(r.WaitPutTimeOut)
				}
			}
		}
	}
}

//得到房间用户信息
func (r *RoomXiangshan) GetRoomUser() *mjgame.ACK_Room_User {
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
					if v.OfflineTime.Add(3 * time.Minute).Before(time.Now()) {
						user.CanKick = true
					}

					user.OfflineTime = int32(v.OfflineTime.Add(3*time.Minute).Unix() - time.Now().Unix())
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
			for _, kickUser := range r.KickUsers {
				if kickUser.UserID == v.User.ID {
					user.Index = int32(kickUser.Position)
					//					fmt.Println("user.Index::" + strconv.Itoa(int(user.Index)))
					if user.State == def.Online {
						user.State = def.Kick
					} else if user.State == def.Offline {
						user.State = def.OffKick
					}

					//					fmt.Println("user.State::" + strconv.Itoa(int(user.State)))
					break
				}
			}

			if v, ok := r.QuitSitUsers[v.ID]; ok {
				if v.Accumulation != nil {
					user.Score = int32(v.Accumulation.Score)
				}
			}

			userList = append(userList, &user)
		}
	}

	roomUser := mjgame.ACK_Room_User{
		RID:   int32(r.RoomId),
		Users: userList,
	}

	return &roomUser
}

//初始化用户的牌
// uid 顺序
// sid 顺序
func (r *RoomXiangshan) InitUserCard() {
	index := 0

	for i, _ := range r.Seats {
		userCardRecord := "发牌:(" + r.Seats[i].User.NickName + "," + strconv.Itoa(i) + ") " //牌局信息
		for {
			if r.Seats[i].Cards.List.Count == InitUserCardsNumber {
				break
			}
			if index > 143 {
				fmt.Println("RoomXiangshan InitUserCard index > 143,index = ", index)
			}
			if &r.AllCards[index] == nil {
				fmt.Println("RoomXiangshan InitUserCard r.AllCards[index] == nil,index = ",
					index, "allcardlen = ", len(r.AllCards), "r.AllCards =", r.AllCards)
			}
			r.Seats[i].Cards.List.Add(&r.AllCards[index])
			userCardRecord += r.AllCards[index].MSG + " " // 牌局信息
			(r.StlCtrl).(st.IStatement).AddTool(st.T_Deal, r.Seats[i].Index, -1, []int{r.AllCards[index].ID})
			index++
		}
		r.RoomBase.RoomRecord += userCardRecord + "\r\n"
	}

	r.StartIndex = index
	r.CurMJIndex = index
}

//换花
func (r *RoomXiangshan) TransferHua() ([]int32, []int32) {
	var count, huaLength int
	var huaCardIDs, cardIDs []int32
	var moveCards []*rb.Card
	forward := true
	fromLast := forward
	seat := r.Seats[r.CurIndex]

	listLength := seat.Cards.List.Count
	for i := 0; i < listLength; i++ {
		if *seat.Cards.List.Index(i) != nil {
			card := (*seat.Cards.List.Index(i)).(*rb.Card)
			if card.Type == rb.H {
				huaLength++
				huaCardIDs = append(huaCardIDs, int32(card.ID))
				moveCards = append(moveCards, card)
			}
		}

	}

	for _, card := range moveCards {
		r.MoveToList(seat.Cards.List, []*rb.Card{card}, seat.Cards.Hua)
	}

	if huaLength == 0 {
		return huaCardIDs, cardIDs
	}

	for {
		card := r.GetNewCard(fromLast, r.CurIndex)
		if card != nil {
			if count > 0 {
				forward = false
			}
			count = 0
			if card.Type == rb.H { //是否花牌，花牌重新抓取
				seat.Cards.Hua.Add(card)
				count++
				fromLast = false
				huaCardIDs = append(huaCardIDs, int32(card.ID))
			} else {
				seat.Cards.List.Add(card)
				cardIDs = append(cardIDs, int32(card.ID))
			}

			if forward {
				(r.StlCtrl).(st.IStatement).AddTool(st.T_Mo, r.CurIndex, -1, []int{card.ID})
			} else {
				(r.StlCtrl).(st.IStatement).AddTool(st.T_MoBack, r.CurIndex, -1, []int{card.ID})
			}

			r.RoomBase.RoomRecord += "换花摸牌(" + r.Seats[r.CurIndex].User.NickName + ") " + card.MSG + "\r\n"

			if len(cardIDs) == huaLength {
				break
			}
		}
	}

	return huaCardIDs, cardIDs
}

// 广播换花
func (r *RoomXiangshan) BoardCastTransferHua() {
	huaCardIDs, cardIDs := r.TransferHua()
	if len(huaCardIDs) == 0 {
		return
	}

	transferHuaResponse := &mjgame.TransferHuaResponse{
		Index:   int32(r.CurIndex),
		HuaList: huaCardIDs,
		CardIds: cardIDs,
	}

	recordTransferHuaResponse := &mjgame.TransferHuaResponse{
		Index:   int32(r.CurIndex),
		HuaList: huaCardIDs,
		CardIds: cardIDs,
	}

	for k, v := range r.Seats {
		if k != r.CurIndex {
			transferHuaResponse.CardIds = []int32{}
		} else {
			transferHuaResponse.CardIds = cardIDs
		}
		if v.User != nil {
			v.User.SendMessage(mjgame.MsgID_MSG_ACKBC_TRANSFER_HUA, transferHuaResponse)
		}
	}

	r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_TRANSFER_HUA, recordTransferHuaResponse)

	transferHuaResponse.CardIds = []int32{}
	for _, v := range r.WatchSeats {
		if v != nil {
			v.SendMessage(mjgame.MsgID_MSG_ACKBC_TRANSFER_HUA, transferHuaResponse)
		}
	}
}

//判断是否可以吃牌
func (r *RoomXiangshan) CheckChow(pCard *rb.Card) {

	if pCard == nil || pCard.Type == rb.F || pCard.Type == rb.H || pCard.Type == rb.J {
		//fmt.Println("检查吃牌错误  ---", pCard)
		return
	}
	index := (r.CurIndex + 1) % 4

	r.RoomBase.RoomRecord += "测Chow(" + r.Seats[index].User.NickName + ") " + pCard.MSG + "  " + strconv.Itoa(pCard.Num) + "\r\n"
	if len(r.Seats[index].ChowCardIDs) > 0 && r.ChengBao[index].Seat[r.CurIndex] < 3 {
		for _, v := range r.Seats[index].ChowCardIDs {
			if r.CurCard.ID == v {
				r.RoomBase.RoomRecord += "测到过手吃(" + r.Seats[index].User.NickName + ") " + pCard.MSG + "  " + strconv.Itoa(pCard.Num) + "\r\n"
				return
			}
		}
	}

	list := r.Seats[index].Cards.List
	str := ""
	var n1, n2, n3, n4 int
	n1 = pCard.Num - 2
	n2 = pCard.Num - 1
	n3 = pCard.Num + 1
	n4 = pCard.Num + 2
	if n1 < 0 {
		n1 = -1
	}
	if n2 < 0 {
		n2 = -1
	}
	if n3 > 8 {
		n3 = -1
	}
	if n4 > 8 {
		n4 = -1
	}

	var c1, c2, c3, c4 bool
	for i := 0; i < list.Count; i++ {
		if *list.Index(i) != nil {
			card := (*list.Index(i)).(*rb.Card)
			if card.Type == pCard.Type {
				str += (card.MSG + strconv.Itoa(card.Num))
				if n1 != -1 && n1 == card.Num {
					c1 = true
				}
				if n2 != -1 && n2 == card.Num {
					c2 = true
				}
				if n3 != -1 && n3 == card.Num {
					c3 = true
				}
				if n4 != -1 && n4 == card.Num {
					c4 = true
				}
			}
		}

	}

	r.RoomBase.RoomRecord += "Chow  " + str + " \r\n"
	r.RoomBase.RoomRecord += "Chow  " + strconv.FormatBool(c1) + strconv.FormatBool(c2) + strconv.FormatBool(c3) + strconv.FormatBool(c4) + " \r\n"

	if (c1 && c2) || (c2 && c3) || (c3 && c4) {
		r.AddToolUser(index, 0, 0, 0, 1, 0, 1)
		r.RoomBase.RoomRecord += "检测(" + r.Seats[index].User.NickName + ") 吃" + pCard.MSG + "\r\n"
	}

}

//检查玩家换花以后是否可以吃
func (r *RoomXiangshan) CheckChowAfterHuanhua(tIdx int, pCard *rb.Card) bool {
	if pCard == nil || pCard.Type == rb.F || pCard.Type == rb.H || pCard.Type == rb.J {
		//fmt.Println("检查吃牌错误  ---", pCard)
		return false
	}

	index := tIdx

	list := r.Seats[index].Cards.List
	str := ""
	var n1, n2, n3, n4 int
	n1 = pCard.Num - 2
	n2 = pCard.Num - 1
	n3 = pCard.Num + 1
	n4 = pCard.Num + 2
	if n1 < 0 {
		n1 = -1
	}
	if n2 < 0 {
		n2 = -1
	}
	if n3 > 8 {
		n3 = -1
	}
	if n4 > 8 {
		n4 = -1
	}

	var c1, c2, c3, c4 bool
	for i := 0; i < list.Count; i++ {
		if *list.Index(i) != nil {
			card := (*list.Index(i)).(*rb.Card)
			if card.Type == pCard.Type {
				str += (card.MSG + strconv.Itoa(card.Num))
				if n1 != -1 && n1 == card.Num {
					c1 = true
				}
				if n2 != -1 && n2 == card.Num {
					c2 = true
				}
				if n3 != -1 && n3 == card.Num {
					c3 = true
				}
				if n4 != -1 && n4 == card.Num {
					c4 = true
				}
			}
		}

	}
	if (c1 && c2) || (c2 && c3) || (c3 && c4) {
		r.AddToolUser(index, 0, 0, 0, 1, 0, 1)
		return true
	}
	return false
}

//设置过手吃
func (r *RoomXiangshan) SetPassChowCards(pCard *rb.Card) {
	index := r.CurIndex % 4
	if pCard.Type == rb.F || pCard.Type == rb.H || pCard.Type == rb.J {
		return
	}
	if r.Seats[index].User == nil { //玩家断线则不操作,跳过
		return
	}
	list := r.Seats[index].Cards.List
	var n1, n2, n3, n4 bool
	for i := 0; i < list.Count; i++ {
		if *list.Index(i) != nil {
			card := (*list.Index(i)).(*rb.Card)
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
	isTrue := false
	if n1 && n2 {
		lPassCard := rb.GetCardById(pCard.ID - 3)
		if lPassCard != nil && lPassCard.Type == pCard.Type {
			r.Seats[index].ChowCardIDs = append(r.Seats[index].ChowCardIDs, lPassCard.ID)
		} else {
			//fmt.Println("SetPassChowCards lPassCard failed,ERR is lPassCard wrong")
		}
		isTrue = true
	}

	if n3 && n4 {
		rPassCard := rb.GetCardById(pCard.ID + 3)
		if rPassCard != nil && rPassCard.Type == pCard.Type {
			r.Seats[index].ChowCardIDs = append(r.Seats[index].ChowCardIDs, rPassCard.ID)
		} else {
			//fmt.Println("SetPassChowCards rPassCard failed,ERR is rPassCard wrong")
		}
		isTrue = true

	}
	if (n2 && n3) || isTrue {
		r.Seats[index].ChowCardIDs = append(r.Seats[index].ChowCardIDs, pCard.ID)
	}
}

//是否可以解散
func (r *RoomXiangshan) GetDisApproveCount() int {
	var count int
	for _, v := range r.Votes {
		if v == DisApprove {
			count++
		}
	}
	return count
}
