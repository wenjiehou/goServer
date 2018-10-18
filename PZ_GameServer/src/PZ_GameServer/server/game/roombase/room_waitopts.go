/*
	胡杠碰吃出过 操作基类
*/
package roombase

import (
	//"container/list"
	"sync"
)

var maxOpt = 6 // 用来排序, 值越小优先级越高

// 需要等待请求
type NeedWait struct {
	Index    int   // 座位号, 和用户分离
	IsPass   bool  // 是否过
	CanTools []int // 用户可以的选择 {0胡 1杠 2碰 3吃 4出 5过}    <=0 不能操作    >0 可以操作   是多选
	Choice   int   // 用户的选择 {0胡 1杠 2碰 3吃 4出 5过}    -1 = 没有操作
	TopOpt   int   // 最高级别的操作
	Param    []int // 操作的cids
}

// 当前等待的操作
type RoomWaitOpts struct {
	IsSelf       bool        // 是否是判断自己
	WaitOpts     []int       // 等待操作的用户(一炮多响, 需要等待多个用户)
	OptIndex     []int       // 操作的用户 0,1,2,3
	NeedWaitTool []*NeedWait //
	mx           *sync.Mutex
}

// 清空所有操作
func (rw *RoomWaitOpts) ClearAll() {
	if rw.mx == nil {
		rw.mx = new(sync.Mutex)
	}
	rw.mx.Lock()
	rw.IsSelf = false
	rw.WaitOpts = []int{}
	rw.OptIndex = []int{}
	if len(rw.NeedWaitTool) > 0 {
		for i := 0; i < len(rw.NeedWaitTool); i++ {
			rw.NeedWaitTool[i].Index = -1
			rw.NeedWaitTool[i].Choice = maxOpt
			rw.NeedWaitTool[i] = nil
		}
	}
	rw.NeedWaitTool = nil
	rw.NeedWaitTool = []*NeedWait{}
	rw.mx.Unlock()
}

// 清除
func (rw *RoomWaitOpts) ClearUser(uIndex int) {

	if rw.mx == nil {
		rw.mx = new(sync.Mutex)
	}

	rw.mx.Lock()
	for i := 0; i < len(rw.NeedWaitTool); i++ {
		if rw.NeedWaitTool[i] == nil || rw.NeedWaitTool[i].Index == uIndex {
			rw.NeedWaitTool = append(rw.NeedWaitTool[:i], rw.NeedWaitTool[i+1:]...)
		}
	}
	rw.mx.Unlock()
}

// 添加用户操作
// <=0 没有操作  >0 有操作
func (rw *RoomWaitOpts) AddCanTool(uIndex int, iWin int, iKong int, iPeng int, iChow int, iPut int, iPass int) {
	findcount := -1

	for i := 0; i < len(rw.NeedWaitTool); i++ {
		if rw.NeedWaitTool[i].Index == uIndex { // 经存在
			rw.NeedWaitTool[i].CanTools[0] += iWin
			rw.NeedWaitTool[i].CanTools[1] += iKong
			rw.NeedWaitTool[i].CanTools[2] += iPeng
			rw.NeedWaitTool[i].CanTools[3] += iChow
			rw.NeedWaitTool[i].CanTools[4] += iPut
			rw.NeedWaitTool[i].CanTools[5] += iPass
			rw.NeedWaitTool[i].TopOpt = rw.getUserTopOpt(uIndex)
			findcount = i
			break
		}
	}

	if findcount == -1 {
		nw := NeedWait{ // 新添加操作
			Index: uIndex, Choice: -1,
			CanTools: []int{
				iWin,
				iKong,
				iPeng,
				iChow,
				iPut,
				iPass,
			},
		}
		rw.NeedWaitTool = append(rw.NeedWaitTool, &nw)
		nw.TopOpt = rw.getUserTopOpt(uIndex)
	}

}

