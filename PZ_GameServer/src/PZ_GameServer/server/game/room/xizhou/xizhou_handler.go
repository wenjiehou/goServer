/*
 * +-----------------------------------------------
 * | xizhou_handler.go
 * +-----------------------------------------------
 * | Version: 1.0
 * +-----------------------------------------------
 * | Context: 西周麻将路由处理
 * +-----------------------------------------------
 */
package xizhou

import (
	"PZ_GameServer/protocol/def"
	"PZ_GameServer/protocol/pb"
	"PZ_GameServer/server/game/common"
	"PZ_GameServer/server/game/error"
	rb "PZ_GameServer/server/game/roombase"
	st "PZ_GameServer/server/game/statement"
	"PZ_GameServer/server/user"
	"fmt"

	"strconv"
	"time"
)

// 创建房间
func (r *RoomXiZhou) Create(rid int, t int32, user *user.User, rule *rb.RoomRule) {
	r.RoomBase.Create(rid, t, user.ID, rule)
	r.QuitSitUsers = make(map[int]rb.SeatBase)

	user.RoomId = rid
	user.GameType = IFCXiZhouType

	go r.TimeTicker() //定时器
	r.StlCtrl = GetXiZhouStatement(
		t,
		500,
		[]string{"", "", "", ""},
		r,
	)
}

// 进入房间
func (r *RoomXiZhou) IntoUser(user *user.User) {
	user.GameType = IFCXiZhouType
	r.IntoRoom(user)

	index := r.GetSeatIndexById(user.ID)
	if index >= 0 {
		message := r.Seats[index].Message
		if message != nil {
			user.SendMessage(message.Id, message.Content)
		}
	}

	if index >= 0 {
		if r.Seats[index].State == int(mjgame.StateID_GameState_Total) {
			if r.RoundResult != nil {
				user.SendMessage(mjgame.MsgID_MSG_ACKBC_Total, r.RoundResult)
			}
		}
	}
}

//开始游戏
func (r *RoomXiZhou) Start(user *user.User) {
	//fmt.Println("XZ Start StateMutex Lock", time.Now(), r.RoomId)
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

	r.InitRandAllCard() //洗牌

	r.StlCtrl = GetXiZhouStatement( // 结算控制器
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
			common.AddDiamondLog(seat.User, int(r.Type), -r.Rules.Play_NeedDiamond)
		}
	}

	// Redis
	for i := 0; i < len(r.Seats); i++ {
		rb.Redis_AddPlayingUser(r.RoomId, r.Seats[i].UID)
	}

	r.RoundTime = time.Now()
	r.State = rb.Dealt

	r.NewBattleRecord()
	r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACK_RoomInfo, r.GetRecordRoomInfo())

	roomUser := r.GetRoomUser()

	for _, u := range roomUser.Users {
		u.Ip = ""
		u.RoomId = 0
		u.Icon = ""
		u.Name = ""
	}
	r.SaveBattleRecord(-1, mjgame.MsgID_MSG_ACK_Room_User, roomUser)

	//@andy0920
	//go r.Process()

	index := r.GetSeatIndexById(user.ID)
	r.AddTool(st.T_Start, index, -1, []int{})

	//@andy0920
	r.NewProcess()
	//fmt.Println("XZ Start StateMutex UnLock", time.Now(), r.RoomId)
}

// 处理器
func (r *RoomXiZhou) Process() {
	for {

		if !r.IsRun {
			break
		}

		time.Sleep(10)

		switch r.State {
		case rb.Dealt: //开始
			r.State = -1
			r.InitRound()    //初始化每局
			r.InitUserCard() //创建13张牌
			r.SendGameInfo(nil, true)
			//参与用户
			for _, v := range r.Seats {
				r.SendSeatCard(v.User.ID)
			}
			//旁观用户
			r.SendSeatCard(-1)
			r.RecordSeatCard()
			r.CurIndex = r.BankerIndex - 1 // 方便TurnNextPlayer统一
			r.WaitTimeCount = 3
			r.TurnNextPlayer(true, true, false)
		case rb.WaitPut:
			r.State = -1
			r.WaitPut(r.WaitPutTimeOut)
		case rb.WaitTool:
			r.State = -1
			r.WaitPutTool()
		}

	}
}

