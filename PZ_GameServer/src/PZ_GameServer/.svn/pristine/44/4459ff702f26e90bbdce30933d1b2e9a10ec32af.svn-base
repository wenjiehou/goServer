package roombase

import (
	"fmt"
	"sync"
)

// 操作
type IndexTool struct {
	SeatIndex int   // 座位Index
	CanTool   []int // 可以进行的操作
}

// 玩家操作按优先级排序[]
// 胡 杠 碰 吃 出 过 摸

type IToolChecker interface {
	CheckEvtTool(iSeatIndex int, toolType int, cids []int) // 检查操作(吃 碰 杠 胡 过 出牌)
	SetEvtCheck(iSeatIndex int, toolType int, cids []int)  // 操作成功后, 设置检查内容
	ClearEvtCheck()                                        // 清空检查
}

// 操作检查者 (Tool Check Main)
type ToolChecker struct {
	Users     []IndexTool //
	WaitUser  []int       // 等待用户
	WaitTool  []int       // 等待操作 []
	SeatCount int         // 座位总数
	inited    bool        //
	toolIndex map[int]int // 操作索引
	Mutex     sync.Mutex  //状态机互斥锁
}

// 初始化 SeatCount  座位总数
func (tc *ToolChecker) Init(seatCount int) {

	tc.Users = make([]IndexTool, seatCount)
	for i := 0; i < seatCount; i++ {
		tc.Users[i] = IndexTool{SeatIndex: i, CanTool: []int{0, 0, 0, 0, 0, 0, 0, 0}} // 胡 杠 碰 吃 出 过 摸 等
	}
	tc.SeatCount = seatCount

	tc.toolIndex = make(map[int]int)
	tc.toolIndex[320] = 0 //MsgID_MSG_Win
	tc.toolIndex[760] = 1 //MsgID_MSG_ACKBC_Kong
	tc.toolIndex[750] = 2 //MsgID_MSG_ACKBC_Peng
	tc.toolIndex[740] = 3 //MsgID_MSG_ACKBC_Chow
	tc.toolIndex[720] = 4 //MsgID_MSG_ACKBC_PutCard
	tc.toolIndex[330] = 5 //MsgID_MSG_Pass
	tc.toolIndex[710] = 6 //MsgID_MSG_ACKBC_GetCard
	tc.toolIndex[730] = 7 //MsgID_MSG_ACK_WaitTool

	tc.inited = true
}

// 得到ToolIndex
func (tc *ToolChecker) getToolIndex(toolType int) int {
	if i, ok := tc.toolIndex[toolType]; ok {
		return i
	}
	return -1
}

// 检查操作(胡0 杠1 碰2 吃3 出4 过5 摸6 等7)
// -1 为不能操作
// 0 可以操作
func (tc *ToolChecker) CheckTool(iSeatIndex int, toolType int) bool {

	if iSeatIndex < 0 || iSeatIndex > 3 {
		return false // 找不到玩家
	}
	toolindex := tc.getToolIndex(toolType)
	if toolindex < 0 {
		return false // 没有这种操作
	}
	can := tc.Users[iSeatIndex].CanTool[toolindex]
	if can < 0 {
		fmt.Println("nocan ", tc.Users[iSeatIndex].CanTool)
		return false
	} else {
		return true
	}
}