// 得到最高优先级别的操作(单个用户)
func (rw *RoomWaitOpts) getUserTopOpt(uIndex int) int {
	needWait := rw.GetOpt(uIndex)
	if needWait == nil {
		return -1
	}
	topOpt := maxOpt
	for t := 0; t < len(needWait.CanTools); t++ {
		if needWait.CanTools[t] > 0 {
			if topOpt >= t {
				topOpt = t
			}
		}
	}
	if topOpt == maxOpt {
		return -1
	} else {
		return topOpt
	}
}

// 用户操作
func (rw *RoomWaitOpts) SetUserOpt(index int, optType int, param []int) {
	for i := 0; i < len(rw.NeedWaitTool); i++ {
		if rw.NeedWaitTool[i].Index == index { //  存在
			if rw.NeedWaitTool[i].Choice == -1 && rw.NeedWaitTool[i].CanTools[optType] > 0 { // 必须要未操作的用户才可以设置
				rw.NeedWaitTool[i].Choice = optType // 设置用户操作
				rw.NeedWaitTool[i].Param = param
				if optType == Pass {
					rw.NeedWaitTool[i].IsPass = true // 过
				}
			}
			break
		}
	}
}

// 得到操作数量
func (rw *RoomWaitOpts) Count() int {
	return len(rw.NeedWaitTool)
}

// 得到NeedWait
func (rw *RoomWaitOpts) GetOpt(userIndex int) *NeedWait {
	for i := 0; i < len(rw.NeedWaitTool); i++ {
		if rw.NeedWaitTool[i] != nil && rw.NeedWaitTool[i].Index == userIndex { // 已经存在
			return rw.NeedWaitTool[i]
		}
	}
	return nil
}

// 得到可以操作的类型
func (rw *RoomWaitOpts) GetOptCanTools(userIndex int) []int {
	if rw.GetOpt(userIndex) != nil {
		return rw.GetOpt(userIndex).CanTools
	}
	var rValue []int
	return rValue

}

// 判断操作是否完毕, true =则开始执行最高优先级的操作
// 检查并返回最高优先级的操作.
// 返回 (0用户索引[]int, 1操作类型int, 2是否成功bool)
//
func (rw *RoomWaitOpts) CheckGetCpt() ([]int, int, int, bool) {
	huList := make([]int, 0)

	// 获得最高级别的操作和用户
	topOpt := maxOpt // 最高级别的操作
	topIndex := -1
	for i := 0; i < len(rw.NeedWaitTool); i++ {
		needWait := rw.NeedWaitTool[i]
		if needWait != nil && !needWait.IsPass && topOpt >= needWait.TopOpt {
			topOpt = needWait.TopOpt
			topIndex = needWait.Index
		}
	}
	// 全部点过
	if topIndex == -1 {
		return huList, -1, Pass, true
	}

	if topOpt == 0 { // 获得所有多人胡牌的列表
		for i := 0; i < len(rw.NeedWaitTool); i++ {
			if !rw.NeedWaitTool[i].IsPass {
				if rw.NeedWaitTool[i].CanTools[0] > 0 && rw.NeedWaitTool[i].Choice == 0 {
					huList = append(huList, rw.NeedWaitTool[i].Index)
					topIndex = rw.NeedWaitTool[i].Index
				}
				if rw.NeedWaitTool[i].CanTools[0] > 0 && rw.NeedWaitTool[i].Choice == -1 {
					// 只要有一个人没有操作
					return huList, -1, -1, false
				}
			}
		}
	}

	// 判断最高级别的用户是否操作
	nw := rw.GetOpt(topIndex)
	if nw != nil {
		//		fmt.Println("nw choice ", nw.Choice)
		//		fmt.Println("nw ", nw.CanTools)
		//		fmt.Println("nw index ", nw.Index)
	}
	if nw != nil && nw.Choice >= 0 { // 最高级别的用户已经操作了
		return huList, nw.Index, nw.Choice, true
	} else {
		return huList, -1, -1, false
	}
}
