/*
 * +-----------------------------------------------
 * | ningbo_handler.go
 * +-----------------------------------------------
 * | Version: 1.0
 * +-----------------------------------------------
 * | Context: 宁波麻将路由处理
 * +-----------------------------------------------
 */
package srddz

import (
	//	al "PZ_GameServer/common/util/arrayList"
	"PZ_GameServer/common/util"
	"PZ_GameServer/model"
	"PZ_GameServer/protocol/def"
	"PZ_GameServer/protocol/pb"
	"PZ_GameServer/server/game/common"
	"PZ_GameServer/server/game/error"
	px "PZ_GameServer/server/game/room/srddz/paixingLogic"
	rb "PZ_GameServer/server/game/roombase"
	st "PZ_GameServer/server/game/statement"
	"PZ_GameServer/server/user"
	us "PZ_GameServer/server/user"
	"fmt"
	"strconv"
	"time"
)

// 创建房间
func (r *RoomSrddz) Create(rid int, t int32, user *user.User, rule *rb.RoomRule) {
	r.RoomBase.Create(rid, t, user.ID, rule)
	r.QuitSitUsers = make(map[int]rb.SeatBase)

	if r.RoomBase.CheckRule(rule.Rules, int32(Jiaofen)) {
		r.JiaoType = Jiaofen //清10混6
	} else {
		r.JiaoType = Qiangdizhu //清12混8
	}

	user.RoomId = rid
	user.GameType = IFCSrddzType

	go r.TimeTicker() //定时器
	r.StlCtrl = GetSrddzStatement(
		t,
		500,
		[]string{"", "", "", ""},
		r,
	)
}

// 进入房间
func (r *RoomSrddz) IntoUser(user *user.User) {

	index := r.GetSeatIndexById(user.ID)
	if r.IsRun || r.RoundCount > 0 { //游戏开始了之后不让其他玩家进入
		if index < 0 {
			user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrMessageIsRun)
			return
		}
	}

	idx := 0
	for _, seat := range r.Seats {
		if seat.User != nil && seat.User.ID != user.ID { //不是自己已经满员了
			idx++
		}
	}

	if idx >= r.Rules.SeatLen {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrPlayersHasFull)
		return
	}

	if r.RoundCount == 0 && r.IsRun == false {
		//房卡不够不让进
		if r.Rules.PayType == 4 {
			if user.Diamond < r.Rules.Play_NeedDiamond {
				user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrDiamondNotEnough)
				return
			}
		} else if r.Rules.PayType == 1 {
			if user.ID == r.CreateUserId {
				if user.Diamond < r.Rules.Play_NeedDiamond {
					user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrDiamondNotEnough)
					return
				}
			}
		}
	}

	user.GameType = IFCSrddzType
	r.IntoRoom(user)

	if index >= 0 {
		message := r.Seats[index].Message
		if message != nil {
			user.SendMessage(message.Id, message.Content)
		}
	}

	if index >= 0 {
		if r.Seats[index].State == int(mjgame.StateID_GameState_Total) {
			if r.RoundResult != nil {
				user.SendMessage(mjgame.MsgID_MSG_ACKBC_Sddz_Total, r.RoundResult)
			}
		}
	}
}

//开始游戏
func (r *RoomSrddz) Start(user *user.User) {
	r.StateMutex.Lock()
	defer r.StateMutex.Unlock()
	if r.IsRun { // 不要放到 CheckCanStart
		fmt.Println("游戏已经开始, 不能重复开始")
		return
	}

	flag, err := r.CheckCanStart()
	if !flag {
		r.IsRun = false
		r.BCMessage(mjgame.MsgID_MSG_ACK_Error, err)
		return
	}

	r.IsRun = true

	r.Init() //初始化数据

	r.InitRandPockAllCard() //洗牌

	r.StlCtrl = GetSrddzStatement( // 结算控制器
		r.Type,
		r.Rules.BaseScore,
		[]string{
			r.Seats[0].UID,
			r.Seats[1].UID,
			r.Seats[2].UID,
			r.Seats[3].UID,
		},
		r,
	)

	//首次开局
	if r.RoundCount == 0 {
		for _, seat := range r.Seats {
			if r.Rules.PayType == 1 { //只扣房主的
				if seat.User.ID == r.CreateUserId {
					common.AddDiamondLog(seat.User, int(r.Type), -r.Rules.Play_NeedDiamond)
				}
			} else if r.Rules.PayType == 4 {
				common.AddDiamondLog(seat.User, int(r.Type), -r.Rules.Play_NeedDiamond)
			}

		}
	}

	// Redis
	for i := 0; i < len(r.Seats); i++ {
		rb.Redis_AddPlayingUser(r.RoomId, r.Seats[i].UID)
	}

	r.RoundTime = time.Now()
	r.State = rb.Dealt

	r.NewBattleRecord()
	//r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACK_RoomInfo, r.GetRecordRoomInfo())

	roomUser := r.GetRoomUser()

	for _, u := range roomUser.Users {
		u.Ip = ""
		u.RoomId = 0
		u.Icon = ""
		u.Name = ""
	}
	//r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACK_Room_User, roomUser)

	go r.Process()

	//	index := r.GetSeatIndexById(user.ID)
	//	r.AddTool(st.T_Start, index, -1, []int{})
}

// 处理器
func (r *RoomSrddz) Process() {

	if !r.IsRun {
		return
	}
	r.State = -1
	r.InitRound()    //初始化每局
	r.InitUserPock() //创建13张牌

	cards := r.GetListArray(r.Seats[r.StartIdx].Cards.List)
	if px.GetPockType(int(cards[5].Cid)) == px.POCK_WANG {
		r.IsMingW = true
	} else {
		r.IsMingW = false
	}
	r.Beishu = 1
	if r.IsHuang {
		r.Beishu = r.Beishu * 2
	}
	if r.IsMingW {
		r.Beishu = r.Beishu * 2
	}

	r.SendGameInfo(nil, true)
	//参与用户
	for _, v := range r.Seats {
		r.SendSeatCard(v.User.ID)
	}
	//旁观用户
	r.SendSeatCard(-1)
	//底牌
	r.Dipai = append(r.Dipai, r.AllCards[r.StartIndex].ID, r.AllCards[r.StartIndex+1].ID,
		r.AllCards[r.StartIndex+2].ID, r.AllCards[r.StartIndex+3].ID, r.AllCards[r.StartIndex+4].ID,
		r.AllCards[r.StartIndex+5].ID, r.AllCards[r.StartIndex+6].ID, r.AllCards[r.StartIndex+7].ID)
	//底牌
	r.Stage = Stage_jiaofen
	//r.RecordSeatCard()
	r.CurIndex = r.StartIdx - 1 // 方便TurnNextPlayer统一
	r.WaitTimeCount = 1
	r.TurnNextPlayer()

}

//

// 起立
func (r *RoomSrddz) StandUp(arg *mjgame.StandUp, user *user.User) {
	r.StateMutex.Lock()
	defer r.StateMutex.Unlock()
	index := r.GetSeatIndexById(user.ID)
	if r.IsRun || index < 0 || r.RoundCount > 0 {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrCanNotStandUp)
		return
	}

	if index >= 0 {
		r.QuitSitUsers[user.ID] = *r.Seats[index]
	}
	seat := r.Seats[index]

	seat.State = int(mjgame.StateID_UserState_Stand)
	seat.Index = -1
	seat.CreateTimeStamp = 0

	r.WatchSeats = append(r.WatchSeats, user)

	startUser, _ := r.GetFirstSitSeatInfo()
	var NickTemp string
	if startUser != nil {
		NickTemp = startUser.NickName
	}
	ackBCStandUp := &mjgame.ACKBC_Standup{
		Uid:   int32(user.ID),
		Index: -1,
		//NickName: user.NickName,
		NickName: NickTemp,
	}
	if startUser != nil {
		ackBCStandUp.NickName = startUser.NickName
	}
	r.AddTool(st.T_StandUp, int(index), -1, []int{})

	r.Seats[index] = &rb.SeatBase{}
	r.Seats[index].OfflineTime = time.Now()

	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Standup, ackBCStandUp) // 起立消息
}