// 胡0 杠1 碰2 吃3 出4 过5 摸6 等7
// 设置完成的操作,   该函数自动生成
func (tc *ToolChecker) SetCptTool(iSeatIndex int, toolType int, cids []int, uid string) bool {
	//玩家顺序 0-3
	if iSeatIndex < 0 || iSeatIndex > 3 {
		return false // 找不到玩家
	}
	toolindex := tc.getToolIndex(toolType)
	if toolindex < 0 {
		return false // 没有这种操作
	}

	tc.SetAllUserTool(-1)                 // 全部禁止
	elseIndex := tc.ElseIndex(iSeatIndex) // 其他玩家
	nextIndex := tc.NextIndex(iSeatIndex) // 下家

	switch toolType {
	case 320: //MsgID_MSG_Win 胡
		// 自己禁止任何操作
		// 其他人可以胡
		for i := 0; i < len(elseIndex); i++ {
			tc.SetTool(elseIndex[i], 0, 0)
		}

	case 760: // MsgID_MSG_ACKBC_Kong 杠
		// 自己可以摸牌
		// 其他人可以 胡(抢杠胡)
		for i := 0; i < len(elseIndex); i++ {
			tc.SetTools(elseIndex[i], []int{0, -1, -1, -1, -1, -1, -1, -1}) // 胡0 杠1 碰2 吃3 出4 过5 摸6 等7
		}
		tc.SetTool(iSeatIndex, 6, 0)

	case 750: //MsgID_MSG_ACKBC_Peng 碰
		// 自己可以出牌
		// 其他人可以胡 杠 吃 过 (WaitTool)
		for i := 0; i < len(elseIndex); i++ {
			tc.SetTools(elseIndex[i], []int{0, -1, -1, -1, -1, 0, -1, -1}) // 胡0 杠1 碰2 吃3 出4 过5 摸6 等7
		}
		tc.SetTool(iSeatIndex, 4, 0)

	case 740: // MsgID_MSG_ACKBC_Chow 吃
		// 自己可以出牌
		// 其他人可以胡 杠 碰 过
		for i := 0; i < len(elseIndex); i++ {
			tc.SetTools(elseIndex[i], []int{-1, -1, -1, -1, 0, -1, -1, -1}) // 胡0 杠1 碰2 吃3 出4 过5 摸6 等7
		}
		tc.SetTool(iSeatIndex, 4, 0)

	case 720: // MsgID_MSG_ACKBC_PutCard 出牌
		// 自己禁止任何操作
		// 其他人可   胡 杠 碰 过
		// 下家可以   胡 杠 碰 吃 出 过 摸
		//		for i := 0; i < len(elseIndex); i++ {
		//			tc.SetTools(elseIndex[i], []int{0, 0, 0, 0, -1, 0, -1, -1}) // 胡0 杠1 碰2 吃3 出4 过5 摸6 等7
		//		}
		tc.SetAllTool(nextIndex, 0)

	case 330: // MsgID_MSG_Pass  过
		// 自己禁止任何操作
		// 其他人可    胡 杠 碰 吃 过 摸
		//for i := 0; i < len(elseIndex); i++ {
		//	tc.SetAllTool(i, 1)
		//}

	case 710:
		// MsgID_MSG_ACKBC_GetCard  摸牌
		// 自己可以 胡 杠 出
		// 其他人禁止任何操作
		tc.SetTools(iSeatIndex, []int{0, 0, -1, -1, 0, -1, -1, -1}) // 胡0 杠1 碰2 吃3 出4 过5 摸6 等7
	case 730:
		// MsgID_MSG_ACK_WaitTool  等待操作
		// 自己禁止任何操作
		// 其他人可以胡 杠 碰 过
		// 下家可以 胡 杠 碰 吃 过 摸
		for i := 0; i < len(elseIndex); i++ {
			tc.SetTools(elseIndex[i], []int{0, 0, 0, -1, -1, 0, -1, -1})
		}
		tc.SetTool(nextIndex, 3, 0) // 胡0 杠1 碰2 吃3 出4 过5 摸6 等7
		//tc.SetTools(iSeatIndex, []int{0, 0, -1, -1, 0, 0, -1, -1})
	}
	return true
}

// 下一个位置
func (tc *ToolChecker) NextIndex(iSeatIndex int) int {
	iSeatIndex++
	if iSeatIndex > 3 {
		iSeatIndex = 0
	}
	return iSeatIndex
}

// 其他位置
func (tc *ToolChecker) ElseIndex(iSeatIndex int) []int {
	eindex := []int{0, 0, 0}
	index := 0
	for i := 0; i < tc.SeatCount; i++ {
		if i != iSeatIndex {
			eindex[index] = i
			index++
		}
	}
	return eindex
}

// 清空检查
func (tc *ToolChecker) ClearAllCheck() {
	tc.Mutex.Lock()
	for i := 0; i < tc.SeatCount; i++ {
		tc.SetAllTool(i, 0)
	}
	tc.Mutex.Unlock()
}

// 设置全部用户操作
func (tc *ToolChecker) SetAllUserTool(t int) {

	tc.Mutex.Lock()
	for u := 0; u < tc.SeatCount; u++ {
		for i := 0; i < len(tc.Users[u].CanTool); i++ {
			tc.Users[u].CanTool[i] = t
		}
	}
	tc.Mutex.Unlock()
}

// 设置全部操作
func (tc *ToolChecker) SetAllTool(iUserIndex int, t int) {
	//tc.Mutex.Lock()
	if iUserIndex > len(tc.Users)-1 {
		return
	}
	for i := 0; i < len(tc.Users[iUserIndex].CanTool); i++ {
		tc.Users[iUserIndex].CanTool[i] = t
	}
	//tc.Mutex.Unlock()
}

// 设置操作
func (tc *ToolChecker) SetTool(iUserIndex int, index int, t int) {
	if iUserIndex > len(tc.Users)-1 {
		return
	}
	tc.Users[iUserIndex].CanTool[index] = t
}

// 设置多个操作
func (tc *ToolChecker) SetTools(iUserIndex int, t []int) {
	tc.Mutex.Lock()
	for i := 0; i < len(t); i++ {
		tc.Users[iUserIndex].CanTool[i] = t[i]
	}
	tc.Mutex.Unlock()
}

func (tc *ToolChecker) ShowAllTools() {
	//	for i := 0; i < tc.SeatCount; i++ {
	//		fmt.Println("", i, tc.Users[i].CanTool)
	//	}
}
