// 宁波麻将 GameType : 4000 这个是斗地主告诉你哈哈哈，名字回头再说
package pinshi

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
	//	"PZ_GameServer/server/game/room/pinshi/paixingLogic"
	rb "PZ_GameServer/server/game/roombase"
	st "PZ_GameServer/server/game/statement"
	"PZ_GameServer/server/user"
	//	"encoding/json"
	"sync"
)

//开局筛子判断常量
var (
	FrontPoints   = []int{2, 6, 10}
	BackPoints    = []int{4, 8, 12}
	FacePoints    = []int{3, 7, 11}
	SelfPoints    = []int{5, 9}
	IFCPinshiType = reflect.TypeOf(&RoomPinshi{})

	MingQiangzhuang   = 101 //=(清12混8)
	NormalQiangzhuang = 102 //=(清10混6
)

//投票常量
const (
	//0 = 未操作
	Agree      = 1 //同意
	DisApprove = 2 //不同意
)

//开局类型
const (
	CalcRound = 1 //按局数
	CalcTime  = 2 //按时间
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

const (
	Stage_qiangzhuang = 1
	Stage_xiazhu      = 2
	Stage_dapai       = 3
	Stage_End         = 4

	QiangTime  = 10
	XiazhuTime = 10
	LiangTime  = 12
	ReayTime   = 8
)

// 西周规则
var Pinshi_RoomRule = rb.RoomRule{
	GameType:           4002,         //
	Create_NeedDiamond: 100,          // 创建房间需要的钻石
	SeatLen:            10,           // 座位数量  2, 3, 4
	DefCardLen:         5,            // 默认手牌数量 13
	AllCardLen:         52,           // 牌的数量
	Card_W:             1,            // 默认带万 1代表一次
	Card_B:             1,            // 默认带饼
	Card_T:             1,            // 默认带条
	Card_F:             1,            // 默认带风
	Card_J:             1,            // 默认带箭 (中发白)
	Card_H:             1,            // 默认带花
	Card_Else:          []int{},      // 特殊牌
	CanLaiZi:           0,            // 赖子数量
	CanPeng:            1,            // 可以碰
	CanChow:            1,            // 可以吃
	CanKong:            1,            // 可以直杠
	CanAnKong:          1,            // 可以暗杠
	CanPengKong:        1,            // 可以碰杠
	CanTing:            1,            // 可以听
	CanWin:             1,            // 可以直胡
	CanZiMo:            1,            // 可以自摸胡
	MaxWinCount:        1,            // 最大胡牌数量<0为不限次数, 大众麻将为1
	MaxTime:            0,            // 最大时间(<=0 为不限时间)
	MaxRound:           0,            // 最大局数(<=0 为不限局数)
	BaseScore:          1,            // 底分
	MaxTai:             0,            // 封顶台数(<=0 为不限)
	WinNeedTai:         0,            // 胡牌最小台数 (<=0 为不限)
	Rules:              []int32{},    // 全部规则(包含特殊规则)
	Temp:               RoomPinshi{}, // 模板
}

type RoomPinshi struct {
	rb.RoomBase                               //
	ChengBao         []ChengBaoSeat           //
	FengQuan         int                      // 风圈(0-3)东南西北
	Status           int                      //
	Bankers          []int                    // 风圈变了,清空
	Votes            []int                    // 投票  -1未操作   0反对   1同意
	VoteTimeCount    int                      // 投票超时
	IsKongHu         bool                     // 是否拉杠胡
	KongHuCardID     int                      // 拉杠胡cardid
	IsContinueBanker bool                     // 是否连庄
	RoundResult      *mjgame.ACKBC_Sddz_Total // 记录每一局信息，做短线重连使用,后一局覆盖前一局
	KickUsers        []*KickInfo              // 被踢掉的玩家信息
	QuitSitUsers     map[int]rb.SeatBase      // 退出参与玩家信息
	PlayType         int                      // 叫地主类型   101 叫分 102 抢地主
	Mux              sync.RWMutex             // map读写锁
	ZhuangPos        int                      //地主的位置
	StartIdx         int                      //每局开始叫地主的位置
	Difen            int                      //底分
	Beishu           int                      //倍数
	Stage            int                      //阶段
	CurDizhuIdx      int                      //最后叫地主的玩家
	Dipai            []int                    //底牌
	CurOutputCards   *mjgame.Sddz_Chupai      //最后出的牌
	CurOutputIdx     int                      //最后出牌的位置
	LeftTime         int                      //当前阶段剩余时间
	CanStartInto     bool                     //是否可以在游戏开始后加入
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
type Seat_Pinshi struct {
	rb.SeatBase
}

// 初始化
func (r *RoomPinshi) Init() {
	r.RoomBase.Init()
	r.State = 0
	r.IsDraw = false
	r.IsRun = true
	r.WinUserCount = 0
	r.IsKongHu = false
	r.RoundResult = nil
	r.RoomBase.MToolChecker = rb.ToolChecker{}
	r.RoomBase.MToolChecker.Init(r.Rules.SeatLen)
	r.RoomBase.RoomRecord = ""
	r.RoomBase.RoundTotaled = false
	r.Show = false
	r.Votes = make([]int, r.Rules.SeatLen)
	r.RoomBase.VoteStarter = -1
	r.KickUsers = []*KickInfo{}
	r.FanCard = nil
	r.ZhuangPos = -1
	r.Difen = 0
	r.Beishu = 1
	r.StartIdx = 0
	r.CurDizhuIdx = -1
	r.Dipai = []int{}
	r.CurOutputCards = nil
	r.CurOutputIdx = -1
	r.Stage = 0
	r.LeftTime = 0

	if r.RoundCount == 0 {
		for _, v := range r.Seats {
			v.Accumulation = &rb.Accumulation{}
		}
		r.StartTime = int(time.Now().Unix())

	}

	r.ChengBao = make([]ChengBaoSeat, r.Rules.SeatLen)
	for i := 0; i < r.Rules.SeatLen; i++ {
		r.ChengBao[i].Seat = make([]int, r.Rules.SeatLen)
		r.Seats[i].IsPutCard = false
		r.Seats[i].HaveJiao = false
		r.Seats[i].JiaoFen = 0
		r.Seats[i].XiaZhu = 0
		r.Seats[i].IsMing = false
		r.Seats[i].IsJiabei = false
		r.Seats[i].HaveJiabei = false
		r.Seats[i].OutputNum = 0
		r.Seats[i].LastOpt = rb.Last_opt_non
		r.Seats[i].LastOptParam = nil
		r.Seats[i].JoinPlay = false
	}
}

//定时器
func (r *RoomPinshi) TimeTicker() {
	var flag bool

	for {

		if r.StopTicker {
			flag = true
			break
		}

		r.LeftTime++

		if r.Stage == Stage_End && r.RoundCount > 0 {
			if r.LeftTime >= ReayTime {
				//自动准备（准备）
				for _, v := range r.Seats {
					if v.User != nil && v.State != int(mjgame.StateID_UserState_Ready) {
						r.Ready(v.User)
					}
				}
			}
		} else if r.Stage == Stage_qiangzhuang {
			if r.LeftTime >= QiangTime {
				//自动抢庄（不抢）
				for _, v := range r.Seats {
					if v.JoinPlay && v.User != nil && v.HaveJiao == false {
						r.Jiaofen(v.User, 0)
					}
				}

			}
		} else if r.Stage == Stage_xiazhu {
			if r.LeftTime >= XiazhuTime {
				for _, v := range r.Seats {
					if v.JoinPlay && v.User != nil && v.HaveJiabei == false {
						r.Xiazhu(v.User, 2)
					}
				}
			}

		} else if r.Stage == Stage_dapai {
			if r.LeftTime >= LiangTime {
				//自动亮牌
				for _, v := range r.Seats {
					if v.JoinPlay && v.User != nil && v.IsMing == false {
						r.Mingpai(v.User)
					}
				}
			}
		}

		time.Sleep(1 * time.Second)

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

		//房间到了，自动解散
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
				//当前正在进行中，有玩家掉线超过180s,游戏流局，房间结束
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

		} else if r.Rules.MaxRound >= 0 && r.StartTime > 0 {

			if !r.IsRun { //两局之间
				//当前正在进行中，有玩家掉线超过180m,游戏流局，房间结束
				if r.ByRoundCanFinish() {
					//fmt.Println("有玩家掉线超过180m,游戏流局，房间结束 ", r.RoomId)
					flag = true
					r.StopTicker = true
					r.DestoryRoom()
					r.Mux.Lock()
					rb.ChanRoom <- r.RoomId //销毁房间
					r.Mux.Unlock()
				}
			} else {
				//当前正在进行中，有玩家掉线超过180m,游戏流局，房间结束
				if r.RoundCanFinish() {
					fmt.Println("有玩家掉线超过180m,游戏流局，房间结束 ", r.RoomId)
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

		if flag {
			fmt.Println("timetick flag break")
			break
		}
	}
}

// 检查是否可以开始
func (r *RoomPinshi) CheckCanStart() (bool, *mjgame.ErrorItem) {
	for i := 0; i < len(r.Seats); i++ {
		if r.Seats[i] == nil { //if r.Seats[i] == nil || r.Seats[i].User == nil {
			return false, error.ErrSomePeopleNotReady
		}
	}

	if r.Rules.MaxRound > 0 {
		if r.RoundCount >= r.Rules.MaxRound {
			return false, error.ErrCurRoundHasOver
		}
	}

	// 检查是否超时
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
func (r *RoomPinshi) AddChengBao(index int, tindex int) {
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
func (r *RoomPinshi) InitRound() {

	r.StartIdx = util.RandInt(0, r.Rules.SeatLen) //确定了从谁开始叫
	ack := mjgame.ACKBC_Sddz_Start{
		RoundCount:      int32(r.RoundCount),
		TotalRoundCount: int32(r.Rules.MaxRound),
	}
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Sddz_Start, &ack) // 广播游戏开始消息
}

// 一局结算
func (r *RoomPinshi) RoundTotal() {

	if r.RoomBase.RoundTotaled {
		fmt.Println("已经结算过.", r.RoomId)
		return
	}
	r.RoomBase.RoundTotaled = true

	r.RoomBase.RoomRecord += "开始结算\r\n"
	huCard := -1
	attached := ""
	var huIndexes []int
	var lastTool *st.OnceRecord

	if r.StlCtrl != nil && (r.StlCtrl).(*Pinshi_Statement) != nil {
		lastTool = (r.StlCtrl).(*Pinshi_Statement).Get(0)
	}

	if lastTool != nil && lastTool.Tool.ToolType == st.T_Hu_ZiMo {
		huIndex := lastTool.Tool.Index
		huCard = lastTool.Tool.Val[0]
		if huIndex < 0 || huCard < 0 {
			return
		}
		huIndexes = append(huIndexes, huIndex)
	}
	if lastTool.Tool.ToolType == st.T_Hu {
		huIndexes, huCard = r.GetMultiWinInfo()
	}

	if r.RoundCount == 0 {
		room := &model.Room{
			Type:         int(mjgame.MsgID_GTYPE_Pinshi),
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

	var scores = make([]int32, r.Rules.SeatLen)
	var ackTotal *mjgame.ACKBC_Total
	if lastTool.Tool.ToolType == st.T_Draw {
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
			TotalScore: make([]int32, r.Rules.SeatLen),
			TotalMsg:   make([]string, r.Rules.SeatLen),
		}
		totalTai := make([]int32, r.Rules.SeatLen)

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
			if result.Attached != "" {
				attached = result.Attached
			}
		}

		r.UpdateScore(total.TotalScore) // 更新分数
		maxScoreIndex := r.GetMaxIndex(total.TotalScore)

		var rewards = make([]*mjgame.Reward, 0)
		for k, v := range total.TotalScore {
			if r.Seats[k] != nil && r.Seats[k].Accumulation != nil {
				o := &mjgame.Reward{
					Score:      v,
					TotalScore: r.Seats[k].Accumulation.Score,
				}
				rewards = append(rewards, o)
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
			Attached:   attached,
		}
		scores = total.TotalScore
	}

	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Total, ackTotal)
	r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Total, ackTotal)
	//r.RoundResult = ackTotal
	r.InsertRoomRecord(scores)
	//房间结束
	if !flag {
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

	if !r.IsContinueBanker { // 如果庄胡 或者 流局 则连庄
		r.CalcDirection()
	}
}

// 获得开始信息  开始叫分的位置
//func (r *RoomPinshi) GetStartInfo() (int, int, int, int) {

//	startIdx := 0;
//	//seed := rand.New(rand.NewSource(time.Now().UnixNano()))
//	rollIndex := 0
//	bankerIndex := 0
//	point1 := 0
//	point2 := 0

//	//rollIndex = r.StartIndex

//	if r.RoundCount == 0 { // fix 只有第一局需要掷骰子
//		_, rollIndex = r.GetFirstSitSeatInfo()
//		bankerIndex = rollIndex
//		point1 = util.RandInt(1, 6)
//		point2 = util.RandInt(1, 6)
//	}

//	return rollIndex, bankerIndex, point1, point2
//}

// 改变用户座位
func (r *RoomPinshi) ChangeUsersPosition() []*mjgame.ACK_User_Info {
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
func (r *RoomPinshi) FanCalc(index int, cardid int) st.TotalResult {
	arrMj := r.Seats[index].Cards.List
	mjs := make([]int, 42)
	for i := 0; i < arrMj.Count; i++ {
		if *arrMj.Index(i) != nil {
			c := (*arrMj.Index(i)).(*rb.Card)
			mjs[c.ID]++
		}

	}
	mjs[cardid]++

	//fmt.Println(mjs)

	arg := make([]interface{}, 0)
	(r.StlCtrl).(*Pinshi_Statement).BaseCtl.SiChuan = false
	return (r.StlCtrl).(*Pinshi_Statement).FanCalc(index, arg)
}

//检查是否胡牌
func (r *RoomPinshi) CheckHu(index int, cardId int) int {
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

	//fmt.Println(r.Seats[index].User.NickName, mjs)
	result := (r.StlCtrl).(*Pinshi_Statement).CheckHu(mjs)
	//fmt.Println(result, "CheckHu result")
	if result > 0 {
		isWin := (r.StlCtrl).(*Pinshi_Statement).CheckCanWin(index)
		if !isWin {
			result = 0
		}
	}

	return result
}

//添加操作
func (r *RoomPinshi) AddTool(toolType int, index int, tindex int, val []int) {
	(r.StlCtrl).(*Pinshi_Statement).AddTool(toolType, index, tindex, val)
}

func (r *RoomPinshi) AddListCard(index int, listcard *al.ArrayList) {

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

//吃
func (r *RoomPinshi) CheckCanChow(index int, tIndex int, cards []*rb.Card, card *rb.Card) bool {
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
func (r *RoomPinshi) CheckPeng(pCard *rb.Card) {
	index := r.CurIndex
	for i := 0; i < 3; i++ {
		index++ //从下家开始判断
		index = index % r.Rules.SeatLen

		if r.Seats[index].User == nil {
			continue
		}
		cardCount := r.GetUserCardCount(index, pCard.ID)
		if cardCount >= 2 {
			var flag bool
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
func (r *RoomPinshi) NotifyTool() {
	r.RoomBase.NotifyTool()
}

// 转到下一个出牌的玩家
// get = 摸牌
// forward = 从后面摸牌
// pass = 是否过掉本次操作
func (r *RoomPinshi) TurnNextPlayer() {
	r.CurIndex++
	r.CurIndex = r.CurIndex % r.Rules.SeatLen

	if !r.IsRun {
		return
	}

	uid, err := strconv.Atoi(r.Seats[r.CurIndex].UID)
	if err != nil {
		return
	}

	switch r.Stage {
	case Stage_qiangzhuang:
		ack := mjgame.ACKBC_Sddz_JiaoDizhu{Uid: int32(uid), CurFen: int32(r.Difen)}
		r.BCMessage(mjgame.MsgID_MSG_ACKBC_Sddz_JiaoDizhu, &ack) // 广播当前出牌的玩家

		for i := 0; i < r.Rules.SeatLen; i++ {
			r.Seats[i].Message = &rb.Message{
				Id:      mjgame.MsgID_MSG_ACKBC_Sddz_JiaoDizhu,
				Content: &ack,
			}
		}
	//				int32  Uid = 1;		//通知叫地主的玩家
	//				int32 CurFen = 2; //当前的叫分 玩家叫的分必须大于这个分
	case Stage_dapai:
		ack := mjgame.ACKBC_CurPlayer{Seat: int32(r.Seats[r.CurIndex].Index), Type: int32(rb.WaitPut), RoundTime: int32(r.WaitPutTimeOut)}
		r.BCMessage(mjgame.MsgID_MSG_ACKBC_CurPlayer, &ack) // 广播当前出牌的玩家

		for i := 0; i < r.Rules.SeatLen; i++ {
			r.Seats[i].Message = &rb.Message{
				Id:      mjgame.MsgID_MSG_ACKBC_CurPlayer,
				Content: &ack,
			}
		}
	}

	return
	//	r.WaitOptTool.ClearAll()
	//	r.WaitOptTool.IsSelf = true
	//	r.Seats[r.CurIndex].IsCanWin = false
	//	r.Seats[r.CurIndex].IsCanKong = false
	//	r.Seats[r.CurIndex].IsCanPeng = false
	//	r.Seats[r.CurIndex].IsCanChow = false
	//	//fmt.Println("-----------------------------------------> ", r.Seats[r.CurIndex].User.NickName, r.CurIndex, r.Seats[r.CurIndex].Cards.List.Count)
	//	ack := mjgame.ACKBC_CurPlayer{Seat: int32(r.Seats[r.CurIndex].Index), Type: int32(rb.WaitPut), RoundTime: int32(r.WaitPutTimeOut)}
	//	r.BCMessage(mjgame.MsgID_MSG_ACKBC_CurPlayer, &ack) // 广播当前出牌的玩家

	//	if get && r.Seats[r.CurIndex].Cards.List.Count%3 == 2 { //如果再摸就是相公牌了
	//		get = false
	//	}

	//	canHu := 0
	//	kongType := 0

	//	if get {
	//		// 摸牌
	//		var ackbc_GetCard mjgame.ACKBC_GetCard
	//		r.CurCard, ackbc_GetCard = r.GetNextCard(forward) // 摸牌并发送数据
	//		r.Show = true
	//		r.Seats[r.CurIndex].HuCardIDs = []int{}

	//		if r.CurCard != nil {

	//			//fmt.Println("GetCard----------------------------------------------> ", r.Seats[r.CurIndex].User.NickName, r.CurIndex, r.Seats[r.CurIndex].Cards.List.Count)
	//			if r.Seats[r.CurIndex].Cards.List.Count != 13 && r.Seats[r.CurIndex].Cards.List.Count%3 == 1 {
	//				fmt.Println("手牌数量错误. ", r.CurIndex, r.Seats[r.CurIndex].User.NickName, r.CurCard)
	//			}
	//			strHu := ""
	//			strKong := ""
	//			canHu = r.CheckHu(r.CurIndex, -1) //检查是否能胡
	//			if !pass {
	//				_, kongType = r.CheckCanKong(r.CurIndex, r.CurCard.ID, true) //检查是否可以杠牌
	//			}
	//			if canHu > 0 {
	//				r.AddToolUser(r.CurIndex, 1, 0, 0, 0, 1, 1)
	//				strHu = "胡"
	//			}

	//			if kongType <= 0 {
	//				kongType = 0
	//			} else {
	//				strHu = "杠"
	//			}
	//			r.RoomBase.RoomRecord += "转到(" + r.Seats[r.CurIndex].User.NickName + ") 摸牌:" + r.CurCard.MSG + " " + strHu + strKong + "\r\n"
	//			ackbc_GetCard.Tool[0] = int32(canHu)
	//			ackbc_GetCard.Tool[1] = int32(kongType)
	//			r.RoomBase.MToolChecker.SetCptTool(r.CurIndex, int(mjgame.MsgID_MSG_ACKBC_GetCard), []int{r.CurCard.ID}, r.Seats[r.CurIndex].User.NickName)
	//			r.RoomBase.MToolChecker.ShowAllTools()
	//			for _, v := range r.Seats {
	//				if v.Index == r.CurIndex {
	//					ackbc_GetCard.Cid = int32(r.CurCard.ID)
	//				} else {
	//					ackbc_GetCard.Cid = -1
	//					ackbc_GetCard.Tool[0] = 0
	//					ackbc_GetCard.Tool[1] = 0
	//				}
	//				v.User.SendMessage(mjgame.MsgID_MSG_ACKBC_GetCard, &ackbc_GetCard)
	//			}

	//			rec := [5]interface{}{ackbc_GetCard.Index, r.CurCard.ID, ackbc_GetCard.FromLast, canHu, kongType}
	//			r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_GetCard, &rec)

	//		} else {
	//			// 没有牌摸的情况下,  流局
	//			//if r.CheckDraw() {
	//			r.Draw()
	//			//}
	//		}
	//	} else {
	//		// 不摸牌
	//		//fmt.Println("NoGetCard----------------------------------------------> ", r.Seats[r.CurIndex].User.NickName, r.CurIndex, r.Seats[r.CurIndex].Cards.List.Count)

	//		r.RoomBase.MToolChecker.SetAllUserTool(-1)
	//		lastopt := r.WaitOptTool.GetOpt(r.CurIndex)

	//		canKong := 0
	//		strKong := ""
	//		if lastopt != nil && !pass {
	//			_, kongType = r.CheckCanKong(r.CurIndex, -1, true)
	//			if kongType > 0 {
	//				canKong = 1
	//				strKong = "杠"
	//			}
	//		}

	//		r.RoomBase.RoomRecord += "转到(" + r.Seats[r.CurIndex].User.NickName + ") 不摸牌 " + strKong + "\r\n"
	//		r.RoomBase.MToolChecker.SetTools(r.CurIndex, []int{-1, canKong - 1, -1, -1, 0, -1, -1, -1}) // 胡0 杠1 碰2 吃3 出4 过5 摸6 等7
	//		r.RoomBase.MToolChecker.ShowAllTools()
	//	}

	//	if r.IsRun {

	//		//fmt.Println("r.NeedWaitTool.Count ", r.WaitOptTool.Count())
	//		r.CurToolIndex = r.CurIndex
	//		r.WaitTimeCount = r.WaitToolTimeOut
	//		if r.WaitOptTool.Count() > 0 {
	//			r.NotifyTool()
	//		} else {
	//			r.State = rb.WaitPut
	//		}

	//	}
}

//等待操作,自摸判断操作
func (r *RoomPinshi) WaitPut(timeout int) {

}

//开始等待操作
func (r *RoomPinshi) StartWaitTool(card *rb.Card) {
	r.WaitOptTool.ClearAll()
	r.WaitOptTool.IsSelf = false

	r.CheckWin(card.ID)
	r.CheckCanKong(r.CurIndex, card.ID, false)
	r.CheckPeng(card)
	r.CheckChow(card)
	r.RoomBase.RoomRecord += "检查操作(" + r.Seats[r.CurIndex].User.NickName + ") " + strconv.Itoa(r.WaitOptTool.Count()) + "\r\n"

	r.RoomBase.UpdateToolChecker()

	if r.WaitOptTool.Count() > 0 { //等待操作
		r.CurToolIndex = 0 // 从0 开始
		r.WaitTimeCount = r.WaitToolTimeOut
		r.NotifyTool()
		r.WaitPutTool()

	} else {
		//是否流局
		if r.CheckDraw() {
			r.Draw()

		} else {
			//没有要操作的
			r.WaitTimeCount = 0
			r.CurToolIndex = -1
			r.Status = rb.WaitPut
			//r.TurnNextPlayer(true, true, false)
		}
	}
}

//检查是否可以胡
func (r *RoomPinshi) CheckWin(cid int) {
	index := r.CurIndex

	for i := 0; i < 3; i++ {
		index++
		index = index % r.Rules.SeatLen
		r.Seats[index].IsCanWin = false //复位
		result := r.CheckHu(index, cid)
		if result > 0 {
			var flag bool
			for _, v := range r.Seats[index].HuCardIDs {
				if v == cid {
					flag = true
					break
				}
			}

			if flag {
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
func (r *RoomPinshi) CheckDraw() bool {

	result := r.AllCardLength - r.EndBlank - r.CurMJIndex //- def.XiangShanDrawCount
	if result <= 0 {
		r.AddTool(st.T_Draw, -1, -1, []int{})
		return true
	}
	return false
}

// 流局
func (r *RoomPinshi) Draw() {

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

	//回放记录
	recordDraw := &mjgame.ACKBC_Draw{
		Cards: r.GetRecordAllSeatCards(),
	}
	r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Draw, recordDraw)

	r.RoundTotal()
	r.RoundToatlFinish = true

}

//计算风圈
func (r *RoomPinshi) CalcDirection() {
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
		r.FengQuan = r.FengQuan % r.Rules.SeatLen
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
func (r *RoomPinshi) GetSummaryList() mjgame.TotalSummary {
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

//是否可以解散
func (r *RoomPinshi) IsDisbanding() int {
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

	if agreeCount > r.GetRealPlayNum()/2 {
		result = 1
	} else if noagreeCount >= r.GetRealPlayNum()/2 {
		result = 2
	}
	return result
}

//获取实际玩家的人数，作为可以为空
func (r *RoomPinshi) GetRealPlayNum() int {
	var count = 0
	for _, v := range r.Seats {
		if v.User != nil {
			count++
		}
	}
	return count
}

func (r *RoomPinshi) DestoryRoom() {

	r.StopTicker = true
	//发送流局消息
	allSeatCards := r.GetAllSeatCards()
	//dipai := []int32{int32(r.Dipai[0]), int32(r.Dipai[1]), int32(r.Dipai[2])}
	end := &mjgame.ACKBC_Sddz_End{
		Uid:   -1,
		Cards: allSeatCards,
		//Dipai: dipai,
	}
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Sddz_End, end)

	//	recordDraw := &mjgame.ACKBC_Draw{
	//		Cards: r.GetRecordAllSeatCards(),
	//	}
	//	r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Draw, recordDraw)

	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Sddz_Total, &mjgame.ACKBC_Sddz_Total{
		Finished:   true,
		RoundCount: int64(r.RoundCount),
	})

	//	r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Total, &mjgame.ACKBC_Total{
	//		Finished:   true,
	//		RoundCount: int64(r.RoundCount),
	//	})

	//插入流局记录
	if r.RoundCount == 0 {
		room := &model.Room{
			Type:         int(mjgame.MsgID_GTYPE_Pinshi),
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
	var scores = make([]int32, r.Rules.SeatLen)
	r.InsertRoomRecord(scores)

	//发送大结算
	list := r.GetSummaryList()
	r.BCMessage(mjgame.MsgID_MSG_NOTIFY_SUMMARY, &list)
	r.Votes = []int{0, 0, 0}
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

func (r *RoomPinshi) InsertRoomRecord(scores []int32) {

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

			if userId != 0 {
				if _, ok := result[userId]; !ok {
					result[userId] = int(v)
				}
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

		//		PlayBack, e := json.Marshal(r.BattleRecord)
		//		if e == nil {
		//			battleRecord.PlayBack = string(PlayBack)
		//			//fmt.Println("json:: shuju :: " + battleRecord.PlayBack)
		//		}

		if err := m.Create(battleRecord).Error; err != nil {
			m.Rollback()

			return

		}

		for k, v := range result {
			roomRecord := &model.RoomRecord{
				RoomID:         int(r.DbRoomId),
				UserID:         k,
				BattleRecordID: int(battleRecord.ID),
				RoomType:       int(mjgame.MsgID_GTYPE_Pinshi),
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

		//以上是原来的  重播暂时注释掉 后面再来弄这个

		//		m = model.BeginCommit()
		//		replayRecord := &model.ReplayRecord{
		//			ReviewCode: tempStr,
		//		}

		//		PlayBack, e := json.Marshal(r.BattleRecord)
		//		if e == nil {
		//			replayRecord.PlayBack = string(PlayBack)
		//			//fmt.Println("json:: shuju :: " + battleRecord.PlayBack)
		//		}

		//		if err := m.Create(replayRecord).Error; err != nil {
		//			m.Rollback()
		//			return
		//		}

		//		m.Commit()

	}, func(e interface{}) {
		fmt.Println("pinshi InsertRoomRecord ", e)
	})
}

//进入房间
func (r *RoomPinshi) IntoRoom(user *user.User) {
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

	leftTime := 0

	if r.Stage == Stage_End {
		leftTime = ReayTime - r.LeftTime
	} else if r.Stage == Stage_qiangzhuang {
		leftTime = QiangTime - r.LeftTime
	} else if r.Stage == Stage_xiazhu {
		leftTime = XiazhuTime - r.LeftTime
	} else if r.Stage == Stage_dapai {
		leftTime = LiangTime - r.LeftTime
	}

	ack := &mjgame.ACKBC_Into_Room{
		Name:    user.NickName,
		Uid:     strconv.Itoa(user.ID),
		RoomId:  int32(user.RoomId),
		Ip:      user.GetIP(),
		Index:   -1,
		Icon:    user.Icon,
		Coin:    int32(user.Coin),
		Type:    int32(mjgame.MsgID_GTYPE_Pinshi),
		Diamond: int32(user.Diamond),
		Level:   0,
		Robot:   int32(user.IsRobot),
		GPS_LNG: user.GPS_LNG,
		GPS_LAT: user.GPS_LAT,
		Rule:    r.Rules.Rules,
	}

	roomInfo := r.GetRoomInfo() //房间信息
	roomInfo.LeftTime = int64(leftTime)
	user.SendMessage(mjgame.MsgID_MSG_ACK_RoomInfo, roomInfo)

	if isWatch {
		r.BCMessage(mjgame.MsgID_MSG_ACKBC_Into_Room, ack)
	} else {
		user.SendMessage(mjgame.MsgID_MSG_ACKBC_Into_Room, ack)
	}

	user.SendMessage(mjgame.MsgID_MSG_ACK_Room_User, r.GetRoomUser())

	if r.IsRun {
		r.SendSeatCard(user.ID)
		r.SendGameInfo(user, false)
		if r.VoteStarter >= 0 { //解散状态
			r.DisbandRoom(user, nil)
		}

	} else {
		if index >= 0 {
			if r.Seats[index].State == int(mjgame.StateID_GameState_Total) {
				r.SendSeatCard(user.ID)
				r.SendGameInfo(user, false)

			}
		}
	}
}

// 发送玩家牌消息
func (r *RoomPinshi) SendSeatCard(userId int) {

	allUserCards := make([]*mjgame.SeatCard, 0)

	//fmt.Println("SendSeatCard ...")
	var index int = -1
	for i, v := range r.Seats { // TODO : 这里可能报错

		if v.Cards == nil { //零时Fix
			continue
		}

		ack_card := &mjgame.SeatCard{
			Seat: int32(v.Index),
			Out:  r.GetListArray(v.Cards.Out),
		}

		if r.Stage == Stage_dapai { //打牌阶段
			if v.JoinPlay == true {
				ack_card.ListLen = int32(v.Cards.List.Count)
			} else if v.User != nil && v.User.ID == userId {
				index = i
			}

			if v.JoinPlay == true && v.User != nil && v.User.ID == userId {
				ack_card.List = r.GetListArray(v.Cards.List) //append(v.Cards.List[:len(v.Cards.List)-1])
				index = i

				//				test := make([]int, 0)
				//				for i := 0; i < len(ack_card.List); i++ {
				//					test = append(test, int(ack_card.List[i].Cid))

				//				}
				//				paixingLogic.GetPaixing(test)
			}
		} else { //只给前4张牌

			if v.JoinPlay == true {
				ack_card.ListLen = int32(v.Cards.List.Count) - 1
			} else if v.User != nil && v.User.ID == userId {
				index = i
			}

			if v.JoinPlay == true && v.User != nil && v.User.ID == userId {
				preList := al.New()

				for i := 0; i < 4; i++ {
					preList.Add(*v.Cards.List.Index(i))
				}
				//fmt.Println("preList len:", preList.Count)
				ack_card.List = r.GetListArray(preList) //append(v.Cards.List[:len(v.Cards.List)-1])

				index = i
				//fmt.Println("ack_card.List len:", len(ack_card.List))
			}

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

func (r *RoomPinshi) SendGameInfo(a *user.User, needRecord bool) {
	var iRun int //是否游戏运行中
	if r.IsRun {
		iRun = 1
	}

	//	 repeated int32 CurCards 	= 1;           // 当前打出的牌
	//	 int32 CurIndex  			= 2;           // 当前用户(需要出牌的)
	//	 int32 CurTime   			= 3;           // 当前回合时间
	//	 int32 DizhuIndex  		    = 4;       	   // 地主Index
	//	 int32 StartGame 			= 5;           // 是否已经开始 0=否 1=是
	//	 int32 CurOutputIndex 		= 6; 		   // 当前出牌的玩家
	//	 int32 Difen                  = 7;   //当前的底分
	//	 repeated int32 Dipai  = 8; //当前的底牌
	//	 int32 Stage              = 9;//当前的阶段
	//	 repeated bool HaveJiao = 10;//玩家是否叫过分
	//	 repeated int32 JiaoFen = 11;//玩家的叫分
	//	 repeated bool HaveJiabei = 12;//是否选择过加倍
	//	 repeated bool IsJiabei =13;//是否加倍
	//	 repeated bool IsMing = 14;//玩家是否明牌

	//	dipai := make([]int32, 3)

	//fmt.Println("r.DizhuPos", r.ZhuangPos)

	//	if r.ZhuangPos != -1 {
	//		for i := 0; i < len(r.Dipai); i++ {
	//			dipai[i] = int32(r.Dipai[i])
	//		}
	//	}

	//	fmt.Println("dipai", dipai)

	haveJiao := make([]bool, r.Rules.SeatLen)
	jiaoFen := make([]int32, r.Rules.SeatLen)
	haveJiabei := make([]bool, r.Rules.SeatLen)
	xiazhu := make([]int32, r.Rules.SeatLen)
	isJiabei := make([]bool, r.Rules.SeatLen)
	isMing := make([]bool, r.Rules.SeatLen)
	lastOpt := make([]int32, r.Rules.SeatLen)
	lastOptParam := make([]*mjgame.SddzPerCards, r.Rules.SeatLen)

	for i := 0; i < r.Rules.SeatLen; i++ {
		haveJiao[i] = r.Seats[i].HaveJiao
		jiaoFen[i] = int32(r.Seats[i].JiaoFen)
		haveJiabei[i] = r.Seats[i].HaveJiabei
		xiazhu[i] = int32(r.Seats[i].XiaZhu)
		isJiabei[i] = r.Seats[i].IsJiabei
		isMing[i] = r.Seats[i].IsMing
		lastOpt[i] = int32(r.Seats[i].LastOpt)
		lastOptParam[i] = &mjgame.SddzPerCards{
			Cards: r.Seats[i].LastOptParam,
		}
	}

	var cards []int32

	if r.CurOutputCards != nil {
		cards = r.CurOutputCards.Cards
	}
	var leftTime = 0

	if r.Stage == Stage_End {
		leftTime = ReayTime - r.LeftTime
	} else if r.Stage == Stage_qiangzhuang {
		leftTime = QiangTime - r.LeftTime
	} else if r.Stage == Stage_xiazhu {
		leftTime = XiazhuTime - r.LeftTime
	} else if r.Stage == Stage_dapai {
		leftTime = LiangTime - r.LeftTime
	}

	ack := &mjgame.ACKBC_Sddz_Card_Init{
		CurCards:       cards,
		CurIndex:       int32(r.CurIndex),
		CurTime:        int32(r.WaitToolTimeOut), //
		DizhuIndex:     int32(r.ZhuangPos),
		StartGame:      int32(iRun), //
		CurOutputIndex: int32(r.CurOutputIdx),
		Difen:          int32(r.Difen),
		//		Dipai:          dipai,
		Stage:        int32(r.Stage),
		HaveJiao:     haveJiao,
		JiaoFen:      jiaoFen,
		HaveJiabei:   haveJiabei,
		XiaZhu:       xiazhu,
		IsJiabei:     isJiabei,
		IsMing:       isMing,
		Beishu:       int32(r.Beishu),
		LastOpt:      lastOpt,
		LastOptParam: lastOptParam,
		BaodaoNum:    int32(leftTime), //这里是当前阶段玩家剩余操作的时间
	}

	//	fmt.Println("dipai1", dipai)

	if a == nil {
		r.BCMessage(mjgame.MsgID_MSG_ACKBC_Sddz_Card_Init, ack)
	} else {
		a.SendMessage(mjgame.MsgID_MSG_ACKBC_Sddz_Card_Init, ack)
	}

	//
	//	ack := &mjgame.ACKBC_Card_Init{ //  ------ 牌面信息
	//		Dict:         int32(r.Dict),            // 筛子
	//		CardCount:    int32(iAllCardLen),       // 总共有多少张牌
	//		CardFirst:    int32(r.CurMJIndex),      // 拿牌的位置
	//		EndBlank:     int32(r.EndBlank),        // 结尾空余的牌
	//		CardLeft:     int32(iCardLeft),         // 还剩下多少张牌
	//		CurCardId:    int32(curCardId),         //
	//		CurIndex:     int32(r.CurIndex),        //
	//		CurTime:      int32(r.WaitToolTimeOut), //
	//		ZhuangIndex:  int32(r.BankerIndex),     // UID
	//		StartGame:    int32(iRun),              //
	//		Defeat:       []int32{0, 0, 0, 0},      //
	//		LastPutIndex: int32(r.LastPutIndex),    //
	//		Show:         r.Show,                   //
	//	}

	//		if a == nil {
	//			r.BCMessage(mjgame.MsgID_MSG_ACKBC_Card_Init, ack)
	//		} else {
	//			a.SendMessage(mjgame.MsgID_MSG_ACKBC_Card_Init, ack)
	//		}

	//	//记录战绩
	//	if needRecord {
	//		r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Card_Init, ack)
	//	}
}

//得到房间信息
func (r *RoomPinshi) GetRoomInfo() *mjgame.ACK_Room_Info {
	ack := mjgame.ACK_Room_Info{
		RoomId:     int32(r.RoomId),
		Type:       int32(mjgame.MsgID_GTYPE_Pinshi),
		City:       int32(0),
		Level:      int32(0),
		Rules:      r.Rules.Rules,
		SeatCount:  int32(len(r.Seats)),
		Starting:   r.IsRun,
		RoundCount: int32(r.RoundCount),
		Direction:  int32(r.FengQuan),
		UniqueCode: r.UniqueCode,
		RoomOwner:  strconv.Itoa(r.CreateUserId),
	}
	// 结算时间
	ack.DisbandLeftTime = int64(def.WaitStartTime - (time.Now().Unix() - r.CreateTime))
	if ack.DisbandLeftTime < 0 {
		ack.DisbandLeftTime = 0
	}

	if r.IsRun || r.RoundCount > 0 {
		if r.Rules.MaxTime > 0 {
			ack.LeftTime = int64((r.Rules.MaxTime * 60) - (int(time.Now().Unix()) - r.StartTime))
			if ack.LeftTime < 0 {
				ack.LeftTime = 0
			}
		}
	}

	startUser, _ := r.GetFirstSitSeatInfo()
	if startUser != nil {
		ack.NickName = startUser.NickName
	}

	return &ack
}

func (r *RoomPinshi) GetRecordRoomInfo() interface{} {
	ack := &struct {
		RID        int32
		Type       int32
		Rule       []int32
		RoundCount int32
		direction  int32
		leftTime   int64
	}{
		RID:        int32(r.RoomId),
		Type:       int32(mjgame.MsgID_GTYPE_Pinshi),
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

//是否可以解散
func (r *RoomPinshi) GetDisApproveCount() int {
	var count int
	for _, v := range r.Votes {
		if v == DisApprove {
			count++
		}
	}

	return count
}

// 等待用户操作
func (r *RoomPinshi) WaitPutTool() {

	if !r.IsRun {
		return
	}
	huList, index, opType, ok := r.WaitOptTool.CheckGetCpt()
	if !ok {
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

		//getCard := true           // 摸牌
		if r.WaitOptTool.IsSelf { // 过了自己的暗杠不摸牌
			if r.CurIndex <= 0 {
				r.CurIndex = r.CurIndex + r.Rules.SeatLen - 1
			} else {
				r.CurIndex--
			}
			//getCard = false
		}

		if r.IsKongHu { //处理拉杠胡
			index := r.CurIndex
			r.Show = false
			if index <= 0 {
				r.CurIndex = index - 1 + r.Rules.SeatLen
			} else {
				r.CurIndex = index - 1
			}
			r.Status = rb.WaitPut
			//r.TurnNextPlayer(getCard, false, false)
			r.Seats[index].Message = nil
			r.IsKongHu = false

		}

		//r.TurnNextPlayer(getCard, true, false)
	}

	//r.WaitOptTool.ClearAll()
}

func (r *RoomPinshi) GetMultiWinInfo() ([]int, int) {
	//四人麻将最多三个人胡
	var huIndexes []int
	var huCardId = -1

	for i := 0; i < 3; i++ {
		lastTool := (r.StlCtrl).(*Pinshi_Statement).Get(i)
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

func (r *RoomPinshi) GetMaxIndex(arr []int32) int {
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
func (r *RoomPinshi) RoundCanFinish() bool {
	var count int

	for _, v := range r.Seats {
		if v == nil || v.User == nil {
			//count++
			continue
		} else if v.User.State == def.Offline {
			if v.OfflineTime.Add(def.MaxOfflineTime * time.Minute).Before(time.Now()) {
				if v.OfflineTime.IsZero() == false {
					count++
				}
				break
			}
		}
	}

	if count > 0 {
		return true
	}

	return false
}

// 以局数计算两局之间可否结束
func (r *RoomPinshi) ByRoundCanFinish() bool {

	var count int

	for _, v := range r.Seats {
		if v == nil {
			fmt.Println("v nil 椅子信息不村子")
			//count++
			continue
		} else if v.User == nil || v.User.State == def.Offline {
			if v.OfflineTime.Add(def.MaxOfflineTime * time.Minute).Before(time.Now()) {

				if v.OfflineTime.IsZero() == false {
					fmt.Println("v.OfflineTime :: ", v.OfflineTime)
					fmt.Println("离开时间超过180分钟")
					count++
				}
				break
			}
		}
	}

	if count > 0 {
		return true
	}

	return false
}

func (r *RoomPinshi) GetKickIndex() []int32 {
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
func (r *RoomPinshi) MoveKongList(uIndex int, tIndex int, cid, kongType int) {
	seat := r.Seats[uIndex]
	t, n := st.GetMjTypeNum(cid)

	// 移动第一张牌
	if kongType == 1 { //明杠
		card := &rb.Card{ID: cid, TIndex: tIndex, Status: kongType - 1, Type: t, Num: n}
		r.MoveToList(r.Seats[tIndex].Cards.Out, []*rb.Card{card}, seat.Cards.Kong)
	} else if kongType == 2 || kongType == 3 { //暗杠, 补杠
		card := &rb.Card{ID: cid, TIndex: tIndex, Status: kongType - 1, Type: t, Num: n}
		r.MoveToList(seat.Cards.List, []*rb.Card{card}, seat.Cards.Kong)
	}

	// 移动3张牌
	for i := 0; i < 3; i++ {
		if kongType == 1 || kongType == 2 { //明杠,移动list牌
			card := &rb.Card{ID: cid, Status: kongType - 1, TIndex: tIndex, Type: t, Num: n}
			r.MoveToList(seat.Cards.List, []*rb.Card{card}, seat.Cards.Kong)
		} else if kongType == 3 { //判断是否碰上杠
			card := &rb.Card{ID: cid, Status: kongType - 1, TIndex: tIndex, Type: t, Num: n}
			r.MoveToList(seat.Cards.Peng, []*rb.Card{card}, seat.Cards.Kong)
		}
	}
}

func (r *RoomPinshi) MoveChowList(index int, cards []*rb.Card) {
	seat := r.Seats[index]
	r.MoveToList(r.Seats[r.CurIndex].Cards.Out, []*rb.Card{cards[0]}, seat.Cards.Chow)
	r.MoveToList(seat.Cards.List, []*rb.Card{cards[1], cards[2]}, seat.Cards.Chow)
}

func (r *RoomPinshi) DealKongHu(cardID int) {

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

func (r *RoomPinshi) ClearRoomUserRoomId() {
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
func (r *RoomPinshi) SetInitCards(uid string, cids []string) {
	for _, v := range r.Seats {
		if v.UID == uid && r.CurMJIndex <= 61 && r.CurMJIndex > 0 {
			v.Cards.List = al.New()
			for k, cid := range cids {
				CardId, err := strconv.Atoi(cid)
				if err != nil {
					fmt.Println("******err is ", err)
					continue
				}
				if CardId >= 0 {
					v.Cards.List.Add(rb.GetCardById(CardId))
				}
				if k == 13 {
					r.CurCard = rb.GetCardById(CardId)
					//清空
					r.WaitPut(r.WaitPutTimeOut)
				}
			}
		}
	}
}

//得到房间用户信息
func (r *RoomPinshi) GetRoomUser() *mjgame.ACK_Room_User {
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
					//					user.State = def.Kick
					if user.State == def.Online {
						user.State = def.Kick
					} else if user.State == def.Offline {
						user.State = def.OffKick
					}
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

//判断是否可以吃牌
func (r *RoomPinshi) CheckChow(pCard *rb.Card) {
	if pCard == nil || pCard.Type == rb.F || pCard.Type == rb.H || pCard.Type == rb.J {
		//fmt.Println("检查吃牌错误  ---", pCard)
		return
	}
	index := (r.CurIndex + 1) % r.Rules.SeatLen

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

//设置不能吃的牌
func (r *RoomPinshi) SetPassChowCards(pCard *rb.Card) {
	index := r.CurIndex % r.Rules.SeatLen
	if pCard.Type == rb.F || pCard.Type == rb.H || pCard.Type == rb.J {
		return
	}
	if r.Seats[index].User == nil { //玩家断线则不操作,跳过
		fmt.Println("SetPassChowCards failed,玩家断线")
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

	if n1 && n2 || n2 && n3 || n3 && n4 {
		r.Seats[index].ChowCardIDs = append(r.Seats[index].ChowCardIDs, pCard.ID)
		r.RoomBase.RoomRecord += "检测(" + r.Seats[index].User.NickName + ") 吃" + pCard.MSG + "\r\n"
	}
}