//坐下
func (r *RoomSrddz) SitDown(user *user.User, arg *mjgame.SitDown) {
	if r.Seats[arg.Index].User != nil {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrCurSeatHasUsed)
		return
	}

	if r.RoundCount == 0 {
		if r.Rules.PayType == 4 {
			if user.Diamond < r.Rules.Play_NeedDiamond {
				user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrDiamondNotEnough)
				return
			}
		} else if r.Rules.PayType == 1 {
			if user.ID == r.CreateUserId {
				if user.Diamond < r.Rules.Play_NeedDiamond {
					user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrDiamondNotEnough)
					return
				}
			}
		}
	}

	//从观察者转为参与者
	wIndex := -1
	for i, v := range r.WatchSeats {
		if v.ID == user.ID {
			wIndex = i
			break
		}
	}

	if wIndex >= 0 {
		r.Seats[arg.Index].User = r.WatchSeats[wIndex]
		r.Seats[arg.Index].Accumulation = &rb.Accumulation{}
		r.WatchSeats = append(r.WatchSeats[:wIndex], r.WatchSeats[wIndex+1:]...)
	} else {
		index := r.GetSeatIndexById(user.ID)
		if index >= 0 {
			r.QuitSitUsers[user.ID] = *r.Seats[index]
			r.Seats[index] = &rb.SeatBase{}
		}
		r.Seats[arg.Index].User = user
	}

	r.Seats[arg.Index].UID = strconv.Itoa(user.ID)
	r.Seats[arg.Index].State = int(mjgame.StateID_UserState_Sit)
	r.Seats[arg.Index].Index, r.Seats[arg.Index].CreateTimeStamp = int(arg.Index), time.Now().Unix()

	startUser, _ := r.GetFirstSitSeatInfo() // fix bug  20170629  修复为空的的问题
	startUserNickName := ""
	if startUser != nil {
		startUserNickName = startUser.NickName
	}

	ackBCSitDown := &mjgame.ACKBC_Sitdown{
		Uid:      int32(user.ID),
		Index:    arg.Index,
		NickName: startUserNickName,
	}

	if v, ok := r.QuitSitUsers[user.ID]; ok {
		if v.Accumulation != nil {
			r.Seats[arg.Index].Accumulation = v.Accumulation
			ackBCSitDown.Score = v.Accumulation.Score
		}
	}

	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Sitdown, ackBCSitDown) //广播已坐下
}

//退出
func (r *RoomSrddz) ExitUser(user *user.User) {
	index := r.GetSeatIndexById(user.ID)

	if r.IsRun { //在玩的时候坐下的玩家不让走
		if index >= 0 { //|| r.RoundToatlFinish == false
			user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrGameHasRunning)
			return
		}
	} else if r.RoundCount > 0 { //不在玩的时候在两局之间也不让走
		if index >= 0 {
			user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrGameHasRunning)
			return
		}
	}
	ackBCExitRoom := &mjgame.ACKBC_Exit_Room{
		Uid: strconv.Itoa(user.ID),
	}

	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Exit_Room, ackBCExitRoom) //退出房间

	if index >= 0 {
		r.QuitSitUsers[user.ID] = *r.Seats[index]
	}

	//座位数据清空
	if index >= 0 {
		//offlineTime := time.Now()

		r.Seats[index] = &rb.SeatBase{}
		r.Seats[index].Message = nil
		r.Seats[index].State = int(mjgame.StateID_UserState_Normal)
		//r.Seats[index].OfflineTime = offlineTime

		//fmt.Println("掉线时间 ：：", offlineTime)

	} else {
		for i, v := range r.WatchSeats {
			if v.User.ID == user.ID {
				r.WatchSeats = append(r.WatchSeats[:i], r.WatchSeats[i+1:]...)
				break
			}
		}
	}

	r.AddTool(st.T_Exit, index, -1, []int{})

	//	if r.IsEmpty() {
	//		rb.ChanRoom <- r.RoomId
	//	}
	rb.Redis_RemoveUser(strconv.Itoa(r.RoomId), strconv.Itoa(user.ID))

	user.RoomId = 0
	user.GameType = nil
	fmt.Println("exit user", r.RoomId)
}

//准备
func (r *RoomSrddz) Ready(user *user.User) {
	index := r.GetSeatIndexById(user.ID)
	if index < 0 || r.IsRun {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrInvalidParam)
		return
	}

	r.Seats[index].State = int(mjgame.StateID_UserState_Ready)

	var readyCount int
	for _, seat := range r.Seats {
		if seat.State == int(mjgame.StateID_UserState_Ready) {
			readyCount++
		}
	}

	if readyCount == r.Rules.SeatLen {
		if r.RoundCount == 0 {
			startUser, index := r.GetFirstSitSeatInfo()
			if startUser != nil {
				notifyStartGame := &mjgame.NotifyStartGame{
					Uid: strconv.Itoa(startUser.ID),
				}
				startUser.SendMessage(mjgame.MsgID_MSG_NOTIFY_START_GAME, notifyStartGame)
				r.Seats[index].Message = &rb.Message{
					Id:      mjgame.MsgID_MSG_NOTIFY_START_GAME,
					Content: notifyStartGame,
				}
			}
		} else {
			r.Start(user)
		}
	}

	ackReady := &mjgame.ACKBC_Ready{
		ReadyCount: int32(readyCount),
		UID:        strconv.Itoa(user.ID),
		MSG:        "",
	}
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Ready, ackReady)

	index = r.GetSeatIndexById(user.ID)
	if index >= 0 {
		r.Seats[index].Message = &rb.Message{
			Id:      mjgame.MsgID_MSG_ACKBC_Ready,
			Content: ackReady,
		}
	}
}

// 出牌
func (r *RoomSrddz) PutCard(user *user.User, arg *mjgame.Put_Card) {

	index := r.GetSeatIndexById(user.ID)
	putCid := int(arg.Cid)

	if !r.IsRun || index < 0 || putCid < 0 || (r.Status == rb.WaitTool && r.CurIndex != index) {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrInvalidParam)
		return
	}

	if !r.RoomBase.MToolChecker.CheckTool(index, int(mjgame.MsgID_MSG_ACKBC_PutCard)) {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrInvalidParam)
		return
	}

	if !r.RoomBase.CheckCanPut(index, putCid) {
		fmt.Println("CheckCanPut xs出牌错误 ", putCid, r.RoomId)
		return
	}
	if r.WinUserCount == 0 {

		seat := r.Seats[index]
		card := seat.Cards.GetCardByCardId(putCid)

		//fmt.Println("----------------------------------------------> ", r.Seats[index].User.NickName, index, r.Seats[index].Cards.List.Count)
		if card != nil {
			r.MoveToList(seat.Cards.List, []*rb.Card{card}, seat.Cards.Out)
			seat.Ting = int(arg.Ting)
			putCard := &mjgame.ACKBC_PutCard{
				Cid:   int32(putCid),
				Index: int32(index),
			}

			r.Seats[index].PengCardIDs = []int{}
			r.Seats[r.CurIndex].ChowCardIDs = []int{} //过了一轮要清空
			r.BCMessage(mjgame.MsgID_MSG_ACKBC_PutCard, putCard)
			//记录回放
			rec := []interface{}{putCard.Index, putCard.Cid}
			r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_PutCard, &rec)

			//r.RoomBase.RoomRecord += "出牌(" + r.Seats[index].User.NickName + ") " + card.MSG + "\r\n"
			r.Seats[index].IsPutCard = true //已经出过牌了
			//设置不能吃的牌
			r.SetPassChowCards(card)
			//设置不能胡的牌
			if r.Seats[r.CurIndex].IsCanWin {
				//此时CurCard还是摸得那张牌 一定确保顺序 下面CurCard是出的牌
				r.Seats[r.CurIndex].HuCardIDs = append(r.Seats[r.CurIndex].HuCardIDs, r.CurCard.ID)
				//过胡的牌不能碰
				for ii := 0; ii < 4; ii++ {
					r.Seats[ii].PengCardIDs = append(r.Seats[ii].PengCardIDs, r.CurCard.ID)
				}
			}

			r.Seats[index].LastCardID = -1
			r.CurCard = card
			r.LastPutIndex = index
			r.AddTool(st.T_Put, r.CurIndex, -1, []int{card.ID})
			r.Show = true
			for i := 0; i < 4; i++ {
				r.Seats[i].Message = nil
			}

			//r.RoomBase.MToolChecker.SetCptTool(index, int(mjgame.MsgID_MSG_ACKBC_PutCard), []int{putCid}, r.Seats[index].User.NickName)

			r.StartWaitTool(card) // 开始判断其他玩家的操作

		} else {
			user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrCardNotExist)
		}
	} else {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrHasNotTurnPlay)
	}
}

