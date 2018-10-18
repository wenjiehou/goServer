package paixingLogic

import (
	"fmt"
	"math"
	"sort"
)

//扑克牌类型的定义
const (
	POCK_MEI   = 1   //梅花
	POCK_HEI   = 2   //黑桃
	POCK_FANG  = 3   //方块
	POCK_HONG  = 4   //红桃
	POCK_WANG  = 5   //王牌
	POCK_JISHU = 100 //扑克牌型之间的差值
)

type PockPaixing struct {
	Paixing int
	Cards   []int
	Jishu   *JishuArrData
	Feiji   *FeijiData
	Sier    *SidaierData
}

//扑克牌型的定义
const (
	PAIXING_NON = -1
	//单张
	PAIXING_DAN = 1
	//对子
	PAIXING_DUIZI = 2
	//三个
	PAIXING_SAN = 3
	//三个带一
	PAIXING_SANYI = 4
	//三个带2
	PAIXING_SANER = 5
	//连对
	PAIXING_LIANDUI = 6
	//顺子
	PAIXING_SHUNZI = 7
	//飞机
	PAIXING_FEIJI = 8
	//四个带2
	PAIXING_SIDAIER = 9
	//炸弹
	PAIXING_ZHADAN = 10
	//对王
	PAIXING_DUIWANG = 11
)

//获取扑克的牌型
func GetPockType(cid int) int {
	return int(math.Floor(float64(cid / 100)))
}

