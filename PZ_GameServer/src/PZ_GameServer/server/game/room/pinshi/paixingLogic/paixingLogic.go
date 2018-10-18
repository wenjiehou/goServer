package paixingLogic

import (
	"fmt"
	"math"
	"sort"
)

type PockPaixing struct {
	Paixing    int
	NeedParePV bool //是否需要牌型的最大值
	PMaxValue  int  //牌型对应的最大值
	MaxValue   int
	MaxType    int
}

//扑克牌类型的定义
const (
	POCK_MEI   = 1   //梅花
	POCK_HEI   = 2   //黑桃
	POCK_FANG  = 3   //方块
	POCK_HONG  = 4   //红桃
	POCK_WANG  = 5   //王牌
	POCK_JISHU = 100 //扑克牌型之间的差值
)

//扑克牌型的定义
const (
	/**没牛*/
	PAIXING_MEINIU = 0
	/**牛1*/
	PAIXING_NIUYI = 1
	/**牛2*/
	PAIXING_NIUER = 2
	/**牛3*/
	PAIXING_NIUSAN = 3
	/**牛4*/
	PAIXING_NIUSI = 4
	/**牛5*/
	PAIXING_NIUWU = 5
	/**牛6*/
	PAIXING_NIULIU = 6
	/**牛7*/
	PAIXING_NIUQI = 7
	/**牛8*/
	PAIXING_NIUBA = 8
	/**牛9*/
	PAIXING_NIUJIU = 9
	/**牛牛*/
	PAIXING_NIUNIU = 10
	/**同花*/
	PAIXING_TONGHUA = 11
	/**顺子*/
	PAIXING_SHUNZI = 12
	/**葫芦*/
	PAIXING_HULU = 13
	/**五花牛*/
	PAIXING_WUHUA = 14
	/**五小牛*/
	PAIXING_WUXIAO = 15
	/**炸弹*/
	PAIXING_ZHADAN = 16
	/**同花顺*/
	PAIXING_TONGSHUN = 17
)

//获取扑克的牌型
func GetPockType(cid int) int {
	return int(math.Floor(float64(cid / 100)))
}

//按照黑4红3梅2方1返回
func GetNPockType(cid int) int {
	typeN := int(math.Floor(float64(cid / 100)))
	if typeN == 2 {
		typeN = 4
	} else if typeN == 4 {
		typeN = 3
	} else if typeN == 1 {
		typeN = 2
	} else if typeN == 3 {
		typeN = 1
	}
	return typeN
}

//获取扑克的牌值
func GetPockValue(cid int) int {
	return cid % 100

}

//12345特性规则
func GetPockValueShang(cid int) int {
	value := cid % 100

	if value == 14 {
		value = 1
	} else if value == 16 {
		value = 2
	}

	return value
}

//牛牛的值计算的特殊规则
func GetPockValueNN(cid int) int {
	value := cid % 100

	if value == 14 {
		value = 1
	} else if value == 16 {
		value = 2
	}
	if value >= 10 {
		value = 10
	}

	return value
}

//判断是否是花牌
func CheckIsHua(cid int) bool {
	value := cid % 100
	if value == 11 || value == 12 || value == 13 {
		return true
	}
	return false
}

//获取扑克的cid
func GetPockIndex(t int, n int) int {
	if t < POCK_MEI || t > POCK_WANG || n < 0 || n > 52 {
		fmt.Println("GetPockIndex :错误的Index　", t, n)
		return -1
	}
	index := -1
	switch t {
	case POCK_MEI:
		index = 1*POCK_JISHU + n
	case POCK_HEI:
		index = 2*POCK_JISHU + n
	case POCK_FANG:
		index = 3*POCK_JISHU + n
	case POCK_HONG:
		index = 4*POCK_JISHU + n
	case POCK_WANG:
		index = 5*POCK_JISHU + n
	}
	return index
}

//判断牌型是否加倍
func GetPockTypeBeishu(t int) int {
	//	if t == PAIXING_ZHADAN {
	//		return 2
	//	} else if t == PAIXING_DUIWANG {
	//		return 4
	//	} else {
	//		return 1
	//	}
	return 1
}

type IntSlice []int

func (a IntSlice) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a IntSlice) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a IntSlice) Less(i, j int) bool { // 重写 Less() 方法， 从小到大排序
	typeA := GetPockType(a[i])
	typeB := GetPockType(a[j])

	valueA := GetPockValue(a[i])
	valueB := GetPockValue(a[j])

	if valueA > valueB {
		return false
	} else if valueA == valueB {
		if typeA > typeB { //红桃在先
			return false
		} else {
			return true
		}
	} else {
		return true
	}
}