//叫分
func (r *RoomSrddz) Jiaofen(user *user.User, fen int) {
	idx := r.GetSeatIndexById(user.ID)
	if idx != r.CurIndex { //不是这个家伙
		return
	}
	if r.Stage != Stage_jiaofen { //不是叫分阶段
		return
	}

	if r.Seats[idx].HaveJiao { //这个家伙叫过牌了
		return
	}
	r.Seats[idx].HaveJiao = true
	r.Seats[idx].JiaoFen = fen

	ack := mjgame.ACKBC_Sddz_Jiaofen{
		Uid: int32(user.ID),
		Fen: int32(fen),
	}
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Sddz_Jiaofen, &ack)

	if fen > r.Difen { //这个家伙叫地主了
		r.CurDizhuIdx = idx
		r.Difen = fen
		if fen == 3 { //玩家叫了三分直接成为地主
			r.DizhuPos = idx
			if idx == r.StartIdx {
				r.IsTouliao = true
				//r.Beishu = r.Beishu * 4 / 3  头撂和报到单独计算
				r.SendGameInfo(nil, true)
			} else {
				r.IsTouliao = false
			}
		}
	}

	allHave := true

	for i := 0; i < r.Rules.SeatLen; i++ {
		if r.Seats[i].HaveJiao == false {
			allHave = false
			break
		}
	}

	if allHave == true {
		if r.Difen == 0 { //大家都不叫
			//			r.Difen = 1
			//			r.DizhuPos = r.StartIdx
			r.IsHuang = true
			for i := 0; i < r.Rules.SeatLen; i++ {
				r.Seats[i].HaveJiao = false
				r.Seats[i].JiaoFen = 0
				r.Seats[i].OutZha = 0
			}

			r.RoomBase.Init()
			r.InitRandPockAllCard() //洗牌
			r.State = -1
			r.InitRound()    //初始化每局
			r.InitUserPock() //创建13张牌

			cards := r.GetListArray(r.Seats[r.StartIdx].Cards.List)
			if px.GetPockType(int(cards[5].Cid)) == px.POCK_WANG {
				r.IsMingW = true
			} else {
				r.IsMingW = false
			}
			r.Beishu = 1
			if r.IsHuang {
				r.Beishu = r.Beishu * 2
			}
			if r.IsMingW {
				r.Beishu = r.Beishu * 2
			}

			r.SendGameInfo(nil, true)
			//参与用户
			for _, v := range r.Seats {
				r.SendSeatCard(v.User.ID)
			}
			//旁观用户
			r.SendSeatCard(-1)
			//底牌
			r.Dipai = []int{}
			r.Dipai = append(r.Dipai, r.AllCards[r.StartIndex].ID, r.AllCards[r.StartIndex+1].ID,
				r.AllCards[r.StartIndex+2].ID, r.AllCards[r.StartIndex+3].ID, r.AllCards[r.StartIndex+4].ID,
				r.AllCards[r.StartIndex+5].ID, r.AllCards[r.StartIndex+6].ID, r.AllCards[r.StartIndex+7].ID)
			//底牌
			r.Stage = Stage_jiaofen
			//r.RecordSeatCard()
			r.CurIndex = r.StartIdx - 1 // 方便TurnNextPlayer统一
			r.WaitTimeCount = 1
			r.TurnNextPlayer()
			return
		} else {
			r.DizhuPos = r.CurDizhuIdx //地主是最后一个叫牌的玩家
		}
	}

	if r.DizhuPos == -1 {
		r.TurnNextPlayer()
	} else { //确认了地主 进入到加倍阶段
		for i := 0; i < r.Rules.SeatLen; i++ {
			r.Seats[i].Message = nil
		}
		r.Stage = Stage_dapai
		uid, err := strconv.Atoi(r.Seats[r.DizhuPos].UID)
		if err != nil {
			return
		}

		//dipai := []int32{int32(r.Dipai[0]), int32(r.Dipai[1]), int32(r.Dipai[2])}
		kdipai := make([]int32, 8)

		dipai := []int32{int32(r.Dipai[0]), int32(r.Dipai[1]), int32(r.Dipai[2]),
			int32(r.Dipai[3]), int32(r.Dipai[4]), int32(r.Dipai[5]), int32(r.Dipai[6]), int32(r.Dipai[7])}

		r.Seats[r.DizhuPos].Cards.List.Add(&r.AllCards[r.StartIndex])
		r.Seats[r.DizhuPos].Cards.List.Add(&r.AllCards[r.StartIndex+1])
		r.Seats[r.DizhuPos].Cards.List.Add(&r.AllCards[r.StartIndex+2])
		r.Seats[r.DizhuPos].Cards.List.Add(&r.AllCards[r.StartIndex+3])
		r.Seats[r.DizhuPos].Cards.List.Add(&r.AllCards[r.StartIndex+4])
		r.Seats[r.DizhuPos].Cards.List.Add(&r.AllCards[r.StartIndex+5])
		r.Seats[r.DizhuPos].Cards.List.Add(&r.AllCards[r.StartIndex+6])
		r.Seats[r.DizhuPos].Cards.List.Add(&r.AllCards[r.StartIndex+7])

		ack_dizhu := mjgame.ACKBC_Sddz_Dizhu{
			Uid: int32(uid),
			Fen: int32(r.Difen),
		}
		//r.BCMessage(mjgame.MsgID_MSG_ACKBC_Sddz_Dizhu, &ack_dizhu)

		for _, v := range r.Seats {
			if v != nil && v.User != nil && v.User.Conn != nil {
				if v.Index == r.DizhuPos {
					ack_dizhu.Dipai = dipai
				} else {
					ack_dizhu.Dipai = kdipai
				}
				v.User.SendMessage(mjgame.MsgID_MSG_ACKBC_Sddz_Dizhu, &ack_dizhu)
			}
		}

		ack_dizhu.Dipai = kdipai
		us.BCMessage(mjgame.MsgID_MSG_ACKBC_Sddz_Dizhu, &ack_dizhu, r.WatchSeats)

		//		for _, v := range r.WatchSeats {
		//			if v != nil && v.User != nil && v.Conn != nil {
		//				ack_dizhu.Dipai = kdipai
		//				v.user.SendMessage(mjgame.MsgID_MSG_ACKBC_Sddz_Dizhu, &ack_dizhu)
		//			}
		//		}
		r.CurIndex = r.DizhuPos - 1
		r.TurnNextPlayer()

	}

}