//@andy0920
func (r *RoomXiZhou) NewProcess() {
	if !r.IsRun {
		return
	}

	switch r.State {
	case rb.Dealt: //开始
		r.State = -1
		r.InitRound()    //初始化每局
		r.InitUserCard() //创建13张牌
		r.SendGameInfo(nil, true)
		//参与用户
		for _, v := range r.Seats {
			r.SendSeatCard(v.User.ID)
		}
		//旁观用户
		r.SendSeatCard(-1)
		r.RecordSeatCard()
		r.CurIndex = r.BankerIndex - 1 // 方便TurnNextPlayer统一
		r.WaitTimeCount = 3
		r.TurnNextPlayer(true, true, false)
	case rb.WaitPut:
		r.State = -1
		r.WaitPut(r.WaitPutTimeOut)
	case rb.WaitTool:
		r.State = -1
		r.WaitPutTool()
	}
}

// 起立
func (r *RoomXiZhou) StandUp(arg *mjgame.StandUp, user *user.User) {
	//@andy0920
	//fmt.Println("XZ StandUp Lock", time.Now(), r.RoomId)
	r.GLMutex.Lock()
	defer r.GLMutex.Unlock()

	//fmt.Println("XZ StandUp StateMutex Lock", time.Now(), r.RoomId)
	r.StateMutex.Lock()
	defer r.StateMutex.Unlock()
	index := r.GetSeatIndexById(user.ID)
	if r.IsRun || index < 0 {
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

	r.BCMessage(mjgame.MsgID_MSG_ACKBC_Standup, ackBCStandUp) // 起立消息
	//fmt.Println("XZ StandUp UnLock", time.Now(), r.RoomId)
}

//坐下
func (r *RoomXiZhou) SitDown(user *user.User, arg *mjgame.SitDown) {
	//@andy0920
	//fmt.Println("XZ SitDown Lock", time.Now(), r.RoomId)
	r.GLMutex.Lock()
	defer r.GLMutex.Unlock()

	if r.Seats[arg.Index].User != nil {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrCurSeatHasUsed)
		return
	}

	if r.RoundCount == 0 {
		if user.Diamond < r.Rules.Play_NeedDiamond {
			user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrDiamondNotEnough)
			return
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
	//fmt.Println("XZ SitDown UnLock", time.Now(), r.RoomId)
}

//退出
func (r *RoomXiZhou) ExitUser(user *user.User) {
	//@andy0920
	//fmt.Println("XZ ExitUser Lock", time.Now(), r.RoomId)
	r.GLMutex.Lock()
	defer r.GLMutex.Unlock()

	index := r.GetSeatIndexById(user.ID)

	if r.IsRun {
		if index >= 0 || r.RoundToatlFinish == false {
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
		r.Seats[index] = &rb.SeatBase{}
		r.Seats[index].Message = nil
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

	user.ConReadCount = 0
	user.RoomId = 0
	user.GameType = nil
	//fmt.Println("XZ ExitUser UnLock", time.Now(), r.RoomId)
}

//准备
func (r *RoomXiZhou) Ready(user *user.User) {
	//@andy0920
	//fmt.Println("XZ Ready Lock", time.Now(), r.RoomId)
	r.GLMutex.Lock()
	defer r.GLMutex.Unlock()

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
	//fmt.Println("XZ Ready UnLock", time.Now(), r.RoomId)
}

// 出牌
func (r *RoomXiZhou) PutCard(user *user.User, arg *mjgame.Put_Card) {
	//fmt.Println("XZ PutCard Lock", time.Now(), r.RoomId)

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
	//fmt.Println("XZ PutCard UnLock", time.Now(), r.RoomId)
}

// 吃牌
func (r *RoomXiZhou) ChowCard(user *user.User, arg *mjgame.Chow) {
	//fmt.Println("XZ ChowCard Lock", time.Now(), r.RoomId)
	index := r.GetSeatIndexById(user.ID)
	//r.SaveBattleRecord(-1, mjgame.MsgID_MSG_Chow, strconv.Itoa(user.ID))
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
			defer r.StateMutex.Unlock()
			r.WaitOptTool.SetUserOpt(index, rb.Chow, []int{int(arg.Cid1), int(arg.Cid2), int(arg.Cid3)})
			r.RoomBase.MToolChecker.SetAllTool(index, -1) // 一个用户只能做一次操作, 操作后不管是否成功都禁止其他任何操作
			r.RoomBase.RoomRecord += "用户发送吃牌(" + r.Seats[index].User.NickName + ") " + card1.MSG + card2.MSG + card3.MSG + "\r\n"
			//r.State <- rb.WaitTool
			r.WaitPutTool()
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

		r.CurCard = nil
		r.TurnNextPlayer(false, true, false)

	}
	//fmt.Println("XZ ChowCard UnLock", time.Now(), r.RoomId)
}

// 碰牌
func (r *RoomXiZhou) PengCard(user *user.User, cid int) {
	//fmt.Println("XZ PengCard Lock", time.Now(), r.RoomId)
	index := r.GetSeatIndexById(user.ID)
	//r.SaveBattleRecord(-1, mjgame.MsgID_MSG_Peng, strconv.Itoa(user.ID))

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
			defer r.StateMutex.Unlock()
			r.WaitOptTool.SetUserOpt(index, rb.Peng, []int{int(cid)})
			r.RoomBase.MToolChecker.SetAllTool(index, -1) // 一个用户只能做一次操作, 操作后不管是否成功都禁止其他任何操作
			r.RoomBase.MToolChecker.ShowAllTools()
			r.RoomBase.RoomRecord += "用户发送碰牌(" + r.Seats[index].User.NickName + ") " + st.GetMjNameForIndex(cid) + "\r\n"
			//r.State <- rb.WaitTool
			r.WaitPutTool()
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

		r.CurCard = nil
		r.TurnNextPlayer(false, true, false)

	} else {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrCardCanNotPeng)
	}
	//fmt.Println("XZ PengCard UnLock", time.Now(), r.RoomId)
}