//获取扑克的牌值
func GetPockValue(cid int) int {
	return cid % 100
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
	if t == PAIXING_ZHADAN {
		return 2
	} else if t == PAIXING_DUIWANG {
		return 4
	} else {
		return 1
	}
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

type IntRSlice []int

func (a IntRSlice) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a IntRSlice) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a IntRSlice) Less(i, j int) bool { // 重写 Less() 方法， 从小到大排序
	typeA := GetPockType(a[i])
	typeB := GetPockType(a[j])

	valueA := GetPockValue(a[i])
	valueB := GetPockValue(a[j])

	if valueA > valueB {
		return true
	} else if valueA == valueB {
		if typeA > typeB { //红桃在先
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

////获取牌的类型
func GetPaixing(cards []int) *PockPaixing {
	paixing := &PockPaixing{
		Paixing: -1,
	}

	sort.Sort(IntSlice(cards))
	jishuData := GetJishuArrData(cards)
	paixing.Cards = cards
	paixing.Jishu = jishuData

	if IsDan(cards) == true {
		paixing.Paixing = PAIXING_DAN
		return paixing
	} else if IsDui(cards) {
		paixing.Paixing = PAIXING_DUIZI
		return paixing
	} else if IsSan(cards, jishuData) == true {
		paixing.Paixing = PAIXING_SAN
		return paixing
	} else if IsSanYi(cards, jishuData) == true {
		paixing.Paixing = PAIXING_SANYI
		return paixing
	} else if IsSanEr(cards, jishuData) == true {
		paixing.Paixing = PAIXING_SANER
		return paixing
	} else if IsLiandui(cards, jishuData) == true {
		paixing.Paixing = PAIXING_LIANDUI
		return paixing
	} else if IsShun(cards, jishuData) == true {
		paixing.Paixing = PAIXING_SHUNZI
		return paixing
	} else {
		var feiji = IsFeiji(cards, jishuData)
		if feiji != nil {
			paixing.Paixing = PAIXING_FEIJI
			paixing.Feiji = feiji
			return paixing
		}

		var sier = IsSidaier(cards, jishuData)
		if sier != nil {
			paixing.Paixing = PAIXING_SIDAIER
			paixing.Sier = sier
			return paixing
		}

		if IsZha(cards, jishuData) == true {
			paixing.Paixing = PAIXING_ZHADAN
			return paixing
		} else if IsWangzha(cards, jishuData) == true {
			paixing.Paixing = PAIXING_DUIWANG
			return paixing
		}
	}

	return paixing
}

//是否是单张
func IsDan(cards []int) bool {
	if len(cards) == 1 {
		return true
	}
	return false
}

//是否是对子
func IsDui(cards []int) bool {
	if len(cards) == 2 {
		if GetPockValue(cards[0]) == GetPockValue(cards[1]) {
			return true
		}
	}
	return false
}

//是否是三个
func IsSan(cards []int, jishuData *JishuArrData) bool {
	if len(cards) == 3 && len(jishuData.ThreeArr) == 1 {
		return true
	}
	return false
}

//是否3+1
func IsSanYi(cards []int, jishuData *JishuArrData) bool {
	if len(cards) == 4 && len(jishuData.ThreeArr) == 1 && len(jishuData.OneArr) == 1 {
		return true
	}
	return false
}

//是否3+2
func IsSanEr(cards []int, jishuData *JishuArrData) bool {
	if len(cards) == 5 && len(jishuData.ThreeArr) == 1 && len(jishuData.TwoArr) == 1 {
		return true
	}
	return false
}

//是否连队
func IsLiandui(cards []int, jishuData *JishuArrData) bool {
	length := len(cards)
	if length >= 6 && length%2 == 0 {
		//		fmt.Println("jishuData", jishuData.TwoArr, cards)

		if len(jishuData.FourArr) == 0 && len(jishuData.ThreeArr) == 0 && len(jishuData.TwoArr) >= 3 && len(jishuData.OneArr) == 0 {
			if CheckValueLian(jishuData.TwoArr) == true {
				return true
			}
		}
	}
	return false
}

//是否顺子
func IsShun(cards []int, jishuData *JishuArrData) bool {
	if len(cards) >= 5 {
		if len(jishuData.FourArr) == 0 && len(jishuData.ThreeArr) == 0 && len(jishuData.TwoArr) == 0 && len(jishuData.OneArr) >= 5 {
			if CheckValueLian(jishuData.OneArr) == true {
				return true
			}
		}
	}
	return false

}

//是否是飞机
func IsFeiji(cards []int, jishuData *JishuArrData) *FeijiData {
	if len(cards) < 6 {
		return nil
	}

	return GetFeijiData(jishuData)
}

//是否是四带二
func IsSidaier(cards []int, jishuData *JishuArrData) *SidaierData {
	if len(cards) < 6 {
		return nil
	}
	return GetSidaierData(jishuData)

}

//是否是炸弹
func IsZha(cards []int, jishuData *JishuArrData) bool {
	if len(cards) != 4 {
		return false
	}

	if len(jishuData.FourArr) == 1 {
		return true
	}

	return false
}

//是否是对王
func IsWangzha(cards []int, jishuData *JishuArrData) bool {
	if len(cards) == 2 {
		if GetPockType(cards[0]) == POCK_WANG && GetPockType(cards[1]) == POCK_WANG {
			return true
		}
	}
	return false
}

//从一组牌里面找出大于当前牌的的牌
func GetBigthan(from []int, target []int) []int {
	var retArr []int = make([]int, 0)
	sort.Sort(IntSlice(from))
	if len(target) == 0 { //目标牌没有就给最小的第一张
		retArr = append(retArr, from[0])
		return retArr
	}

	//	sort.Sort(IntSlice(from))
	sort.Sort(IntSlice(target))

	fjishu := GetJishuArrData(from)
	tjishu := GetJishuArrData(target)

	tpaixing := GetPaixing(target)
	//	fmt.Println("tpaixing...", tpaixing)

	var tempV = 0
	var tempArr []int

	var i = 0
	var lianLen = 0

	switch tpaixing.Paixing {
	case PAIXING_NON:
		return retArr
	case PAIXING_DAN:
		tempV = tjishu.OneArr[0]
		//先从单张里面找一张
		for _, v := range fjishu.OneArr {
			if v > tempV {
				retArr = getCardsFromValues(from, []int{v})
				return retArr
			}
		}
		//单张没有找到
		for _, v := range fjishu.TwoArr {
			if v > tempV {
				retArr = getCardsFromValues(from, []int{v})
				return retArr
			}
		}
		//对子没找到
		for _, v := range fjishu.ThreeArr {
			if v > tempV {
				retArr = getCardsFromValues(from, []int{v})
				return retArr
			}
		}
		//三个也没找到
		for _, v := range fjishu.FourArr {
			if v > tempV {
				retArr = getCardsFromValues(from, []int{v})
				return retArr
			}
		}

		//单张都没有找到
		if len(fjishu.FourArr) > 0 {
			retArr = getCardsFromValues(from, []int{fjishu.FourArr[0], fjishu.FourArr[0], fjishu.FourArr[0], fjishu.FourArr[0]})
			return retArr
		}
		//对王不需要 如果有 一个王就干过了
		return retArr //空的表示没有找到咯

	case PAIXING_DUIZI:
		tempV = tjishu.TwoArr[0]
		for _, v := range fjishu.TwoArr {
			if v > tempV {
				retArr = getCardsFromValues(from, []int{v, v})
				return retArr
			}
		}

		for _, v := range fjishu.ThreeArr {
			if v > tempV {
				retArr = getCardsFromValues(from, []int{v, v})
				return retArr
			}
		}

		//找一个炸弹来
		if len(fjishu.FourArr) > 0 {
			v := fjishu.FourArr[0]
			retArr = getCardsFromValues(from, []int{v, v, v, v})
			return retArr
		}
		//看看有没有对王
		if checkHaveDuiWang(from) == true {
			retArr = []int{551, 552}
			return retArr
		}
		return retArr
	case PAIXING_SAN:
		tempV = tjishu.ThreeArr[0] //只有一个三个嘛
		//从自己的三个里面找
		for _, v := range fjishu.ThreeArr {
			if v > tempV {
				retArr = getCardsFromValues(from, []int{v, v, v})
				return retArr
			}
		}
		//找一个炸弹来
		if len(fjishu.FourArr) > 0 {
			v := fjishu.FourArr[0]
			retArr = getCardsFromValues(from, []int{v, v, v, v})
			return retArr
		}
		//看看有没有对王
		if checkHaveDuiWang(from) == true {
			retArr = []int{551, 552}
			return retArr
		}
		return retArr
	case PAIXING_SANYI:
		tempV = tjishu.ThreeArr[0]
		for i = 0; i < len(fjishu.ThreeArr); i++ {

			if fjishu.ThreeArr[i] > tempV { //就这张了
				threeV := fjishu.ThreeArr[i]
				if len(fjishu.OneArr) > 0 {
					retArr = getCardsFromValues(from, []int{threeV, threeV, threeV, fjishu.OneArr[0]})
					return retArr
				} else if len(fjishu.TwoArr) > 0 {
					retArr = getCardsFromValues(from, []int{threeV, threeV, threeV, fjishu.TwoArr[0]})
					return retArr
				} else if len(fjishu.ThreeArr) > 1 {
					if i == 0 {
						retArr = getCardsFromValues(from, []int{threeV, threeV, threeV, fjishu.ThreeArr[1]})
					} else {
						retArr = getCardsFromValues(from, []int{threeV, threeV, threeV, fjishu.ThreeArr[0]})
					}
					return retArr
				}
			}
		}

		//找一个炸弹来
		if len(fjishu.FourArr) > 0 {
			v := fjishu.FourArr[0]
			retArr = getCardsFromValues(from, []int{v, v, v, v})
			return retArr
		}
		//看看有没有对王
		if checkHaveDuiWang(from) == true {
			retArr = []int{551, 552}
			return retArr
		}
		return retArr

	case PAIXING_SANER:
		tempV = tjishu.ThreeArr[0]
		for i = 0; i < len(fjishu.ThreeArr); i++ {
			if fjishu.ThreeArr[i] > tempV {
				threeV := fjishu.ThreeArr[i]
				if len(fjishu.TwoArr) > 0 {
					retArr = getCardsFromValues(from, []int{threeV, threeV, threeV, fjishu.TwoArr[0], fjishu.TwoArr[0]})
					return retArr
				} else if len(fjishu.ThreeArr) > 1 {
					if i == 0 {
						retArr = getCardsFromValues(from, []int{threeV, threeV, threeV, fjishu.ThreeArr[1], fjishu.ThreeArr[1]})
					} else {
						retArr = getCardsFromValues(from, []int{threeV, threeV, threeV, fjishu.ThreeArr[0], fjishu.ThreeArr[0]})
					}
					return retArr
				}
			}
		}
		//找一个炸弹来
		if len(fjishu.FourArr) > 0 {
			v := fjishu.FourArr[0]
			retArr = getCardsFromValues(from, []int{v, v, v, v})
			return retArr
		}
		//看看有没有对王
		if checkHaveDuiWang(from) == true {
			retArr = []int{551, 552}
			return retArr
		}
		return retArr
	case PAIXING_LIANDUI:
		lianLen = len(tjishu.TwoArr)
		tempV = tjishu.TwoArr[lianLen-1]

		moreTwoArr := append(fjishu.ThreeArr, fjishu.TwoArr...)
		sort.Sort(IntSlice(moreTwoArr))
		//去除不符合的对子
		for i = 0; i < len(moreTwoArr); i++ {
			if moreTwoArr[i] < tempV-(lianLen-2) {
				moreTwoArr = append(moreTwoArr[:i], moreTwoArr[i+1:]...)
				i--
			}
		}
		//剩下的长度要大于连队的长度 从中找出一个连续的长度
		if len(moreTwoArr) >= lianLen {
			tempArr = getFromVlauesLenThan(moreTwoArr, lianLen)
			if tempArr != nil && len(tempArr) > 0 { //找到咯
				tempArr = copyNumEle(tempArr, 2)
				sort.Sort(IntRSlice(tempArr))
				retArr = getCardsFromValues(from, tempArr)
				return retArr
			}
		}
		//找一个炸弹来
		if len(fjishu.FourArr) > 0 {
			v := fjishu.FourArr[0]
			retArr = getCardsFromValues(from, []int{v, v, v, v})
			return retArr
		}
		//看看有没有对王
		if checkHaveDuiWang(from) == true {
			retArr = []int{551, 552}

			return retArr
		}
		return retArr

	case PAIXING_SHUNZI:
		lianLen = len(tjishu.OneArr)
		tempV = tjishu.OneArr[lianLen-1]

		moreOneArr := append(fjishu.ThreeArr, fjishu.TwoArr...)
		moreOneArr = append(moreOneArr, fjishu.OneArr...)

		sort.Sort(IntSlice(moreOneArr))

		for i = 0; i < len(moreOneArr); i++ {
			if moreOneArr[i] < tempV-(lianLen-2) {
				moreOneArr = append(moreOneArr[:i], moreOneArr[i+1:]...)
				i--
			}
		}
		if len(moreOneArr) >= lianLen {
			tempArr = getFromVlauesLenThan(moreOneArr, lianLen)
			if tempArr != nil && len(tempArr) > 0 { //找到咯
				sort.Sort(IntRSlice(tempArr))
				retArr = getCardsFromValues(from, tempArr)
				return retArr
			}
		}

		//找一个炸弹来
		if len(fjishu.FourArr) > 0 {
			v := fjishu.FourArr[0]
			retArr = getCardsFromValues(from, []int{v, v, v, v})
			return retArr
		}
		//看看有没有对王
		if checkHaveDuiWang(from) == true {
			retArr = []int{551, 552}

			return retArr
		}
		return retArr

	case PAIXING_FEIJI:
		var feiji = tpaixing.Feiji
		if feiji == nil { //
			return retArr
		}

		threeArr := copyNumEle(fjishu.ThreeArr, 1)
		spliceArr := make([]int, 0)

		for i = 0; i < len(threeArr); i++ {
			if threeArr[i] < feiji.BigValue-(feiji.Length-2) {
				spliceArr = append(spliceArr, threeArr[i])
				threeArr = append(threeArr[:i], threeArr[i+1:]...)
				i--

			}
		}

		if len(threeArr) >= feiji.Length {
			tempArr = getFromVlauesLenThan(threeArr, feiji.Length)
			fmt.Println("tempArr ...", tempArr)
			ctempArr := copyNumEle(tempArr, 1)
			//将三个剩余的继续加入删除的备用数组
			for i = 0; i < len(threeArr); i++ {
				idx := IndexOf(ctempArr, threeArr[i])
				if idx != -1 {
					ctempArr = append(ctempArr[:idx], ctempArr[idx+1:]...)
					threeArr = append(threeArr[:i], threeArr[i+1:]...)
					fmt.Println("tempArr ...111", tempArr, ctempArr, i)
					i--
				}
			}
			spliceArr = append(spliceArr, threeArr...)
			fmt.Println("spliceArr ...111", tempArr, spliceArr)

			if tempArr != nil && len(tempArr) > 0 {
				fmt.Println("tempArr ...222", tempArr, ctempArr)
				switch feiji.Type { //飞机不带翅膀
				case FEIJI_NON:
					tempArr = copyNumEle(tempArr, 3)
					retArr = getCardsFromValues(from, tempArr)
					return retArr
				case FEIJI_DAN: //飞机带一个翅膀
					if feiji.Length <= 3*len(spliceArr)+2*len(fjishu.TwoArr)+len(fjishu.OneArr) {

						tempArr = copyNumEle(tempArr, 3)
						tuiArr := make([]int, 0)
						for i = 0; i < len(fjishu.OneArr); i++ { //拿单张补腿
							tuiArr = append(tuiArr, fjishu.OneArr[i])
							if len(tuiArr) == feiji.Length {
								break
							}
						}
						if len(tuiArr) < feiji.Length { //拿对子来补腿
							for i = 0; i < len(fjishu.TwoArr); i++ {
								tuiArr = append(tuiArr, fjishu.TwoArr[i])
								if len(tuiArr) == feiji.Length {
									break
								}
								tuiArr = append(tuiArr, fjishu.TwoArr[i])
								if len(tuiArr) == feiji.Length {
									break
								}
							}
						}

						if len(tuiArr) < feiji.Length { //拿三个来补
							for i = 0; i < len(spliceArr); i++ {
								tuiArr = append(tuiArr, spliceArr[i])
								if len(tuiArr) == feiji.Length {
									break
								}
								tuiArr = append(tuiArr, spliceArr[i])
								if len(tuiArr) == feiji.Length {
									break
								}
								tuiArr = append(tuiArr, spliceArr[i])
								if len(tuiArr) == feiji.Length {
									break
								}
							}
						}

						if len(tuiArr) == feiji.Length {
							sort.Sort(IntSlice(tuiArr))

							retArr = getCardsFromValues(from, append(tempArr, tuiArr...))
							return retArr
						}

					}
				case FEIJI_DUI:
					if feiji.Length <= len(spliceArr)+len(fjishu.TwoArr) {
						tempArr = copyNumEle(tempArr, 3)
						tuiArr := make([]int, 0)
						for i = 0; i < len(fjishu.TwoArr); i++ {
							tuiArr = append(tuiArr, fjishu.TwoArr[i])
							if len(tuiArr) == feiji.Length {
								break
							}
						}

						if len(tuiArr) < feiji.Length {
							for i = 0; i < len(spliceArr); i++ {
								tuiArr = append(tuiArr, spliceArr[i])
								if len(tuiArr) == feiji.Length {
									break
								}
							}
						}

						if len(tuiArr) == feiji.Length {
							tuiArr = copyNumEle(tuiArr, 2)
							retArr = getCardsFromValues(from, append(tempArr, tuiArr...))
						}

					}

				}
			}

		}

		//找一个炸弹来
		if len(fjishu.FourArr) > 0 {
			v := fjishu.FourArr[0]
			retArr = getCardsFromValues(from, []int{v, v, v, v})
			return retArr
		}
		//看看有没有对王
		if checkHaveDuiWang(from) == true {
			retArr = []int{551, 552}

			return retArr
		}
		return retArr
	case PAIXING_SIDAIER:
		var sier = tpaixing.Sier
		var fsier = GetSidaierData(fjishu) //这个玩意必须正好是四带二，不然返回的是nil

		if fsier != nil && fsier.Type == sier.Type && fsier.Value > sier.Value {
			retArr = copyNumEle(from, 1)
			return retArr
		}
		//找一个炸弹来
		if len(fjishu.FourArr) > 0 {
			v := fjishu.FourArr[0]
			retArr = getCardsFromValues(from, []int{v, v, v, v})
			return retArr
		}
		//看看有没有对王
		if checkHaveDuiWang(from) == true {
			retArr = []int{551, 552}

			return retArr
		}
		return retArr

	case PAIXING_ZHADAN:
		tempV = tjishu.FourArr[0]
		if len(fjishu.FourArr) > 0 {
			for i = 0; i < len(fjishu.FourArr); i++ {
				if fjishu.FourArr[i] > tempV {
					retArr = getCardsFromValues(from, []int{fjishu.FourArr[i], fjishu.FourArr[i], fjishu.FourArr[i], fjishu.FourArr[i]})
					return retArr
				}
			}
		}
		//看看有没有对王
		if checkHaveDuiWang(from) == true {
			retArr = []int{551, 552}

			return retArr
		}
		return retArr

	case PAIXING_DUIWANG:
		return retArr

	}

	return retArr
}

//从一组牌获得values对应cards
func getCardsFromValues(from []int, values []int) []int {
	var retArr = make([]int, 0)
	var tempArr = copyNumEle(from, 1)
	for i := 0; i < len(values); i++ {
		for j := 0; j < len(tempArr); j++ {
			if GetPockValue(tempArr[j]) == GetPockValue(values[i]) {
				retArr = append(retArr, tempArr[j])
				tempArr = append(tempArr[:j], tempArr[j+1:]...)
				break
			}
		}
	}
	return retArr
}

//从一组数值中找出一个连续的长度等于指定的长度
func getFromVlauesLenThan(values []int, tlen int) []int {
	var retArr = make([]int, 0)
	var i = 0
	var num = 0
	for i = 0; i < len(values)-1; i++ {
		if values[i+1] == values[i]+1 {
			if num == 0 {
				retArr = append(retArr, values[i], values[i+1])
				num += 2
			} else {
				retArr = append(retArr, values[i+1])
				num += 1
			}
		} else {
			num = 0
			retArr = make([]int, 0)
		}

		if num == tlen {
			break
		}
	}

	if len(retArr) < tlen { //没有达到要求
		retArr = make([]int, 0)
	}
	return retArr

}

//获取数组连续的长度
func getLianLen(values []int) []int {
	var retArr = make([]int, 0)
	var tempArr = make([]int, 0)
	var i = 0
	var num = 0
	var temp = 0
	for i = 0; i < len(values)-1; i++ {
		if values[i+1] == values[i]+1 {
			if temp == 0 {
				tempArr = append(tempArr, values[i], values[i+1])
				temp += 2
			} else {
				tempArr = append(tempArr, values[i+1])
				temp += 1
			}
		} else {
			if temp > num {
				num = temp
				retArr = copyNumEle(tempArr, 1)
			}
			temp = 0
			tempArr = make([]int, 0)
		}
	}

	return retArr
}

//检测一组牌里面有没有对王
func checkHaveDuiWang(cards []int) bool {
	var xiaoW = false
	var daW = false
	for _, v := range cards {
		if v == 551 {
			xiaoW = true
		}
		if v == 552 {
			daW = true
		}

		if xiaoW == true && daW == true {
			return true
		}
	}
	return false
}

//将没个元素拷贝几份
func copyNumEle(arr []int, num int) []int {
	var retArr = make([]int, 0)
	var i = 0
	var j = 0
	for i = 0; i < len(arr); i++ {
		for j = 0; j < num; j++ {
			retArr = append(retArr, arr[i])
		}
	}
	return retArr
}

//从一个数组中删除对应的元素
func deleEleFromArr(from []int, eles []int) []int {
	var i = 0

	celes := copyNumEle(eles, 1)

	for i = 0; i < len(from); i++ {
		var idx = IndexOf(celes, from[i])
		if idx != -1 {
			eles = append(celes[:idx], celes[idx+1:]...)
			from = append(from[:i], from[i+1:]...)
			i--
		}
	}

	return from

}