//加倍
func (r *RoomSrddz) Jiabei(user *user.User, jiabei bool) {
	if r.Stage != Stage_jiabei { //不是加倍阶段
		return
	}

	idx := r.GetSeatIndexById(user.ID)

	if r.Seats[idx].HaveJiabei { //这个家伙叫过牌了
		return
	}
	r.Seats[idx].HaveJiabei = true
	r.Seats[idx].IsJiabei = jiabei

	ack := mjgame.ACKBC_Sddz_Jiabei{
		Uid:    int32(user.ID),
		Jiabei: r.Seats[idx].IsJiabei,
	}
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Sddz_Jiabei, &ack)

	allHave := true

	for i := 0; i < r.Rules.SeatLen; i++ {
		if r.Seats[i].HaveJiabei == false {
			allHave = false
			break
		}
	}

	if allHave == true {
		r.Stage = Stage_dapai
		r.CurIndex = r.DizhuPos - 1
		r.TurnNextPlayer()
	}

}

func (r *RoomSrddz) Mingpai(user *user.User) {
	if r.Stage != Stage_dapai { //只有打牌阶段可以明牌
		return
	}

	if r.CurOutputIdx != -1 { //出过牌了不可以明牌
		return
	}
	index := r.GetSeatIndexById(user.ID)
	if index != r.DizhuPos { //只有地主可以明牌
		return
	}

	//明牌通知 1509
	//message ACKBC_Sddz_Mingpai{
	//   int32  Uid = 1;		//明牌的玩家（特指地主可以明牌）
	//   bool Mingpai = 2;    //是否明牌
	//   repeated int32 Cards = 3;//明牌的玩家手里的牌
	//}

	r.Beishu = r.Beishu * 2
	r.Seats[index].IsMing = true
	//r.SendGameInfo(nil, true)

	arrMj := r.Seats[index].Cards.List
	cards := make([]int32, arrMj.Length())
	for i := 0; i < arrMj.Length(); i++ {

		cards[i] = int32((*arrMj.Index(i)).(*rb.Card).ID) //
	}

	ack := mjgame.ACKBC_Sddz_Mingpai{
		Uid:     int32(user.ID),
		Mingpai: true,
		Cards:   cards,
	}
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Sddz_Mingpai, &ack)
}

//出牌
func (r *RoomSrddz) Chupai(user *user.User, arg *mjgame.Sddz_Chupai) {
	if r.Stage != Stage_dapai { //不是打牌阶段不可以打牌
		return
	}
	index := r.GetSeatIndexById(user.ID)
	if index != r.CurIndex { //不是当前这个人出牌
		return
	}
	cards := make([]int, len(arg.Cards))
	for i, v := range arg.Cards {
		cards[i] = int(v)
	}

	r.CurOutputCards = arg
	r.CurOutputIdx = index
	seat := r.Seats[index]
	seat.LastOpt = rb.Last_opt_chupai
	seat.LastOptParam = r.CurOutputCards.Cards

	if arg.Type >= 8 {
		seat.OutZha += 1
	}

	//r.BaodaoCards.indexof()
	if index == r.DizhuPos {
		for _, va := range arg.Cards {
			for i := 0; i < len(r.BaodaoCards); i++ {
				if r.BaodaoCards[i] == va {
					r.BaodaoCards = append(r.BaodaoCards[:i], r.BaodaoCards[i+1:]...)
					i--
				}
			}
		}
	}

	seat.OutputNum += 1 //出牌次数
	//tempArr := al.New() // 打出的牌
	for _, v := range cards {
		card := seat.Cards.GetCardByCardId(v)
		if card != nil {
			r.MoveToList(seat.Cards.List, []*rb.Card{card}, seat.Cards.Out)
		}
	}
	//seat.Cards.Out.Add(tempArr) //这个可能需要调整
	r.Beishu = r.Beishu * px.GetPockTypeBeishu(int(arg.Type))

	ack := mjgame.ACKBC_Sddz_Chupai{
		Uid:    int32(user.ID),
		Type:   arg.Type,
		Cards:  arg.Cards,
		Beishu: int32(r.Beishu),
	}
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Sddz_Chupai, &ack)

	if seat.Cards.List.Length() > 0 {
		r.TurnNextPlayer()
	} else { //本局结束
		r.Sddz_end()
	}

}

//三人斗地主不出
func (r *RoomSrddz) Sddz_Pass(user *user.User) {
	if r.Stage != Stage_dapai { //不是打牌阶段不可以打牌
		return
	}
	index := r.GetSeatIndexById(user.ID)
	if index != r.CurIndex { //不是当前这个人出牌
		return
	}

	seat := r.Seats[index]
	seat.LastOpt = rb.Last_opt_buchu
	seat.LastOptParam = nil

	ack := mjgame.ACKBC_Sddz_Pass{
		Uid: int32(user.ID),
	}
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Sddz_Pass, &ack)

	r.TurnNextPlayer()
}

//报到
func (r *RoomSrddz) Baodao(user *user.User, arg *mjgame.Srddz_Baodao) {
	if r.Stage != Stage_dapai { //只有打牌阶段可以报到
		return
	}

	if r.CurOutputIdx != -1 { //出过牌了不可以报到了
		return
	}
	index := r.GetSeatIndexById(user.ID)
	if index != r.DizhuPos { //只有地主可以报到
		return
	}

	if r.IsSeleBaodao == true { //已经选择过了报到
		return
	}
	r.IsSeleBaodao = true
	//检查报到的牌玩家手里有没有
	seat := r.Seats[index]
	cards := make([]int, len(arg.Cards))
	for i, v := range arg.Cards {
		cards[i] = int(v)
	}

	for _, v := range cards {
		card := seat.Cards.GetCardByCardId(v)
		if card == nil { //报到的牌玩家没有
			return
		}
	}

	r.BaodaoNum = int(arg.BaodaoNum)
	if r.BaodaoNum == 0 {
		r.BaodaoState = 0
	} else if r.BaodaoNum == 1 {
		r.BaodaoState = 1
	} else if r.BaodaoNum > 1 {
		r.BaodaoState = 2
	}

	r.BaodaoCards = arg.Cards

	//开始报到
	ack := mjgame.ACKBC_Srddz_Baodao{
		Uid:       int32(user.ID),
		BaodaoNum: arg.BaodaoNum,
		Cards:     arg.Cards,
	}
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Srddz_Baodao, &ack)
	//fmt.Println("baodaode pai :: ", r.BaodaoCards)
}

//直接赢
func (r *RoomSrddz) StrictWin(user *user.User, arg *mjgame.Srddz_StrictWin) {
	if r.Stage != Stage_dapai { //只有打牌阶段可以直接赢
		return
	}

	if r.CurOutputIdx != -1 { //出过牌了不可以直接赢
		return
	}
	index := r.GetSeatIndexById(user.ID)
	if index != r.DizhuPos { //只有地主可以直接赢
		return
	}
	if r.IsSeleBaodao == false { //没有报到不可以直接赢
		return
	}
	if r.IsSeleStrictWin == true { //已经选择了不可以继续选择
		return
	}
	r.IsSeleStrictWin = true

	r.IsStrictWin = arg.IsStrictWin

	//开始报到
	ack := mjgame.ACKBC_Srddz_StrictWin{
		Uid:         int32(user.ID),
		IsStrictWin: arg.IsStrictWin,
	}
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Srddz_StrictWin, &ack)

	if r.IsStrictWin == true { //直接赢 本局结束
		r.Sddz_end()
	}

}