// 杠牌
func (r *RoomXiZhou) KongCard(user *user.User, cid int) {
	//fmt.Println("XZ KongCard Lock", time.Now(), r.RoomId)
	index := r.GetSeatIndexById(user.ID)
	//r.SaveBattleRecord(-1, mjgame.MsgID_MSG_Kong, strconv.Itoa(user.ID))

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

		if KType == st.T_Kong {
			r.CurCard = nil
		}

		r.TurnNextPlayer(true, false, false)

	} else {
		user.SendMessage(mjgame.MsgID_MSG_ACK_Error, error.ErrCardCanNotKong)
	}
	//fmt.Println("XZ KongCard UnLock", time.Now(), r.RoomId)
}

//过
func (r *RoomXiZhou) Pass(user *user.User) {
	//fmt.Println("XZ Pass Lock", time.Now(), r.RoomId)

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

	//fmt.Println("XZ Pass UnLock", time.Now(), r.RoomId)
}

//胡
func (r *RoomXiZhou) WinCard(users []*user.User, cid int) {
	//fmt.Println("XZ WinCard Lock", time.Now(), r.RoomId)
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
				defer r.StateMutex.Unlock()
				r.WaitOptTool.SetUserOpt(index, rb.Hu, []int{int(cid)})
				r.RoomBase.MToolChecker.SetAllTool(index, -1) // 一个用户只能做一次操作, 操作后不管是否成功都禁止其他任何操作
				r.State = rb.WaitTool
				//@andy0920
				//r.NewProcess()
				fmt.Println("wanjiacaozuole hu .....")
				r.WaitPutTool()
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

		r.RoundTotal()
		r.IsRun = false
		r.RoundToatlFinish = true

	}
	//fmt.Println("XZ WinCard UnLock", time.Now(), r.RoomId)
}

//解散
func (r *RoomXiZhou) DisbandRoom(user *user.User, args *mjgame.Disband) {
	//fmt.Println("XZ DisbandRoom Lock", time.Now(), r.RoomId)
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
	//fmt.Println("XZ DisbandRoom UnLock", time.Now(), r.RoomId)
}

// 投票 0=未操作   1=同意  2=反对
// 超时后 未操作的玩家默认同意
func (r *RoomXiZhou) Vote(user *user.User, args *mjgame.Vote) {
	//fmt.Println("XZ Vote Lock", time.Now(), r.RoomId)
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
	//fmt.Println("XZ Vote UnLock", time.Now(), r.RoomId)
}

//踢人
func (r *RoomXiZhou) Kick(user *user.User, args *mjgame.KickRequest) {
	//fmt.Println("XZ Kick Lock", time.Now(), r.RoomId)
	r.GLMutex.Lock()
	defer r.GLMutex.Unlock()

	//fmt.Println("XZ Kick StateMutex Lock", time.Now(), r.RoomId)
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
	//fmt.Println("XZ Kick UnLock", time.Now(), r.RoomId)
}
