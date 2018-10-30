package paixingLogic

import (
	"sort"
)

//牌型组合数据结构
type CardGroupData struct {
	Paixing *PockPaixing //牌的类型
	Value   int          //该牌的价值
	Count   int          //含牌的个数
	MaxCard int          //牌中决定大小的牌值，用于对比
}

//手牌权值结构
type HandCardValue struct {
	SumValue  int //手牌总价值
	NeedRound int // 需要打几手牌
}

func get_GroupData(paixing *PockPaixing) *CardGroupData {
	var groupData = &CardGroupData{}
	groupData.Paixing = paixing
	groupData.Count = len(paixing.Cards)
	switch paixing.Paixing {
	case PAIXING_NON: //不出或者不合法
		groupData.MaxCard = 0
		groupData.Value = 0
	case PAIXING_DAN: //单张
		groupData.MaxCard = paixing.Jishu.OneArr[0]
		groupData.Value = groupData.MaxCard - 10
		if groupData.MaxCard == 51 || groupData.MaxCard == 52 {
			groupData.Value = groupData.Value - 34
		}
	case PAIXING_DUIZI: //对子
		groupData.MaxCard = paixing.Jishu.TwoArr[0]
		groupData.Value = groupData.MaxCard - 10
	case PAIXING_SAN: //三条
		groupData.MaxCard = paixing.Jishu.ThreeArr[0]
		groupData.Value = groupData.MaxCard - 10
	case PAIXING_SHUNZI: //顺子
		groupData.MaxCard = paixing.Jishu.OneArr[len(paixing.Jishu.OneArr)-1]
		groupData.Value = groupData.MaxCard - 10 + 1
	case PAIXING_LIANDUI: //连队
		groupData.MaxCard = paixing.Jishu.TwoArr[len(paixing.Jishu.TwoArr)-1]
		groupData.Value = groupData.MaxCard - 10 + 1
	case PAIXING_FEIJI: //飞机 这里后面还要考虑飞机不带 飞机带一 飞机带二 ，暂时抄袭人家的就不管了
		groupData.MaxCard = paixing.Jishu.ThreeArr[len(paixing.Jishu.ThreeArr)-1]
		if paixing.Feiji.Type == FEIJI_NON { //三连 暂时认为三种情况价值一样，其实要出那种牌，更取决于手牌的组合数目
			groupData.Value = (groupData.MaxCard - 3 + 1) / 2
		} else if paixing.Feiji.Type == FEIJI_DAN { //三带一连
			groupData.Value = (groupData.MaxCard - 3 + 1) / 2

		} else if paixing.Feiji.Type == FEIJI_DUI { //三带二连
			groupData.Value = (groupData.MaxCard - 3 + 1) / 2
		}
	case PAIXING_SANYI: //三带一 屁股问题在最佳组合的地方处理过了
		groupData.MaxCard = paixing.Jishu.ThreeArr[0]
		if groupData.MaxCard >= 12 {
			groupData.Value = 2 * (groupData.MaxCard - 10)
		} else {
			groupData.Value = groupData.MaxCard - 10
		}

	case PAIXING_SANER: //三带二
		groupData.MaxCard = paixing.Jishu.ThreeArr[0]
		if groupData.MaxCard >= 12 {
			groupData.Value = 2 * (groupData.MaxCard - 10)
		} else {
			groupData.Value = groupData.MaxCard - 10
		}
	case PAIXING_SIDAIER: //四带二
		groupData.MaxCard = paixing.Jishu.FourArr[0]
		if paixing.Sier.Type == SIDAIER_DAN { //四带两单
			groupData.Value = (groupData.MaxCard - 3) / 2
		} else if paixing.Sier.Type == SIDAIER_DUI { //四带两对
			groupData.Value = (groupData.MaxCard - 3) / 2
		}
	case PAIXING_ZHADAN: //炸弹
		groupData.MaxCard = paixing.Jishu.FourArr[0]
		groupData.Value = groupData.MaxCard - 1 + 7
	case PAIXING_DUIWANG: //王炸
		groupData.MaxCard = 52 //大王
		groupData.Value = 20
	default:
		groupData.Value = 0

	}

	return groupData
}

func get_HandCardValue(cards []int, paixing *PockPaixing) *HandCardValue {
	var handValue = &HandCardValue{
		NeedRound: 0,
		SumValue:  0,
	}

	var tpaixing *PockPaixing

	if paixing == nil {
		tpaixing = GetPaixing(cards)
	} else {
		tpaixing = paixing
	}

	if tpaixing.Paixing != PAIXING_NON && tpaixing.Paixing != PAIXING_SIDAIER {
		handValue.NeedRound = 1
		handValue.SumValue = get_GroupData(tpaixing).Value
		return handValue
	}
	//todo
	bestValue, _ := getBestList(cards)

	return bestValue

}

