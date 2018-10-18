package statement

import (
	al "PZ_GameServer/common/util/arrayList"
	"fmt"
	"strconv"
)

// 新建结算控制器
func NewStatement(mjType int32, uid []string) *StatementCtl {
	s := StatementCtl{
		GameType:  int(mjType),
		IDs:       uid,
		BaseScore: 500,
		Count:     0,
		Record:    al.New(),
	}
	return &s
}

// 初始化 , 可以在派生类中设置, 覆盖此方法, 这里提供基础类型
func (s *StatementCtl) Init() {
	s.Types = make(map[int]string)

	// 这里要初始化, 读配置文件
	s.AddType(T_SitDown, "坐下")
	s.AddType(T_StandUp, "站起来")
	s.AddType(T_Ready, "准备")
	s.AddType(T_Start, "开始")
	s.AddType(T_Dice, "筛子")
	s.AddType(T_Zhuang, "定庄家")
	s.AddType(T_Direct, "定方位")
	s.AddType(T_Deal, "发牌")
	s.AddType(T_Mo, "摸牌")
	s.AddType(T_MoBack, "摸牌(后面)")
	s.AddType(T_Put, "出牌")
	s.AddType(T_Kong, "杠")
	s.AddType(T_AnKong, "暗杠")
	s.AddType(T_PengKong, "补杠")
	s.AddType(T_Peng, "碰")
	s.AddType(T_Chow, "吃")
	s.AddType(T_Ting, "听")
	s.AddType(T_Hu, "胡")
	s.AddType(T_Hu_ZiMo, "自摸")
	s.AddType(T_Exit, "用户退出")
	s.AddType(T_Enter, "用户进入")
	s.AddType(T_Draw, "流局")
}

// 添加操作记录
func (s *StatementCtl) AddRecord(record OnceRecord) {
	s.Count++
	record.Index = s.Count
	s.Record.Add(&record)
	// s.DebugShowRecord()  //TODO : 这个方法有可能崩溃
}

// 添加操作
func (s *StatementCtl) AddTool(toolType int, index int, tindex int, val []int) {

	or := OnceRecord{
		Tool: &OnceTool{
			Index:    index,
			TIndex:   tindex,
			ToolType: toolType,
			Val:      val,
			MSG:      []string{s.GetMsg(toolType)},
		},
		Result: nil,
	}
	s.AddRecord(or)
}

// 添加初始手牌
func (s *StatementCtl) AddListCard(index int, listcard []int) {

	or := OnceRecord{
		Tool: &OnceTool{
			Index:    index,
			TIndex:   -1,
			ToolType: T_Deal,
			Val:      listcard,
			MSG:      []string{s.GetMsg(T_Deal)},
		},
		Result: nil,
	}
	s.AddRecord(or)

}

// 添加类型
func (s *StatementCtl) AddType(toolType int, msg string) {
	s.Types[toolType] = msg
}

// 得到描述
func (s *StatementCtl) GetMsg(toolType int) string {
	val, _ := s.Types[toolType]
	return val
}

// 显示操作记录
func (s *StatementCtl) DebugShowRecord() {

	if s.Record == nil {
		return
	}

	if *s.Record.Index(s.Record.Count - 1) == nil {
		return
	}

	record := (*s.Record.Index(s.Record.Count - 1)).(*OnceRecord) //TODO : 这里会崩溃  panic: runtime error: index out of range

	if record == nil {
		return
	}

	val := ""
	user := ""
	target := ""
	if record.Tool != nil {
		if record.Tool.Val != nil && len(record.Tool.Val) > 0 {
			val = s.GetValMsg(record.Tool.ToolType, record.Tool.Val)
		}

		if record.Tool.Index >= 0 {
			user = s.IDs[record.Tool.Index]
		}
		if record.Tool.TIndex >= 0 {
			target = s.IDs[record.Tool.TIndex]
		}
	}

	fmt.Println("-> 记录 : ", user, record.Tool.MSG, target, val)
}

func (s *StatementCtl) GetValMsg(toolType int, val []int) string {
	str := ""

	if toolType == T_Mo || toolType == T_Deal || toolType == T_Put || toolType == T_Chow || toolType == T_Kong || toolType == T_Peng || toolType == T_Hu || toolType == T_Hu_ZiMo {
		for _, v := range val {
			str += " " + GetMjNameForIndex(v)
		}
	} else {
		for _, v := range val {
			str += " " + strconv.Itoa(v)
		}
	}

	return str
}

// 判断全顺子
func (mh *StatementCtl) CheckQuanShunZi(pM []int) int {
	clen := len(pM)
	mjs := make([]int, clen)
	copy(mjs, pM)

	//递归退出条件 //获得牌的张数
	iNum := 0
	for i := 0; i < 27; i++ {
		iNum += mjs[i]
	}
	if iNum == 0 {
		return 1
	}

	//找到有牌的位置
	iIndex := 0
	for iIndex = 0; iIndex < 27; iIndex++ {
		if mjs[iIndex] > 0 {
			break
		}
	}

	//顺子判断
	if iIndex != 25 && iIndex != 26 {
		if mjs[iIndex] >= 1 && mjs[iIndex+1] >= 1 && mjs[iIndex+2] >= 1 && iIndex != 7 && iIndex != 8 && iIndex != 16 && iIndex != 17 {
			//减去顺子
			//Debug.Log("iIndex = " + iIndex + "  ");
			mjs[iIndex] -= 1
			mjs[iIndex+1] -= 1
			mjs[iIndex+2] -= 1

			//若余下的牌能够胡牌,返回1
			if 1 == mh.CheckQuanShunZi(mjs) {
				return 1
			} else { //若余下的牌不能胡牌,还原牌
				mjs[iIndex] += 1
				mjs[iIndex+1] += 1
				mjs[iIndex+2] += 1
			}
		}
	}
	return 0

}