//三人斗地主游戏结束
func (r *RoomSrddz) Sddz_end() {
	//	var targetIndex int
	//	var seats []int32

	index := r.CurIndex
	uid, _ := strconv.Atoi(r.Seats[index].UID)

	if !r.IsRun || index < 0 {
		return
	}

	allSeatCards := r.GetAllSeatCards()
	//dipai := []int32{int32(r.Dipai[0]), int32(r.Dipai[1]), int32(r.Dipai[2])}
	dipai := []int32{int32(r.Dipai[0]), int32(r.Dipai[1]), int32(r.Dipai[2]),
		int32(r.Dipai[3]), int32(r.Dipai[4]), int32(r.Dipai[5]), int32(r.Dipai[6]), int32(r.Dipai[7])}
	end := mjgame.ACKBC_Sddz_End{
		Uid:   int32(uid),
		Cards: allSeatCards,
		Dipai: dipai,
	}
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Sddz_End, &end)

	r.RoundToatlFinish = false

	r.Sddz_Total()
	r.RoundToatlFinish = true

	for k, _ := range r.Seats {
		r.Seats[k].Message = &rb.Message{
			Id:      mjgame.MsgID_MSG_ACKBC_Sddz_End,
			Content: &end,
		}
	}
}

//三人斗地主结算
func (r *RoomSrddz) Sddz_Total() {

	if r.RoomBase.RoundTotaled {
		fmt.Println("已经结算过.", r.RoomId)
		return
	}
	r.RoomBase.RoundTotaled = true

	if r.RoundCount == 0 {
		room := &model.Room{
			Type:         int(mjgame.MsgID_GTYPE_SirenDizhu),
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
	r.IsRun = false
	flag, _ := r.CheckCanStart()

	//更新用户状态(断线重连)
	for _, seat := range r.Seats {
		seat.State = int(mjgame.StateID_GameState_Total)
	}

	//开始结算算分
	//	Seat
	//	Difen
	//	Beishu
	//	Score
	//	TotalScore

	list := make([]*mjgame.SddzPerTotal, r.Rules.SeatLen)

	//isDizhuMing := r.Seats[r.DizhuPos].IsMing
	//isDizhuJiabei := r.Seats[r.DizhuPos].IsJiabei
	//isDizhuFanchun := r.Seats[r.DizhuPos].OutputNum == 1

	var notDizhuTotalBeishu int32 = 0
	var notDizhuTotalScore int32 = 0

	if r.CurIndex == r.DizhuPos { //地主胜利了
		for i, seat := range r.Seats {
			perCal := &mjgame.SddzPerTotal{}
			perCal.Seat = int32(i)
			//perCal.Difen = int32(r.Difen)

			if r.BaodaoState != 0 { //有报到
				if r.IsStrictWin == true { //直接赢
					perCal.Difen = int32(r.Difen)
				} else { //正常打
					if r.IsTouliao {
						perCal.Difen = int32(r.Difen) * 4 / 3
					} else {
						perCal.Difen = int32(r.Difen)
					}
				}
			} else { //没有报到
				if r.IsTouliao {
					perCal.Difen = int32(r.Difen) * 4 / 3
				} else {
					perCal.Difen = int32(r.Difen)
				}
			}

			//perCal.Fanchun = false //地主胜利了，都不是反春
			if i != r.DizhuPos { //不是地主

				perCal.Beishu = int32(r.Beishu)
				if r.BaodaoState != 0 { //有报到
					if r.IsStrictWin == true { //直接赢
						perCal.Difen = int32(r.Difen)
						perCal.Beishu = perCal.Beishu * int32(r.BaodaoNum)
					} else if r.IsStrictWin == false { //正常打
						if r.IsTouliao {
							perCal.Difen = int32(r.Difen) * 4 / 3
						} else {
							perCal.Difen = int32(r.Difen)
						}
						perCal.Beishu = perCal.Beishu * int32(r.BaodaoNum) * 2
					}
				} else { //没有报到
					if r.IsTouliao {
						perCal.Difen = int32(r.Difen) * 4 / 3
					} else {
						perCal.Difen = int32(r.Difen)
					}
				}

				perCal.Score = int32(-int32(perCal.Difen) * perCal.Beishu)

				if r.CostType == rb.Jinbi {
					seat.User.Coin += int(perCal.Score)
					if seat.User.Coin <= 0 {
						seat.User.Coin = rb.Jiuji_coin
						seat.User.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrBuchongCoin)
					}
					model.GetUserModel().Save(seat.User.User)
					perCal.TotalScore = int32(seat.User.Coin)
				} else {
					perCal.TotalScore = seat.Accumulation.Score + perCal.Score
				}

				//				perCal.TotalScore = seat.Accumulation.Score + perCal.Score

				notDizhuTotalBeishu += perCal.Beishu
				notDizhuTotalScore += perCal.Score
			} else {
				//perCal.Chun = false //地主不会是春天 只会是反春
			}

			list[i] = perCal
		}

	} else { //地主输了
		for i, seat := range r.Seats {
			perCal := &mjgame.SddzPerTotal{}
			perCal.Seat = int32(i)
			perCal.Difen = int32(r.Difen)
			//perCal.Chun = false  //玩家胜利了大家都不是春天
			if i != r.DizhuPos { //不是地主
				perCal.Beishu = int32(r.Beishu)

				if r.BaodaoState != 0 { //有报到
					perCal.Score = 0
				} else {
					perCal.Score = int32(int32(r.Difen) * perCal.Beishu)
				}

				if r.CostType == rb.Jinbi {
					seat.User.Coin += int(perCal.Score)
					if seat.User.Coin <= 0 {
						seat.User.Coin = rb.Jiuji_coin
						seat.User.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrBuchongCoin)
					}
					model.GetUserModel().Save(seat.User.User)
					perCal.TotalScore = int32(seat.User.Coin)
				} else {
					perCal.TotalScore = seat.Accumulation.Score + perCal.Score
				}

				//				perCal.TotalScore = seat.Accumulation.Score + perCal.Score
				//perCal.Fanchun = false //玩家不会反春
				notDizhuTotalBeishu += perCal.Beishu
				notDizhuTotalScore += perCal.Score
			} else { //这个是地主
				//perCal.Fanchun = isDizhuFanchun
			}
			list[i] = perCal
		}

	}
	list[r.DizhuPos].Beishu = notDizhuTotalBeishu
	list[r.DizhuPos].Score = -notDizhuTotalScore

	if r.CostType == rb.Jinbi {
		r.Seats[r.DizhuPos].User.Coin += int(list[r.DizhuPos].Score)

		if r.Seats[r.DizhuPos].User.Coin <= 0 {
			r.Seats[r.DizhuPos].User.Coin = rb.Jiuji_coin
			r.Seats[r.DizhuPos].User.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrBuchongCoin)
		}
		model.GetUserModel().Save(r.Seats[r.DizhuPos].User.User)
		list[r.DizhuPos].TotalScore = int32(r.Seats[r.DizhuPos].User.Coin)
	} else {
		list[r.DizhuPos].TotalScore = r.Seats[r.DizhuPos].Accumulation.Score + list[r.DizhuPos].Score
	}

	//	list[r.DizhuPos].TotalScore = r.Seats[r.DizhuPos].Accumulation.Score + list[r.DizhuPos].Score

	var scores = make([]int32, r.Rules.SeatLen)
	var ackTotal *mjgame.ACKBC_Sddz_Total

	for i := 0; i < r.Rules.SeatLen; i++ {
		scores[i] = list[i].Score
	}

	r.UpdateScore(scores) // 更新分数
	ackTotal = &mjgame.ACKBC_Sddz_Total{
		List:       list,
		Finished:   !flag,
		RoundCount: int64(r.RoundCount),
	}
	//scores = total.TotalScore

	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Sddz_Total, ackTotal)
	//	r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Total, ackTotal)
	r.RoundResult = ackTotal
	r.InsertRoomRecord(scores)
	//房间结束
	if !flag {
		fmt.Println("局数到了")
		r.StopTicker = true
		list := r.GetSummaryList()
		r.BCMessage(mjgame.MsgID_MSG_NOTIFY_SUMMARY, &list)
		r.ClearRoomUserRoomId()
		r.Mux.Lock()
		rb.ChanRoom <- r.RoomId //销毁房间
		r.Mux.Unlock()
		return
	}
}