//目前只排列，以后会加入各种特殊情况
func getBestList(cards []int) (*HandCardValue, [][]int) {
	retList1 := get_PutCardList(cards, 1)
	retList2 := get_PutCardList(cards, 2)

	var tempValue1 = &HandCardValue{
		NeedRound: 0,
		SumValue:  0,
	}

	var tempValue2 = &HandCardValue{
		NeedRound: 0,
		SumValue:  0,
	}

	for i := 0; i < len(retList1); i++ {
		tempp := GetPaixing(retList1[i])
		if tempp.Paixing == PAIXING_NON {
			continue
		} else {
			tempg := get_GroupData(tempp)
			tempValue1.NeedRound += 1
			tempValue1.SumValue += tempg.Value
		}
	}

	for i := 0; i < len(retList2); i++ {
		tempp := GetPaixing(retList2[i])
		if tempp.Paixing == PAIXING_NON {
			continue
		} else {
			tempg := get_GroupData(tempp)
			tempValue2.NeedRound += 1
			tempValue2.SumValue += tempg.Value
		}
	}

	if tempValue1.SumValue-tempValue1.NeedRound*7 <= tempValue2.SumValue-tempValue2.NeedRound*7 { //取tempArr1
		return tempValue1, retList1
	} else {
		return tempValue2, retList2
	}
}