// 胡牌判断
func (mh *StatementCtl) CheckHu(pM []int) int {

	clen := len(pM)
	duiCount := 0 // 对子

	// 去掉风箭刻子
	if clen > 27 {
		for i := 27; i < clen; i++ {
			if pM[i] == 1 {
				return 0 // 单张
			}
			if pM[i] == 2 {
				duiCount++ // 对子
				continue
			}
			if pM[i] == 3 { // 刻子
				pM[i] -= 3
				continue
			}
			if pM[i] > 3 && pM[i]%3 > 0 {
				return 0
			}

		}
	}

	//fmt.Println("duiCount ", duiCount)
	if duiCount > 1 { // 风字里超过1副将牌
		return 0
	}

	//递归退出条件 //获得牌的张数
	iNum := 0
	for i := 0; i < 27; i++ {
		iNum += pM[i]
	}
	if iNum == 0 {
		return 1
	}
	//	fmt.Println("iNum%3 ", iNum%3)
	//	if iNum%3 == 0 {
	//		return 0 // 没有将
	//	}

	//--- 判断7对
	for ii := 0; ii < 27; ii++ {
		if pM[ii] == 0 {
			continue
		}

		if pM[ii]%2 == 0 {
			duiCount++
		} else {
			continue
		}
	}

	//fmt.Println("duiCount_ ", duiCount)
	//	if duiCount == 7 {
	//		return 0 //7对
	//	}

	//找到有牌的位置
	iIndex := 0
	for iIndex = 0; iIndex < 27; iIndex++ {
		if pM[iIndex] > 0 {
			break
		}
	}

	//刻子判断
	if pM[iIndex] >= 3 {

		pM[iIndex] -= 3 //减去刻子

		//若余下的牌能够胡牌,返回1
		if 1 == mh.CheckHu(pM) {
			return 1
		} else { //若余下的牌不能胡牌,还原牌
			pM[iIndex] += 3
		}
	}
	//fmt.Println("jianuipanduan pm=", pM, iIndex, mh.SiChuan)
	//将对判断
	if pM[iIndex] >= 2 && mh.SiChuan == false {
		//减去将对,并设置标记
		pM[iIndex] -= 2
		mh.SiChuan = true
		//若余下的牌能够胡牌,返回1
		//fmt.Println("jianuipanduan pm=", pM)
		if 1 == mh.CheckHu(pM) {
			mh.SiChuan = false
			return 1
		} else { //若余下的牌不能胡牌,还原牌
			pM[iIndex] += 2
			mh.SiChuan = false
		}
	}

	//顺子判断
	if iIndex != 25 && iIndex != 26 {
		if pM[iIndex] >= 1 && pM[iIndex+1] >= 1 && pM[iIndex+2] >= 1 && iIndex != 7 && iIndex != 8 && iIndex != 16 && iIndex != 17 {
			//减去顺子
			//Debug.Log("iIndex = " + iIndex + "  ");
			pM[iIndex] -= 1
			pM[iIndex+1] -= 1
			pM[iIndex+2] -= 1

			//若余下的牌能够胡牌,返回1
			if 1 == mh.CheckHu(pM) {
				return 1
			} else { //若余下的牌不能胡牌,还原牌
				pM[iIndex] += 1
				pM[iIndex+1] += 1
				pM[iIndex+2] += 1
			}
		}
	}
	return 0

}

//得到麻将的位置
func GetMjIndex(t int, n int) int {

	if t < 0 || t > 5 || n < 0 || n > 9 {
		fmt.Println("GetMjIndex :错误的Index　", t, n)
		return -1
	}
	index := -1
	switch t {
	case W:
		index = n
	case B:
		index = 9 + n
	case T:
		index = 18 + n
	case F:
		index = 27 + n
	case J:
		index = 31 + n
	case H:
		index = 34 + n
	}
	return index
}

//得到麻将的类型数字
func GetMjTypeNum(index int) (int, int) {

	t := -1
	n := -1
	switch {
	case index >= 0 && index < 9:
		t = 0
		n = index
	case index > 8 && index < 18:
		t = 1
		n = index - 9
	case index > 17 && index < 27:
		t = 2
		n = index - 18
	case index > 26 && index < 31:
		t = 3
		n = index - 27
	case index > 30 && index < 34:
		t = 4
		n = index - 31
	case index > 33 && index <= 41:
		t = 5
		n = index - 34
	}

	return t, n
}

//得到麻将的名字
func GetMjNameForIndex(index int) string {
	if index < 0 {
		return ""
	}
	t, n := GetMjTypeNum(index)
	return GetMjName(t, n)
}

var Dict_wans = []string{"万", "筒", "条", "风", "", ""}
var Dict_nums = []string{"一", "二", "三", "四", "五", "六", "七", "八", "九"}
var Dict_fengs = []string{"东", "南", "西", "北"}
var Dict_jians = []string{"红中", "发财", "白板"}
var Dict_huas = []string{"春", "夏", "秋", "冬", "梅", "兰", "竹", "菊"}

//得到麻将的名字
func GetMjName(t int, n int) string {
	tp := ""
	num := ""
	switch t {
	case 0, 1, 2:
		if n >= 0 && n < 9 {
			tp = Dict_wans[t]
			num = Dict_nums[n]
		}
	case 3:
		if n >= 0 && n < 4 {
			tp = Dict_wans[t]
			num = Dict_fengs[n]
		}
	case 4:
		if n >= 0 && n < 3 {
			tp = Dict_wans[t]
			num = Dict_jians[n]
		}
	case 5:
		if n >= 0 && n < 8 {
			tp = Dict_wans[t]
			num = Dict_huas[n]
		}
	}
	return num + tp
}