type NIntSlice []int

func (a NIntSlice) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a NIntSlice) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a NIntSlice) Less(i, j int) bool { // 重写 Less() 方法， 从小到大排序
	typeA := GetNPockType(a[i])
	typeB := GetNPockType(a[j])

	valueA := GetPockValueShang(a[i])
	valueB := GetPockValueShang(a[j])

	if valueA > valueB {
		return false
	} else if valueA == valueB {

		if typeA > typeB { // 黑2 红4 梅1 方3
			return false
		} else {
			return true
		}
	} else {
		return true
	}
}

//获取某个牌型的倍数
func GetPaxingBeishu(p *PockPaixing) int {
	if p.Paixing == PAIXING_TONGSHUN {
		return 6
	} else if p.Paixing == PAIXING_ZHADAN || p.Paixing == PAIXING_WUXIAO || p.Paixing == PAIXING_WUHUA {
		return 5
	} else if p.Paixing == PAIXING_HULU || p.Paixing == PAIXING_SHUNZI || p.Paixing == PAIXING_TONGHUA {
		return 4
	} else if p.Paixing == PAIXING_NIUNIU {
		return 3
	} else if p.Paixing == PAIXING_NIUJIU || p.Paixing == PAIXING_NIUBA {
		return 2
	} else {
		return 1
	}
}

//比较两个牌型的大小 返回true pa > pb
func CompPaixing(pa *PockPaixing, pb *PockPaixing) bool {
	if pa.Paixing > pb.Paixing {
		return true
	} else if pa.Paixing == pb.Paixing {
		if pa.NeedParePV == true { //这个需要比牌型中值的大小
			if pa.PMaxValue > pb.PMaxValue {
				return true
			} else if pa.PMaxValue == pb.PMaxValue { //只有同花顺和顺子需要
				//比较花色
				if pa.MaxType > pb.MaxType {
					return true
				} else {
					return false
				}
			} else {
				return false
			}
		} else { //不需要比牌型的值 直接比最大牌和花色
			if pa.MaxValue > pb.MaxValue {
				return true
			} else if pa.MaxValue == pb.MaxValue {
				if pa.MaxType > pb.MaxType {
					return true
				} else {
					return false
				}
			} else {
				return false
			}
		}
	} else {
		return false
	}
}

//获取牌的类型
func GetPaixing(cards []int) *PockPaixing {
	paixing := &PockPaixing{}
	maxCard := GetMaxVT(cards)
	paixing.MaxValue = GetPockValueShang(maxCard)
	paixing.MaxType = GetNPockType(maxCard)

	sort.Sort(IntSlice(cards))
	jishuData := *GetJishuArrData(cards)

	isTongHua := IsTonghua(cards, jishuData)

	isShun, maxSV := IsShunzi(cards, jishuData)
	isZha, maxZhaV := IsZhadan(cards, jishuData)
	isHulu, maxHuluV := IsHulu(cards, jishuData)

	if isTongHua == true && isShun == true {
		paixing.Paixing = PAIXING_TONGSHUN //同花顺
		paixing.NeedParePV = true
		paixing.PMaxValue = maxSV
	} else if isZha == true {
		paixing.Paixing = PAIXING_ZHADAN //炸弹
		paixing.NeedParePV = true
		paixing.PMaxValue = maxZhaV

	} else if IsWuxiao(cards, jishuData) == true {
		paixing.Paixing = PAIXING_WUXIAO //五小
		paixing.NeedParePV = false
	} else if IsWuhua(cards, jishuData) == true {
		paixing.Paixing = PAIXING_WUHUA
		paixing.NeedParePV = false
	} else if isHulu == true {
		paixing.Paixing = PAIXING_HULU
		paixing.NeedParePV = true
		paixing.PMaxValue = maxHuluV
	} else if isShun == true {
		paixing.Paixing = PAIXING_SHUNZI
		paixing.NeedParePV = true
		paixing.PMaxValue = maxSV
	} else if isTongHua == true {
		paixing.Paixing = PAIXING_TONGHUA
		paixing.NeedParePV = false
	} else {
		paixing.NeedParePV = false
		var niuData = GetNiuData(cards)
		if niuData.NiuNum == 0 {
			paixing.Paixing = PAIXING_MEINIU
		} else if niuData.NiuNum == 1 {
			paixing.Paixing = PAIXING_NIUYI
		} else if niuData.NiuNum == 2 {
			paixing.Paixing = PAIXING_NIUER
		} else if niuData.NiuNum == 3 {
			paixing.Paixing = PAIXING_NIUSAN
		} else if niuData.NiuNum == 4 {
			paixing.Paixing = PAIXING_NIUSI
		} else if niuData.NiuNum == 5 {
			paixing.Paixing = PAIXING_NIUWU
		} else if niuData.NiuNum == 6 {
			paixing.Paixing = PAIXING_NIULIU
		} else if niuData.NiuNum == 7 {
			paixing.Paixing = PAIXING_NIUQI
		} else if niuData.NiuNum == 8 {
			paixing.Paixing = PAIXING_NIUBA
		} else if niuData.NiuNum == 9 {
			paixing.Paixing = PAIXING_NIUJIU
		} else if niuData.NiuNum == 10 {
			paixing.Paixing = PAIXING_NIUNIU
		}
	}
	//fmt.Println("huoqupaixing zuihou:", paixing)
	return paixing
}