// 吃牌
func (r *RoomSrddz) ChowCard(user *user.User, arg *mjgame.Chow) {
	index := r.GetSeatIndexById(user.ID)

	card1 := r.CurCard
	card2 := r.GetCard(index, int(arg.Cid2))
	card3 := r.GetCard(index, int(arg.Cid3))

	if !r.IsRun || index < 0 || card1 == nil || card2 == nil || card3 == nil {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrInvalidChow)
		return
	}

	//	if !r.RoomBase.MToolChecker.CheckTool(index, int(mjgame.MsgID_MSG_ACKBC_Chow)) {
	//		fmt.Println("CheckTool不能吃 ", card1, card2, card3, len(r.WaitOptTool.NeedWaitTool))
	//		r.WaitOptTool.ClearUser(index)
	//		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrInvalidParam)
	//		return
	//	}

	if !r.CheckCanChow(index, r.CurIndex, []*rb.Card{card1, card2, card3}, r.CurCard) {
		fmt.Println("不能吃的消息 ", card1, card2, card3, len(r.WaitOptTool.NeedWaitTool))
		r.WaitOptTool.ClearUser(index)
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrInvalidParam)
		return
	}

	if r.WaitOptTool.GetOpt(index) == nil || r.WaitOptTool.GetOpt(index).CanTools[3] <= 0 {
		r.WaitOptTool.ClearUser(index)
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrInvalidParam)
		return
	}

	if r.WinUserCount == 0 {

		if r.WaitOptTool.GetOpt(index) != nil && r.WaitOptTool.GetOpt(index).Choice < 0 { //没有操作, 开始操作
			r.StateMutex.Lock()
			r.WaitOptTool.SetUserOpt(index, rb.Chow, []int{int(arg.Cid1), int(arg.Cid2), int(arg.Cid3)})
			r.RoomBase.MToolChecker.SetAllTool(index, -1) // 一个用户只能做一次操作, 操作后不管是否成功都禁止其他任何操作
			r.RoomBase.RoomRecord += "用户发送吃牌(" + r.Seats[index].User.NickName + ") " + card1.MSG + card2.MSG + card3.MSG + "\r\n"
			//r.State <- rb.WaitTool
			r.WaitPutTool()
			r.StateMutex.Unlock()
			return
		}
		//fmt.Println("----------------------------------------------> ", r.Seats[index].User.NickName, index, r.Seats[index].Cards.List.Count)
		ack := mjgame.ACKBC_Chow{
			Seat:  int32(r.Seats[index].Index),
			TSeat: int32(r.Seats[r.CurIndex].Index),
			CID1:  int32(card1.ID),
			CID2:  int32(card2.ID),
			CID3:  int32(card3.ID),
		}

		card1.TIndex, card1.Status = r.CurIndex, 1
		r.MoveChowList(index, []*rb.Card{card1, card2, card3})
		r.AddTool(st.T_Chow, index, r.CurIndex, []int{r.CurCard.ID, card2.ID, card3.ID})
		r.RoomBase.MToolChecker.SetCptTool(index, int(mjgame.MsgID_MSG_ACKBC_Chow), []int{card1.ID, card2.ID, card3.ID}, r.Seats[index].User.NickName)
		r.WaitOptTool.ClearAll()
		r.AddChengBao(index, r.CurIndex)
		r.Seats[index].LastCardID = card1.ID
		r.Seats[index].LastToolType = rb.Chow

		r.BCMessage(mjgame.MsgID_MSG_ACKBC_Chow, &ack)
		//记录回放
		r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Chow, &ack)

		r.Show = false
		r.RoomBase.RoomRecord += "吃牌(" + r.Seats[index].User.NickName + ") " + card1.MSG + card2.MSG + card3.MSG + "\r\n"
		r.Status = rb.WaitPut
		r.Seats[index].HuCardIDs = []int{}
		for i := 0; i < 4; i++ {
			r.Seats[i].Message = nil
		}
		//fmt.Println("----------------------------------------------> ", r.Seats[index].User.NickName, index, r.Seats[index].Cards.List.Count)

		//r.TurnNextPlayer(false, true, false)

	}
}

// 碰牌
func (r *RoomSrddz) PengCard(user *user.User, cid int) {
	index := r.GetSeatIndexById(user.ID)

	if !r.IsRun || index < 0 {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrMessageIsEnd)
		return
	}

	if !r.RoomBase.MToolChecker.CheckTool(index, int(mjgame.MsgID_MSG_ACKBC_Peng)) {
		fmt.Println("CheckTool  can't peng ")
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrInvalidParam)
		return
	}

	if !r.RoomBase.CheckCanPeng(index, cid, r.CurCard) {
		fmt.Println("can't peng ")
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrInvalidParam)
		return
	}
	if r.WinUserCount == 0 {

		if r.WaitOptTool.GetOpt(index) != nil && r.WaitOptTool.GetOpt(index).Choice < 0 { //没有操作, 开始操作
			r.StateMutex.Lock()
			r.WaitOptTool.SetUserOpt(index, rb.Peng, []int{int(cid)})
			r.RoomBase.MToolChecker.SetAllTool(index, -1) // 一个用户只能做一次操作, 操作后不管是否成功都禁止其他任何操作
			r.RoomBase.MToolChecker.ShowAllTools()
			r.RoomBase.RoomRecord += "用户发送碰牌(" + r.Seats[index].User.NickName + ") " + st.GetMjNameForIndex(cid) + "\r\n"
			//r.State <- rb.WaitTool
			r.WaitPutTool()
			r.StateMutex.Unlock()
			return
		}
		//fmt.Println("----------------------------------------------> ", r.Seats[index].User.NickName, index, r.Seats[index].Cards.List.Count)

		cardType, number := st.GetMjTypeNum(cid)
		targetCard := &rb.Card{
			ID:     cid,
			Type:   cardType,
			Num:    number,
			TIndex: r.CurIndex,
			MSG:    st.GetMjName(cardType, number),
		}
		r.MoveToList(r.Seats[r.CurIndex].Cards.Out, []*rb.Card{targetCard}, r.Seats[index].Cards.Peng)
		sourceCard := &rb.Card{
			ID:   cid,
			Type: cardType,
			Num:  number,
		}
		r.MoveToList(r.Seats[index].Cards.List, []*rb.Card{sourceCard, sourceCard}, r.Seats[index].Cards.Peng)
		ack := &mjgame.ACKBC_Peng{
			Seat:  int32(r.Seats[index].Index),
			TSeat: int32(r.Seats[r.CurIndex].Index),
			CID:   int32(cid),
		}

		r.RoomBase.RoomRecord += "碰牌(" + r.Seats[index].User.NickName + ") " + targetCard.MSG + "\r\n"
		r.RoomBase.MToolChecker.SetCptTool(index, int(mjgame.MsgID_MSG_ACKBC_Peng), []int{cid}, r.Seats[index].User.NickName)

		r.AddTool(st.T_Peng, index, r.CurIndex, []int{cid, cid, cid})
		r.AddChengBao(index, r.CurIndex)

		r.Seats[index].LastCardID = cid
		r.Seats[index].LastToolType = rb.Peng
		r.Show = false
		r.WaitOptTool.ClearAll()
		r.BCMessage(mjgame.MsgID_MSG_ACKBC_Peng, ack)
		r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Peng, ack)
		r.Status = rb.WaitPut
		if index <= 0 {
			r.CurIndex = index - 1 + 4
		} else {
			r.CurIndex = index - 1
		}
		r.Seats[index].HuCardIDs = []int{}

		for i := 0; i < 4; i++ {
			r.Seats[i].Message = nil
		}
		//fmt.Println("----------------------------------------------> ", r.Seats[index].User.NickName, index, r.Seats[index].Cards.List.Count)

		//r.TurnNextPlayer(false, true, false)

	} else {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrCardCanNotPeng)
	}
}