//获取当前当前出牌的最优方案
func get_PutCardList(cards []int, order int) [][]int {

	tcards := copyNumEle(cards, 1)
	paixing := GetPaixing(tcards)

	retList := [][]int{}

	if paixing.Paixing != PAIXING_NON { //一手出完就跑吧
		retList = append(retList, tcards)
		return retList
	}

	if checkDuiwangAndOne(cards) == true {
		retList = append(retList, []int{551, 552})
		tcards = deleEleFromArr(tcards, []int{551, 552})
		retList = append(retList, tcards)
		return retList
	}

	fourArr := paixing.Jishu.FourArr
	//我们先把four踢出掉

	fv := copyNumEle(fourArr, 4)
	fc := getCardsFromValues(tcards, fv)

	tcards = deleEleFromArr(tcards, fc) //最后记得补上4个的就好了

	//如果有对王，先去掉对王
	var idxX = IndexOf(tcards, 551)
	var idxD = IndexOf(tcards, 552)

	var haveDuiwang = false

	if idxX != -1 && idxD != -1 { //有对王
		tcards = deleEleFromArr(tcards, []int{551, 552})
		haveDuiwang = true
	}

	zhadanArr := [][]int{}
	for i := 0; i < len(fc)/4; i++ {
		zhadanArr = append(zhadanArr, []int{fc[i*4], fc[i*4+1], fc[i*4+2], fc[i*4+3]})
	}

	if haveDuiwang == true {
		zhadanArr = append(zhadanArr, []int{551, 552})
	}

	if order == 1 {
		for { //这里我们把三个和三个带一，三个带二都搞完了
			tempArr := GetBigthan(tcards, []int{1, 1, 1})
			tempArr1 := GetBigthan(tcards, []int{1, 1, 1, 2})
			tempArr2 := GetBigthan(tcards, []int{1, 1, 1, 2, 2})

			tempJishu1 := GetJishuArrData(tempArr1)
			tempJishu2 := GetJishuArrData(tempArr2)

			if len(tempArr2) > 0 { //三个都有
				tmpPaixing := GetPaixing(tcards)
				if tmpPaixing.Paixing != PAIXING_NON { //剩下的就一手了
					retList = append(retList, tcards)
					//把炸弹补回来
					retList = append(retList, zhadanArr...)

					return retList
				}

				if tempJishu1.OneArr[0] >= tempJishu2.TwoArr[0] { //单张比对子大，一样大的话，就带对子好了
					if tempJishu2.TwoArr[0] >= 14 { //出三个不带
						retList = append(retList, tempArr)
						tcards = deleEleFromArr(tcards, tempArr)
					} else { //出三带2
						retList = append(retList, tempArr2)
						tcards = deleEleFromArr(tcards, tempArr2)
					}

				} else {
					if tempJishu1.OneArr[0] >= 16 {
						retList = append(retList, tempArr)
						tcards = deleEleFromArr(tcards, tempArr)
					} else { //出三带1
						retList = append(retList, tempArr1)
						tcards = deleEleFromArr(tcards, tempArr1)
					}

				}
			} else if len(tempArr1) > 0 {
				if tempJishu1.OneArr[0] >= 16 {
					retList = append(retList, tempArr)
					tcards = deleEleFromArr(tcards, tempArr)
				} else { //出三带1
					retList = append(retList, tempArr1)
					tcards = deleEleFromArr(tcards, tempArr1)
				}
			} else if len(tempArr) > 0 {
				retList = append(retList, tempArr)
				tcards = deleEleFromArr(tcards, tempArr)
			} else {
				break
			}
		}

		for { //从剩下的牌里面搞顺子
			tmpPaixing := GetPaixing(tcards)
			if tmpPaixing.Paixing != PAIXING_NON { //剩下的就一手了
				retList = append(retList, tcards)
				retList = append(retList, zhadanArr...)
				return retList
			}

			lianArr := []int{}
			for i := 0; i < len(tmpPaixing.Jishu.ValueArr)-1; i++ {
				if tmpPaixing.Jishu.ValueArr[i+1] == tmpPaixing.Jishu.ValueArr[i] {
					continue
				} else if tmpPaixing.Jishu.ValueArr[i+1] == tmpPaixing.Jishu.ValueArr[i]+1 {
					if len(lianArr) == 0 {
						lianArr = append(lianArr, tmpPaixing.Jishu.ValueArr[i], tmpPaixing.Jishu.ValueArr[i+1])
					} else {
						lianArr = append(lianArr, tmpPaixing.Jishu.ValueArr[i+1])
					}
				} else {
					if len(lianArr) >= 5 { //大于五个组成连子咯
						break
					} else {
						lianArr = []int{}
					}
				}
			}

			if len(lianArr) >= 5 {

				tmpCards := getCardsFromValues(tcards, lianArr)

				retList = append(retList, tmpCards)
				tcards = deleEleFromArr(tcards, tmpCards)

			} else { //没有连字咯
				break
			}
		}

		//这个时候可能还有连队来
		for {
			tmpPaixing := GetPaixing(tcards)
			if tmpPaixing.Paixing != PAIXING_NON { //剩下的就一手了
				retList = append(retList, tcards)
				retList = append(retList, zhadanArr...)
				return retList
			}

			//我们取出所有连队，由于前面取过三个头，所以这里只需考虑剩下的对子就好了
			twoArr := tmpPaixing.Jishu.TwoArr
			lianArr := []int{}
			for i := 0; i < len(twoArr)-1; i++ {
				if twoArr[i] == twoArr[i+1]-1 {
					if len(lianArr) == 0 {
						lianArr = append(lianArr, twoArr[i], twoArr[i+1])
					} else {
						lianArr = append(lianArr, twoArr[i+1])
					}
				} else {
					if len(lianArr) >= 3 {
						break
					} else {
						lianArr = []int{}
					}
				}

			}

			if len(lianArr) >= 3 {
				tmpCards := getCardsFromValues(tcards, copyNumEle(lianArr, 2))
				retList = append(retList, tmpCards)
				tcards = deleEleFromArr(tcards, tmpCards)
			} else { //没有连字咯
				break
			}

		}

	} else if order == 2 {
		for { //从剩下的牌里面搞顺子
			tmpPaixing := GetPaixing(tcards)
			if tmpPaixing.Paixing != PAIXING_NON { //剩下的就一手了
				retList = append(retList, tcards)
				retList = append(retList, zhadanArr...)
				return retList
			}

			lianArr := []int{}
			for i := 0; i < len(tmpPaixing.Jishu.ValueArr)-1; i++ {
				if tmpPaixing.Jishu.ValueArr[i+1] == tmpPaixing.Jishu.ValueArr[i] {
					continue
				} else if tmpPaixing.Jishu.ValueArr[i+1] == tmpPaixing.Jishu.ValueArr[i]+1 {
					if len(lianArr) == 0 {
						lianArr = append(lianArr, tmpPaixing.Jishu.ValueArr[i], tmpPaixing.Jishu.ValueArr[i+1])
					} else {
						lianArr = append(lianArr, tmpPaixing.Jishu.ValueArr[i+1])
					}
				} else {
					if len(lianArr) >= 5 { //大于五个组成连子咯
						break
					} else {
						lianArr = []int{}
					}
				}
			}

			if len(lianArr) >= 5 {
				tmpCards := getCardsFromValues(tcards, lianArr)
				retList = append(retList, tmpCards)
				tcards = deleEleFromArr(tcards, tmpCards)

			} else { //没有连字咯
				break
			}
		}

		for { //这里我们把三个和三个带一，三个带二都搞完了
			tempArr := GetBigthan(tcards, []int{1, 1, 1})
			tempArr1 := GetBigthan(tcards, []int{1, 1, 1, 2})
			tempArr2 := GetBigthan(tcards, []int{1, 1, 1, 2, 2})

			tempJishu1 := GetJishuArrData(tempArr1)
			tempJishu2 := GetJishuArrData(tempArr2)

			if len(tempArr2) > 0 { //三个都有
				tmpPaixing := GetPaixing(tcards)
				if tmpPaixing.Paixing != PAIXING_NON { //剩下的就一手了
					retList = append(retList, tcards)
					//把炸弹补回来
					retList = append(retList, zhadanArr...)

					return retList
				}

				if tempJishu1.OneArr[0] >= tempJishu2.TwoArr[0] { //单张比对子大，一样大的话，就带对子好了
					if tempJishu2.TwoArr[0] >= 14 { //出三个不带
						retList = append(retList, tempArr)
						tcards = deleEleFromArr(tcards, tempArr)
					} else { //出三带2
						retList = append(retList, tempArr2)
						tcards = deleEleFromArr(tcards, tempArr2)
					}

				} else {
					if tempJishu1.OneArr[0] >= 16 {
						retList = append(retList, tempArr)
						tcards = deleEleFromArr(tcards, tempArr)
					} else { //出三带1
						retList = append(retList, tempArr1)
						tcards = deleEleFromArr(tcards, tempArr1)
					}

				}
			} else if len(tempArr1) > 0 {
				if tempJishu1.OneArr[0] >= 16 {
					retList = append(retList, tempArr)
					tcards = deleEleFromArr(tcards, tempArr)
				} else { //出三带1
					retList = append(retList, tempArr1)
					tcards = deleEleFromArr(tcards, tempArr1)
				}
			} else if len(tempArr) > 0 {
				retList = append(retList, tempArr)
				tcards = deleEleFromArr(tcards, tempArr)
			} else {
				break
			}
		}

		//这个时候可能还有连队来
		for {
			tmpPaixing := GetPaixing(tcards)
			if tmpPaixing.Paixing != PAIXING_NON { //剩下的就一手了
				retList = append(retList, tcards)
				retList = append(retList, zhadanArr...)
				return retList
			}

			//我们取出所有连队，由于前面取过三个头，所以这里只需考虑剩下的对子就好了
			twoArr := tmpPaixing.Jishu.TwoArr
			lianArr := []int{}
			for i := 0; i < len(twoArr)-1; i++ {
				if twoArr[i] == twoArr[i+1]-1 {
					if len(lianArr) == 0 {
						lianArr = append(lianArr, twoArr[i], twoArr[i+1])
					} else {
						lianArr = append(lianArr, twoArr[i+1])
					}
				} else {
					if len(lianArr) >= 3 {
						break
					} else {
						lianArr = []int{}
					}
				}

			}

			if len(lianArr) >= 3 {
				tmpCards := getCardsFromValues(tcards, copyNumEle(lianArr, 2))
				retList = append(retList, tmpCards)
				tcards = deleEleFromArr(tcards, tmpCards)
			} else { //没有连字咯
				break
			}

		}
	} else if order == 3 {

		for { //从剩下的牌里面搞顺子
			tmpPaixing := GetPaixing(tcards)
			if tmpPaixing.Paixing != PAIXING_NON { //剩下的就一手了
				retList = append(retList, tcards)
				retList = append(retList, zhadanArr...)
				return retList
			}

			lianArr := []int{}
			for i := 0; i < len(tmpPaixing.Jishu.ValueArr)-1; i++ {
				if tmpPaixing.Jishu.ValueArr[i+1] == tmpPaixing.Jishu.ValueArr[i] {
					continue
				} else if tmpPaixing.Jishu.ValueArr[i+1] == tmpPaixing.Jishu.ValueArr[i]+1 {
					if len(lianArr) == 0 {
						lianArr = append(lianArr, tmpPaixing.Jishu.ValueArr[i], tmpPaixing.Jishu.ValueArr[i+1])
					} else {
						lianArr = append(lianArr, tmpPaixing.Jishu.ValueArr[i+1])
					}
				} else {
					if len(lianArr) >= 5 { //大于五个组成连子咯
						break
					} else {
						lianArr = []int{}
					}
				}
			}

			if len(lianArr) >= 5 {
				tmpCards := getCardsFromValues(tcards, lianArr)
				retList = append(retList, tmpCards)
				tcards = deleEleFromArr(tcards, tmpCards)

			} else { //没有连字咯
				break
			}
		}

		//这个时候可能还有连队来
		for {
			tmpPaixing := GetPaixing(tcards)
			if tmpPaixing.Paixing != PAIXING_NON { //剩下的就一手了
				retList = append(retList, tcards)
				retList = append(retList, zhadanArr...)
				return retList
			}

			//我们取出所有连队，由于前面取过三个头，所以这里只需考虑剩下的对子就好了
			twoArr := tmpPaixing.Jishu.TwoArr
			lianArr := []int{}
			for i := 0; i < len(twoArr)-1; i++ {
				if twoArr[i] == twoArr[i+1]-1 {
					if len(lianArr) == 0 {
						lianArr = append(lianArr, twoArr[i], twoArr[i+1])
					} else {
						lianArr = append(lianArr, twoArr[i+1])
					}
				} else {
					if len(lianArr) >= 3 {
						break
					} else {
						lianArr = []int{}
					}
				}

			}

			if len(lianArr) >= 3 {
				tmpCards := getCardsFromValues(tcards, copyNumEle(lianArr, 2))
				retList = append(retList, tmpCards)
				tcards = deleEleFromArr(tcards, tmpCards)
			} else { //没有连字咯
				break
			}

		}

		for { //这里我们把三个和三个带一，三个带二都搞完了
			tempArr := GetBigthan(tcards, []int{1, 1, 1})
			tempArr1 := GetBigthan(tcards, []int{1, 1, 1, 2})
			tempArr2 := GetBigthan(tcards, []int{1, 1, 1, 2, 2})

			tempJishu1 := GetJishuArrData(tempArr1)
			tempJishu2 := GetJishuArrData(tempArr2)

			if len(tempArr2) > 0 { //三个都有
				tmpPaixing := GetPaixing(tcards)
				if tmpPaixing.Paixing != PAIXING_NON { //剩下的就一手了
					retList = append(retList, tcards)
					//把炸弹补回来
					retList = append(retList, zhadanArr...)

					return retList
				}

				if tempJishu1.OneArr[0] >= tempJishu2.TwoArr[0] { //单张比对子大，一样大的话，就带对子好了
					if tempJishu2.TwoArr[0] >= 14 { //出三个不带
						retList = append(retList, tempArr)
						tcards = deleEleFromArr(tcards, tempArr)
					} else { //出三带2
						retList = append(retList, tempArr2)
						tcards = deleEleFromArr(tcards, tempArr2)
					}

				} else {
					if tempJishu1.OneArr[0] >= 16 {
						retList = append(retList, tempArr)
						tcards = deleEleFromArr(tcards, tempArr)
					} else { //出三带1
						retList = append(retList, tempArr1)
						tcards = deleEleFromArr(tcards, tempArr1)
					}

				}
			} else if len(tempArr1) > 0 {
				if tempJishu1.OneArr[0] >= 16 {
					retList = append(retList, tempArr)
					tcards = deleEleFromArr(tcards, tempArr)
				} else { //出三带1
					retList = append(retList, tempArr1)
					tcards = deleEleFromArr(tcards, tempArr1)
				}
			} else if len(tempArr) > 0 {
				retList = append(retList, tempArr)
				tcards = deleEleFromArr(tcards, tempArr)
			} else {
				break
			}
		}

	}

	//这个时候只剩对子和单张了
	//考虑把对子出完
	//只剩对子和单张的时候，谁小谁献出，哈哈哈，一些特殊情况暂时不考虑啦
	tmpPaixing := GetPaixing(tcards)
	if tmpPaixing.Paixing != PAIXING_NON { //剩下的就一手了
		retList = append(retList, tcards)
		retList = append(retList, zhadanArr...)
		return retList
	}

	arr := copyNumEle(tmpPaixing.Jishu.OneArr, 1)
	arr = append(arr, tmpPaixing.Jishu.TwoArr...)

	sort.Sort(IntSlice(arr))
	for i := 0; i < len(arr); i++ {
		if IndexOf(tmpPaixing.Jishu.OneArr, arr[i]) != -1 { //在一个里面
			tmpCards := getCardsFromValues(tcards, []int{arr[i]})
			retList = append(retList, tmpCards)
			tcards = deleEleFromArr(tcards, tmpCards)
		} else if IndexOf(tmpPaixing.Jishu.TwoArr, arr[i]) != -1 { //在两个里面
			tmpCards := getCardsFromValues(tcards, []int{arr[i], arr[i]})
			retList = append(retList, tmpCards)
			tcards = deleEleFromArr(tcards, tmpCards)
		}
	}

	retList = append(retList, zhadanArr...)

	return retList

}