/**是否是同花顺*/
func IsTongShun(cards []int, jishuData JishuArrData) (bool, int) {
	if IsTonghua(cards, jishuData) == true {
		isShun, maxV := IsShunzi(cards, jishuData)
		if isShun == true {
			return true, maxV
		}
	}

	return false, -1
}

/**是否是炸弹*/
func IsZhadan(cards []int, jishuData JishuArrData) (bool, int) {
	if len(jishuData.FourArr) == 1 {
		maxV := 1
		if jishuData.FourArr[0] == 14 {
			maxV = 1
		} else if jishuData.FourArr[0] == 16 {
			maxV = 2
		} else {
			maxV = jishuData.FourArr[0]
		}

		return true, maxV
	}
	return false, -1
}

/**是否是五小*/
func IsWuxiao(cards []int, jishuData JishuArrData) bool {
	var totalV = 0
	for i := 0; i < len(jishuData.NValueArr); i++ {
		if jishuData.NValueArr[i] >= 5 { //大于等于5就不行啦

			return false
		}
		totalV += jishuData.NValueArr[i]
		if totalV >= 10 {
			return false
		}
	}
	return true
}

/**是否是五花牛*/
func IsWuhua(cards []int, jishuData JishuArrData) bool {
	for i := 0; i < len(jishuData.IsHuaArr); i++ {
		if jishuData.IsHuaArr[i] == false {
			return false
		}
	}
	return true
}

/**是否是葫芦*/
func IsHulu(cards []int, jishuData JishuArrData) (bool, int) {
	if len(jishuData.ThreeArr) == 1 && len(jishuData.TwoArr) == 1 {
		maxV := 1
		if jishuData.ThreeArr[0] == 14 {
			maxV = 1
		} else if jishuData.ThreeArr[0] == 16 {
			maxV = 2
		} else {
			maxV = jishuData.ThreeArr[0]
		}

		return true, maxV
	}
	return false, -1
}

func IsShunzi(cards []int, jishuData JishuArrData) (bool, int) {
	if len(jishuData.FourArr) == 0 && len(jishuData.ThreeArr) == 0 && len(jishuData.TwoArr) == 0 && len(jishuData.OneArr) >= 5 {
		oneArr := jishuData.OneArr //这个是拷贝的 go就这个鸟样子
		var idx = IndexOf(jishuData.OneArr, 16)
		if idx != -1 { //没有2 不行就不行//UserData.needOneTwo
			ConversValueArr(oneArr)
		}
		if CheckValueLian(oneArr) == true {
			return true, oneArr[len(oneArr)-1]
		}
	}

	return false, -1
}

/**是否为同花*/
func IsTonghua(cards []int, jishuData JishuArrData) bool {
	i := 0
	pk_type := -1
	clen := len(jishuData.TypeArr)
	for i = 0; i < clen; i++ {
		if i == 0 {
			pk_type = jishuData.TypeArr[i]
		} else if jishuData.TypeArr[i] != pk_type {
			return false
		}
	}
	return true
}

/**获取普通状况下 貌似只看这个*/
func GetMaxVT(cards []int) int {
	newCards := append(cards)
	sort.Sort(NIntSlice(newCards))
	return newCards[len(newCards)-1]
}