// 杠牌
func (r *RoomSrddz) KongCard(user *user.User, cid int) {
	index := r.GetSeatIndexById(user.ID)

	if !r.IsRun || index < 0 {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrMessageIsEnd)
		return
	}

	if !r.RoomBase.MToolChecker.CheckTool(index, int(mjgame.MsgID_MSG_ACKBC_Kong)) {
		fmt.Println("CheckTool  can't kong ", index, r.Seats[index].User.NickName)
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrInvalidParam)
		return
	}

	_, kongType := r.RoomBase.CheckCanKong(index, cid, true)

	if index >= 0 && kongType > 0 && r.WinUserCount == 0 {

		t, n := st.GetMjTypeNum(cid)
		count := r.GetCardCount(r.Seats[index].Cards.List, t, n)
		pengcount := r.GetCardCount(r.Seats[index].Cards.Peng, t, n)
		if count <= 0 {
			return
		}

		if kongType == def.KongTypeMing { //明杠
			if count < 3 {
				return
			}
		} else if kongType == def.KongTypeAn { //暗杠
			if count < 4 {
				return
			}
		} else if kongType == def.KongTypePeng { //碰杠
			if count < 1 && pengcount < 3 {
				return
			}
		}

		if r.WaitOptTool.GetOpt(index) != nil && r.WaitOptTool.GetOpt(index).Choice < 0 { //没有操作, 开始操作
			//r.StateMutex.Lock()

			r.WaitOptTool.SetUserOpt(index, rb.Kong, []int{cid})
			r.RoomBase.MToolChecker.SetAllTool(index, -1) // 一个用户只能做一次操作, 操作后不管是否成功都禁止其他任何操作
			r.RoomBase.RoomRecord += "用户发送杠牌(" + r.Seats[index].User.NickName + ") " + st.GetMjNameForIndex(cid) + "\r\n"
			//r.State <- rb.WaitTool
			r.WaitPutTool()
			//r.StateMutex.Unlock()
			return
		}
		//fmt.Println("----------------------------------------------> ", r.Seats[index].User.NickName, index, r.Seats[index].Cards.List.Count)
		KType := st.T_Kong
		TIndex := r.CurIndex // 碰杠的的TargetIndex
		if kongType == def.KongTypeMing {
			KType = st.T_Kong
		} else if kongType == def.KongTypeAn {
			KType = st.T_AnKong
		} else if kongType == def.KongTypePeng {
			KType = st.T_PengKong
			pengCard := r.GetUserCard(r.Seats[index].Cards.Peng, t, n)
			if pengCard != nil {
				TIndex = pengCard.TIndex
			}
		}

		ack := mjgame.ACKBC_Kong{
			Seat:     int32(r.Seats[index].Index),
			TSeat:    int32(TIndex),
			KongType: int32(KType),
			CID:      int32(cid),
		}
		//r.Mux.Lock()
		r.MoveKongList(index, TIndex, cid, kongType)
		r.Show = false
		r.Status = rb.WaitPut

		r.WaitOptTool.ClearAll() // 完成操作, 清空
		r.RoomBase.MToolChecker.SetCptTool(index, int(mjgame.MsgID_MSG_ACKBC_Kong), []int{cid}, r.Seats[index].User.NickName)
		//fmt.Println("----------------------------------------------> ", r.Seats[index].User.NickName, index, r.Seats[index].Cards.List.Count)
		r.RoomBase.RoomRecord += "杠牌(" + r.Seats[index].User.NickName + ") " + st.GetMjName(t, n) + "\r\n"
		r.Seats[index].HuCardIDs = []int{}
		r.AddTool(KType, index, r.CurIndex, []int{cid, cid, cid, cid})
		r.AddChengBao(index, r.CurIndex)

		r.BCMessage(mjgame.MsgID_MSG_ACKBC_Kong, &ack)
		r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Kong, &ack)

		//自摸
		if r.CurIndex == index {
			if KType == st.T_PengKong {
				//拉杠胡
				r.DealKongHu(cid)
				if r.IsKongHu {
					return
				}
			}
		}
		if index <= 0 {
			r.CurIndex = index - 1 + 4
		} else {
			r.CurIndex = index - 1
		}
		for i := 0; i < 4; i++ {
			r.Seats[i].Message = nil
		}
		//r.TurnNextPlayer(true, false, false)

	} else {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrCardCanNotKong)
	}
}

//过
func (r *RoomSrddz) Pass(user *user.User) {

	index := r.GetSeatIndexById(user.ID)

	r.SaveBattleRecord(-1, mjgame.MsgID_MSG_Pass, &mjgame.Pass{
		SID: strconv.Itoa(user.ID),
	})
	if !r.IsRun || index < 0 {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrMessageIsEnd)
		return
	}

	if !r.RoomBase.MToolChecker.CheckTool(index, int(mjgame.MsgID_MSG_ACKBC_Chow)) {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrInvalidParam)
		return
	}

	if r.WaitOptTool.GetOpt(index) != nil && r.WaitOptTool.GetOpt(index).CanTools[0] > 0 { //过手胡
		r.Seats[index].HuCardIDs = append(r.Seats[index].HuCardIDs, r.CurCard.ID)
	}

	if r.WaitOptTool.GetOpt(index) != nil && r.WaitOptTool.GetOpt(index).CanTools[5] > 0 && r.WaitOptTool.GetOpt(index).Choice < 0 {
		r.WaitOptTool.SetUserOpt(index, rb.Pass, []int{})
		//fmt.Println("----------------------------------------------> ", r.Seats[index].User.NickName, index, r.Seats[index].Cards.List.Count)
		r.RoomBase.MToolChecker.SetAllTool(index, -1) // 一个用只能做一个操作, 操作后不管是否成功都禁止其他任何操作
	}

	//	if r.Status == rb.WaitTool && r.WaitOptTool.Count() > 0 {
	//	}
	r.RoomBase.RoomRecord += "过(" + r.Seats[index].User.NickName + ") \r\n"
	if index >= 0 {
		r.Seats[index].Message = nil
	}

	r.WaitPutTool()

	//r.State <- rb.WaitTool
	//	else if r.Status == rb.WaitPut {
	//		if index == r.CurIndex {
	//			r.CurIndex = r.CurIndex - 1
	//			r.TurnNextPlayer(false, false)
	//		} else {
	//			r.TurnNextPlayer(true, true)
	//		}
	//	}

}

//胡
func (r *RoomSrddz) WinCard(users []*user.User, cid int) {
	var targetIndex int
	var seats []int32

	index := r.GetSeatIndexById(users[0].ID)

	if !r.IsRun || index < 0 {
		users[0].SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrInvalidParam)
		return
	}
	if !r.RoomBase.MToolChecker.CheckTool(index, int(mjgame.MsgID_MSG_Win)) {
		users[0].SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrInvalidParam)
		return
	}

	//处理拉杠胡
	if r.IsKongHu {
		cid = r.KongHuCardID
	}

	for _, user := range users {
		index := r.GetSeatIndexById(user.ID)
		if r.Seats[index].IsCanWin {
			if r.WaitOptTool.GetOpt(index) != nil && r.WaitOptTool.GetOpt(index).Choice < 0 { //没有操作, 开始操作
				r.StateMutex.Lock()
				r.WaitOptTool.SetUserOpt(index, rb.Hu, []int{int(cid)})
				r.RoomBase.MToolChecker.SetAllTool(index, -1) // 一个用户只能做一次操作, 操作后不管是否成功都禁止其他任何操作
				r.State = rb.WaitTool
				r.StateMutex.Unlock()
				return
			}
		}
		r.WinUserCount++
		if index == r.CurIndex {
			r.AddTool(st.T_Hu_ZiMo, index, r.CurIndex, []int{cid}) //自摸
			targetIndex = -1
		} else {
			r.AddTool(st.T_Hu, index, r.CurIndex, []int{cid}) // 点炮
			targetIndex = r.CurIndex                          // fix点炮位置
			r.Seats[r.CurIndex].Accumulation.FireCount++
		}

		r.RoomBase.RoomRecord += "胡牌(" + r.Seats[index].User.NickName + ") " + st.GetMjNameForIndex(cid) + "\r\n"
		r.Seats[index].Accumulation.WinCount++
		seats = append(seats, int32(index))
	}

	if len(seats) > 0 {
		allSeatCards := r.GetAllSeatCards()
		win := &mjgame.ACKBC_Win{
			Seat:  seats,
			TSeat: int32(targetIndex),
			CID:   int32(cid),
			Cards: allSeatCards,
		}
		r.BCMessage(mjgame.MsgID_MSG_ACKBC_Win, win)

		recordWin := &mjgame.ACKBC_Win{
			Seat:  seats,
			TSeat: int32(targetIndex),
			CID:   int32(cid),
			Cards: r.GetRecordAllSeatCards(),
		}
		r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACKBC_Win, recordWin)

		for k, _ := range r.Seats {
			r.Seats[k].Message = &rb.Message{
				Id:      mjgame.MsgID_MSG_ACKBC_Win,
				Content: win,
			}
		}
		r.RoomBase.MToolChecker.SetAllUserTool(-1) // 禁止所有操作
		r.RoundToatlFinish = false
		r.IsRun = false
		r.RoundTotal()
		r.RoundToatlFinish = true

	}
}

//解散
func (r *RoomSrddz) DisbandRoom(user *user.User, args *mjgame.Disband) {
	if !r.IsRun && r.RoundCount == 0 { //游戏未开始
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrCanNotDisband)
		return
	}

	index := r.GetSeatIndexById(user.ID)
	if index < 0 {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrWatchUserCannotDisbanding) // 围观用户不能发起解散或投票
		return
	}

	if r.VoteStarter < 0 { // 发起投票
		r.VoteTimeCount = 0
		r.VoteStarter = index
		r.Votes[index] = Agree
	}

	disband := &mjgame.AckDisband{}
	disband.LeftTime = int32(r.VoteTimeOut - r.VoteTimeCount)
	if disband.LeftTime < 0 {
		disband.LeftTime = 0
	}
	disband.List = common.BuildSeatBaseToVotes(r.Votes, r.VoteStarter, r.Seats)
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Disband_Room, disband)
}

func (r *RoomSrddz) OwnerDisbandRoom(user *user.User, args *mjgame.Roomowner_Disband_Room) {
	if r.IsRun || r.RoundCount > 0 { //游戏未开始
		//		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrCanNotDisband)
		return
	}
	if user.ID != r.CreateUserId { //不是房主不能解散
		return
	}
	r.ClearRoomUserRoomId()
	r.BCMessage(mjgame.MsgID_MSG_NOTIFY_DESTORY_ROOM, &mjgame.NotifyDestoryRoom{RoomId: int32(r.RoomId), IsOwnerDisband: true})
	r.StopTicker = true
	fmt.Println("房主解散了房间")
	r.Mux.Lock()
	rb.ChanRoom <- r.RoomId //销毁房间
	r.Mux.Unlock()
}

// 投票 0=未操作   1=同意  2=反对
// 超时后 未操作的玩家默认同意
func (r *RoomSrddz) Vote(user *user.User, args *mjgame.Vote) {
	index := r.GetSeatIndexById(user.ID)
	if index < 0 {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrWatchUserCannotDisbanding)
		return
	}

	if r.Votes[index] > 0 {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrHasAlreadyVoted)
		return
	}

	r.Votes[index] = int(args.Result)

	fmt.Println(r.Votes)

	disband := &mjgame.AckDisband{}
	disband.LeftTime = int32(r.VoteTimeOut - r.VoteTimeCount)
	if disband.LeftTime < 0 {
		disband.LeftTime = 0
	}
	fmt.Println(disband.LeftTime)
	disband.List = common.BuildSeatBaseToVotes(r.Votes, r.VoteStarter, r.Seats)
	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Disband_Room, disband)

	disbandState := r.IsDisbanding()

	if disbandState == Agree {
		// 解散成功
		r.BCMessage(mjgame.MsgID_MSG_NOTIFY_DISBAND, &mjgame.NotifyDisband{RoomId: int32(r.RoomId), Result: DisbandSuccess})
		r.VoteStarter = -1
		r.StopTicker = true
		r.DestoryRoom()

	} else if disbandState == 2 {
		// 解散失败
		r.BCMessage(mjgame.MsgID_MSG_NOTIFY_DISBAND, &mjgame.NotifyDisband{RoomId: int32(r.RoomId), Result: DisbandFail})
		r.Votes = []int{0, 0, 0, 0}
		r.VoteStarter = -1
	}
}

//踢人
func (r *RoomSrddz) Kick(user *user.User, args *mjgame.KickRequest) {
	r.StateMutex.Lock()
	defer r.StateMutex.Unlock()
	index := r.GetSeatIndexById(user.ID)
	if index < 0 {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrHasNoRightsToKickUser)
		return
	}

	kickSeat := r.Seats[int(args.Index)]
	if kickSeat == nil || kickSeat.User == nil {
		return
	}
	kickUserID := kickSeat.User.ID

	//校验是否可以踢
	if kickSeat.User.State == def.Offline {
		if kickSeat.OfflineTime.Add(def.KickTimeDuration * time.Second).After(time.Now()) {
			user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrInvalidParam)
			return
		}
		kickResponse := &mjgame.KickResponse{
			Index:  int32(args.Index),
			UserId: strconv.Itoa(kickUserID),
		}
		r.BCMessage(mjgame.MsgID_MSG_ACKBC_KICK, kickResponse)

		if r.IsRun {
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

			r.AddTool(st.T_Draw, -1, -1, []int{})
			//			r.StateMutex.Lock()
			r.RoundTotal()
			//r.State <- rb.Total
			//			r.StateMutex.Unlock()

			for k, _ := range r.Seats {
				r.Seats[k].Message = &rb.Message{
					Id:      mjgame.MsgID_MSG_ACKBC_Draw,
					Content: ackDraw,
				}
			}
			o := &KickInfo{
				UserID:   kickUserID,
				Position: int(args.Index),
			}
			r.KickUsers = append(r.KickUsers, o)
		}
		seat := r.Seats[int(args.Index)]
		r.WatchSeats = append(r.WatchSeats, seat.User)

		r.Seats[int(args.Index)].UID = ""
		r.Seats[int(args.Index)].User = nil

		if r.IsRun {
			r.Draw()
		}
		//		r.Draw()

		//		r.StopTicker = true
		//		r.DestoryRoom()
		//		r.Mux.Lock()
		//		rb.ChanRoom <- r.RoomId //销毁房间
		//		r.Mux.Unlock()
		//&rb.SeatBase{Accumulation: &rb.Accumulation{}}
	}
}
