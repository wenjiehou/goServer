package paixingLogic

import (
	"fmt"
	"sort"
)

const (
	HandCardMaxLen = 20  //最多手牌
	MinCardsValue  = -25 //价值最小值
	MaxCardsValue  = 106 //价值最大值
)

//自己出的，随便出
func FreeDizhuOutput(self []int, preNum int, nexNum int, other []int,
	preStep [][]int, nexStep [][]int) []int {

	_, retList := getBestList(self)
	fmt.Println("retList::", retList)
	lowest := 100
	bigest := -100
	var lowestPaixing *PockPaixing
	var bigestPaixing *PockPaixing

	var putPaixing *PockPaixing

	paixingList := []*PockPaixing{}

	if len(retList) > 0 {
		if len(retList) == 1 { //就一手牌，直接出了
			return retList[0]
		}

		//这个是最优组合了，那我肯定要出里面价值最低的牌，后期可以认为同一种牌型比较多，而且其中的最小的价值又很低，出这个牌应该是一个不错的选择
		for i := 0; i < len(retList); i++ {
			tmpp := GetPaixing(retList[i])
			tmpg := get_GroupData(tmpp)
			if tmpg.Value < lowest {
				lowestPaixing = tmpp
				lowest = tmpg.Value
			}
			if tmpg.Value > bigest {
				bigestPaixing = tmpp
				bigest = tmpg.Value
			}
			paixingList = append(paixingList, tmpp)
		}

		if len(retList) == 2 { //两手牌
			otherBigCards := GetBigthan(other, bigestPaixing.Cards)
			if len(otherBigCards) == 0 { //有一个绝对大牌，就先跑了
				putPaixing = bigestPaixing
			} else {
				//这里其实很讲究滴，很有策略

			}
		} else if len(retList) == 3 { //
			//尽量出一手自己要的回来的
			if lowestPaixing.Paixing == PAIXING_SANER && bigestPaixing.Paixing == PAIXING_SANYI {
				//交换一下屁股出 3+1
				putPaixing = GetPaixing([]int{lowestPaixing.Cards[0], lowestPaixing.Cards[1], lowestPaixing.Cards[2], bigestPaixing.Cards[3]})
			} else { //如果有顺子先跑顺子，因为这玩意不跑估计没啥机会跑了
				var threeTempp *PockPaixing

				for _, v := range paixingList {
					if v.Paixing == PAIXING_SHUNZI || v.Paixing == PAIXING_LIANDUI {
						if threeTempp == nil {
							threeTempp = v
						} else {
							if get_GroupData(v).Value < get_GroupData(threeTempp).Value {
								threeTempp = v
							}
						}
					}
				}

				if threeTempp != nil {
					putPaixing = threeTempp
				}

			}
		}

		if putPaixing == nil { //前面没有取到就取小的
			putPaixing = lowestPaixing
		}

		return freeDeal(putPaixing, paixingList)
	}
	return []int{}
}

//自己出的，随便出 地主的上家
func FreeDizhuSOutput(self []int, preNum int, nexNum int, other []int,
	preStep [][]int, nexStep [][]int) []int {
	_, retList := getBestList(self)
	lowest := 100
	bigest := -100
	var lowestPaixing *PockPaixing
	var bigestPaixing *PockPaixing

	var putPaixing *PockPaixing

	paixingList := []*PockPaixing{}

	if len(retList) > 0 {
		if len(retList) == 1 { //就一手牌，直接出了
			return retList[0]
		}

		//这个是最优组合了，那我肯定要出里面价值最低的牌，后期可以认为同一种牌型比较多，而且其中的最小的价值又很低，出这个牌应该是一个不错的选择
		for i := 0; i < len(retList); i++ {
			tmpp := GetPaixing(retList[i])
			tmpg := get_GroupData(tmpp)
			if tmpg.Value < lowest {
				lowestPaixing = tmpp
				lowest = tmpg.Value
			}
			if tmpg.Value > bigest {
				bigestPaixing = tmpp
				bigest = tmpg.Value
			}

			paixingList = append(paixingList, tmpp)
		}

		if len(retList) == 2 {
			otherBigCards := GetBigthan(other, bigestPaixing.Cards)
			if len(otherBigCards) == 0 {
				putPaixing = bigestPaixing
			} else {
				if len(lowestPaixing.Cards) != nexNum {
					putPaixing = lowestPaixing
				} else {
					if nexNum != 1 && nexNum != 2 {
						putPaixing = lowestPaixing
					}
				}

			}
		}

		if putPaixing == nil {
			var dizhuNum = nexNum

			if dizhuNum == 1 {
				//地主剩一张，我是他的上家，尽量不要出一张，
				for _, v := range paixingList {
					if v.Paixing != PAIXING_DAN {
						putPaixing = v
						break
					}
				}
				if putPaixing == nil { //尽量出一张大的牌顶一下
					putPaixing = bigestPaixing //出大的
				}
			} else if dizhuNum == 2 { //地主还剩两张
				for _, v := range paixingList {
					if v.Paixing != PAIXING_DUIZI {
						if v.Paixing == PAIXING_DAN { //单张不能出的太小
							if v.Jishu.OneArr[0] >= 12 {
								putPaixing = v
								break
							}
						} else {
							putPaixing = v
							break
						}

					}
				}
				//考虑到不能出最小的
				if putPaixing == nil { //尽量出一张大的牌顶一下
					putPaixing = bigestPaixing //出大的
				}
			} else if dizhuNum == 3 {
				for _, v := range paixingList {
					if v.Paixing != PAIXING_SAN {
						if v.Paixing == PAIXING_DAN { //单张不能出的太小
							if v.Jishu.OneArr[0] >= 12 {
								putPaixing = v
								break
							}
						} else {
							putPaixing = v
							break
						}

					}
				}
				//考虑到不能出最小的
				if putPaixing == nil { //尽量出一张大的牌顶一下
					putPaixing = bigestPaixing //出大的
				}
			} else if dizhuNum == 4 {
				for _, v := range paixingList {
					if v.Paixing != PAIXING_SANYI {
						if v.Paixing == PAIXING_DAN { //单张不能出的太小
							if v.Jishu.OneArr[0] >= 12 {
								putPaixing = v
								break
							}
						} else {
							putPaixing = v
							break
						}

					}
				}
				//考虑到不能出最小的
				if putPaixing == nil { //尽量出一张大的牌顶一下
					putPaixing = bigestPaixing //出大的
				}
			} else if dizhuNum == 5 {
				for _, v := range paixingList {
					if v.Paixing != PAIXING_SANER {
						if v.Paixing == PAIXING_DAN { //单张不能出的太小
							if v.Jishu.OneArr[0] >= 12 {
								putPaixing = v
								break
							}
						} else {
							putPaixing = v
							break
						}

					}
				}
				//考虑到不能出最小的
				if putPaixing == nil { //尽量出一张大的牌顶一下
					putPaixing = bigestPaixing //出大的
				}
			}
		}

		if putPaixing == nil { //我要顶门的，这个再考虑
			putPaixing = lowestPaixing
			if putPaixing.Paixing == PAIXING_DAN {
				for _, v := range paixingList {
					if v.Paixing == PAIXING_DAN {
						if v.Jishu.OneArr[0] >= 12 {
							putPaixing = v
							break
						}
					} else {
						if get_GroupData(v).Value <= 12 {
							putPaixing = v
						}
						break
					}
				}

				if putPaixing == nil { //常规取
					tempout := GetBigthan(self, []int{12})
					if len(tempout) > 0 {
						return tempout
					}
				}

			}

		}

		return freeDeal(putPaixing, paixingList)
	}
	return []int{}
}

//自己出的，随便出 地主的下家
func FreeDizhuXOutput(self []int, preNum int, nexNum int, other []int,
	preStep [][]int, nexStep [][]int) []int {
	_, retList := getBestList(self)
	fmt.Println("retList::", retList)
	lowest := 100
	bigest := -100
	var lowestPaixing *PockPaixing
	var bigestPaixing *PockPaixing

	var putPaixing *PockPaixing

	paixingList := []*PockPaixing{}

	if len(retList) > 0 {
		if len(retList) == 1 { //就一手牌，直接出了
			return retList[0]
		}
		//这个是最优组合了，那我肯定要出里面价值最低的牌，后期可以认为同一种牌型比较多，而且其中的最小的价值又很低，出这个牌应该是一个不错的选择
		for i := 0; i < len(retList); i++ {
			tmpp := GetPaixing(retList[i])
			tmpg := get_GroupData(tmpp)
			if tmpg.Value < lowest {
				lowestPaixing = tmpp
				lowest = tmpg.Value
			}
			if tmpg.Value > bigest {
				bigestPaixing = tmpp
				bigest = tmpg.Value
			}

			paixingList = append(paixingList, tmpp)
		}

		if len(retList) == 2 {
			otherBigCards := GetBigthan(other, bigestPaixing.Cards)
			if len(otherBigCards) == 0 {
				putPaixing = bigestPaixing
			} else {

				var dizhuNum = preNum
				//如果下家还有一张，我们拆一张最小的给他走
				if dizhuNum == 1 {
					danNum := 0
					for _, v := range paixingList {
						if v.Paixing == PAIXING_DAN {
							danNum += 1
						}
					}
					if danNum <= 1 { //最后出这个单牌就好了
						for _, v := range paixingList {
							if v.Paixing != PAIXING_DAN {
								putPaixing = v
							}
						}
					}

					if putPaixing == nil {
						if nexNum == 1 { //下家就一张牌咯
							//如最小的这张牌小于等与6，可以考虑给下家走
							sort.Sort(IntSlice(self))
							if GetPockValue(self[0]) <= 6 {
								return []int{self[0]}
							}
						}
					}
				}

				if putPaixing == nil {

				}

			}
		}

		if putPaixing == nil { //前面没有取到就取小的
			putPaixing = lowestPaixing
		}

		return freeDeal(putPaixing, paixingList)
	}
	return []int{}
}

//地主跟牌  outIdx -1上家出的牌 1下家出的牌  outCards
func DizhuGen(self []int, preNum int, nexNum int, other []int, selfStep [][]int,
	preStep [][]int, nexStep [][]int, outIdx int, outCards []int) []int {
	var retArr = make([]int, 0)

	if checkDuiwangAndOne(self) == true { //大王加一手牌
		return []int{551, 552}
	}

	outp := GetPaixing(outCards)
	selfp := GetPaixing(self)
	outg := get_GroupData(outp)
	selfg := get_GroupData(selfp)
	if outp.Paixing == PAIXING_NON {
		return []int{}
	} else if outp.Paixing == PAIXING_DAN { //单张
		if selfp.Paixing == PAIXING_DAN {

			if selfg.MaxCard > outg.MaxCard {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		BestHandCardValue := get_HandCardValue(self, selfp)
		//我们认为不出牌的话会让对手一个轮次，即加一轮（权值减少7）便于后续的对比参考。
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false

		//同类型压制 地主只用考虑自己，不用考虑是不是其他人
		tempSelf := copyNumEle(self, 1)
		for i := 0; i < len(tempSelf); i++ {
			if GetPockValue(tempSelf[i]) > outg.MaxCard {
				//除去这张牌并且求取临时价值
				tempArr := copyNumEle(tempSelf, 1)
				tempArr = deleEleFromArr(tempArr, []int{tempSelf[i]})

				tempV := get_HandCardValue(tempArr, nil)
				//选取总权值-轮次*7值最高的策略  因为我们认为剩余的手牌需要n次控手的机会才能出完，若轮次牌型很大（如炸弹） 则其-7的价值也会为正
				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
					BestHandCardValue = tempV
					BestMaxCards = []int{tempSelf[i]}
					PutCards = true
				}

			}
		}
		if PutCards == true {
			return BestMaxCards

		}

		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_DUIZI {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_DUIZI {
			if selfg.MaxCard > outg.MaxCard {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		//同类型压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		tempSelf := copyNumEle(self, 1)
		twoArr := append(selfp.Jishu.TwoArr, selfp.Jishu.ThreeArr...)
		twoArr = append(twoArr, selfp.Jishu.FourArr...)

		for i := 0; i < len(twoArr); i++ {
			if twoArr[i] > outp.Jishu.TwoArr[0] {
				v := twoArr[i]
				tempCars := getCardsFromValues(self, []int{v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}
			}
		}
		if PutCards == true {
			return BestMaxCards
		}
		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_SAN {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_SAN {
			if selfg.MaxCard > outg.MaxCard {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		//同类型压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		tempSelf := copyNumEle(self, 1)
		threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)

		for i := 0; i < len(threeArr); i++ {
			if threeArr[i] > outp.Jishu.ThreeArr[0] {
				v := threeArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}
			}
		}
		if PutCards == true {
			return BestMaxCards
		}
		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_SANYI {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_SANYI {
			if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}
		//同类型牌压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		//顺带出去的牌
		//		tmp_1 := 0

		tempSelf := copyNumEle(self, 1)

		threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)
		oneArr := append(selfp.Jishu.OneArr, selfp.Jishu.TwoArr...)
		oneArr = append(oneArr, threeArr...)

		for i := 0; i < len(threeArr); i++ {
			if threeArr[i] > outg.MaxCard { //大于就好
				v := threeArr[i]
				for j := 0; j < len(oneArr); j++ {
					if oneArr[j] != v {
						tempValues := []int{v, v, v, oneArr[j]}
						tempCars := getCardsFromValues(self, tempValues)
						tempArr := copyNumEle(tempSelf, 1)
						tempArr = deleEleFromArr(tempArr, tempCars) //去掉

						tempV := get_HandCardValue(tempArr, nil)

						if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
							BestHandCardValue = tempV
							BestMaxCards = tempCars
							PutCards = true
						}
					}
				}

			}
		}
		if PutCards {
			return BestMaxCards
		}

		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_SANER {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_SANER {
			if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}
		//同类型牌压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		//顺带出去的牌
		//		tmp_1 := 0

		tempSelf := copyNumEle(self, 1)

		threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)
		twoArr := append(selfp.Jishu.TwoArr, threeArr...)

		for i := 0; i < len(threeArr); i++ {
			if threeArr[i] > outg.MaxCard { //大于就好
				v := threeArr[i]
				for j := 0; j < len(twoArr); j++ {
					if twoArr[j] != v {
						tempValues := []int{v, v, v, twoArr[j], twoArr[j]}
						tempCars := getCardsFromValues(self, tempValues)
						tempArr := copyNumEle(tempSelf, 1)
						tempArr = deleEleFromArr(tempArr, tempCars) //去掉

						tempV := get_HandCardValue(tempArr, nil)

						if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
							BestHandCardValue = tempV
							BestMaxCards = tempCars
							PutCards = true
						}
					}
				}

			}
		}
		if PutCards {
			return BestMaxCards
		}

		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_SHUNZI {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_SHUNZI {
			if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		//同类型牌压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		tempSelf := copyNumEle(self, 1)

		oneArr := append(selfp.Jishu.OneArr, selfp.Jishu.TwoArr...)
		oneArr = append(oneArr, selfp.Jishu.ThreeArr...)
		oneArr = append(oneArr, selfp.Jishu.FourArr...)

		sort.Sort(IntSlice(oneArr))

		//验证顺子的标志
		prov := 0
		//顺子起点
		start_i := 0
		//顺子终点
		end_i := 0
		//顺子长度
		length := outg.Count

		for i := outg.MaxCard - length + 2; i < 15; i++ {
			if IndexOf(oneArr, i) != -1 {
				prov++
			} else {
				prov = 0
			}
			if prov >= length {
				end_i = i
				start_i = i - length + 1

				tempValues := []int{}

				for j := start_i; j <= end_i; j++ {
					tempValues = append(tempValues, j)
				}

				tempCars := getCardsFromValues(self, tempValues)
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}

		}

		if PutCards {
			return BestMaxCards
		}
		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_LIANDUI {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_LIANDUI {
			if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		//同类型牌压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		tempSelf := copyNumEle(self, 1)

		twoArr := append(selfp.Jishu.TwoArr, selfp.Jishu.ThreeArr...)
		twoArr = append(twoArr, selfp.Jishu.FourArr...)

		sort.Sort(IntSlice(twoArr))

		//验证顺子的标志
		prov := 0
		//顺子起点
		start_i := 0
		//顺子终点
		end_i := 0
		//顺子长度
		length := outg.Count / 2

		for i := outg.MaxCard - length + 2; i < 15; i++ {
			if IndexOf(twoArr, i) != -1 {
				prov++
			} else {
				prov = 0
			}
			if prov >= length {
				end_i = i
				start_i = i - length + 1

				tempValues := []int{}

				for j := start_i; j <= end_i; j++ {
					tempValues = append(tempValues, j, j)
				}

				tempCars := getCardsFromValues(self, tempValues)
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}

		}

		if PutCards {
			return BestMaxCards
		}
		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_FEIJI && outp.Feiji != nil {
		if outp.Feiji.Type == FEIJI_NON {
			//看看是不是正好大于并且一手出掉了
			if selfp.Paixing == PAIXING_FEIJI && selfp.Feiji != nil && selfp.Feiji.Type == FEIJI_NON {
				if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
					return self
				}
			} else if selfp.Paixing == PAIXING_ZHADAN {
				return self
			} else if selfp.Paixing == PAIXING_DUIWANG {
				return self
			}

			//同类型牌压制
			BestHandCardValue := get_HandCardValue(self, selfp)
			BestHandCardValue.NeedRound += 1
			//暂存最佳的牌号
			var BestMaxCards []int
			//是否出牌的标志
			PutCards := false
			tempSelf := copyNumEle(self, 1)

			threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)

			sort.Sort(IntSlice(threeArr))

			//验证顺子的标志
			prov := 0
			//顺子起点
			start_i := 0
			//顺子终点
			end_i := 0
			//顺子长度
			length := outg.Count / 3

			for i := outg.MaxCard - length + 2; i < 15; i++ {
				if IndexOf(threeArr, i) != -1 {
					prov++
				} else {
					prov = 0
				}
				if prov >= length {
					end_i = i
					start_i = i - length + 1

					tempValues := []int{}

					for j := start_i; j <= end_i; j++ {
						tempValues = append(tempValues, j, j, j)
					}

					tempCars := getCardsFromValues(self, tempValues)
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}

			}

			if PutCards {
				return BestMaxCards
			}
			//考虑出炸弹
			if len(selfp.Jishu.FourArr) > 0 {
				for i := 0; i < len(selfp.Jishu.FourArr); i++ {
					v := selfp.Jishu.FourArr[i]
					tempCars := getCardsFromValues(self, []int{v, v, v, v})
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}
				if PutCards == true {
					return BestMaxCards
				}
			}
			//考虑王炸
			if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
				if BestHandCardValue.SumValue > 20 {
					return []int{551, 552}
				}
			}
			//不出
			return []int{}

		} else if outp.Feiji.Type == FEIJI_DAN {
			//看看是不是正好大于并且一手出掉了
			if selfp.Paixing == PAIXING_FEIJI && selfp.Feiji != nil && selfp.Feiji.Type == FEIJI_DAN {
				if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
					return self
				}
			} else if selfp.Paixing == PAIXING_ZHADAN {
				return self
			} else if selfp.Paixing == PAIXING_DUIWANG {
				return self
			}

			//同类型牌压制
			// 2 3 4就好了
			//同类型牌压制
			BestHandCardValue := get_HandCardValue(self, selfp)
			BestHandCardValue.NeedRound += 1
			//暂存最佳的牌号
			var BestMaxCards []int
			//是否出牌的标志
			PutCards := false
			//验证顺子的标志
			prov := 0
			//顺子起点
			start_i := 0
			//顺子终点
			end_i := 0
			//顺子长度
			length := outp.Feiji.Length

			threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)
			sort.Sort(IntSlice(threeArr))
			tempSelf := copyNumEle(self, 1)

			for i := outg.MaxCard - length + 2; i < 15; i++ {
				if IndexOf(threeArr, i) != -1 {
					prov++
				} else {
					prov = 0
				}

				if prov >= length {
					end_i = i
					start_i = i - length + 1

					tempValues := []int{}

					for j := start_i; j < end_i; j++ {
						tempValues = append(tempValues, j, j, j)
					}

					tempCars := getCardsFromValues(self, tempValues)
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					if length == 2 {
						for j := 0; j < len(tempArr)-1; j++ {
							for k := j + 1; k < len(tempArr); k++ {
								arr := copyNumEle(tempArr, 1)                             //复制一份
								linshiCards := copyNumEle(tempCars, 1)                    //也复制一份
								linshiCards = append(linshiCards, tempArr[j], tempArr[k]) //带上腿

								arr = deleEleFromArr(arr, []int{tempArr[j], tempArr[k]})

								tempV := get_HandCardValue(arr, nil)

								if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
									BestHandCardValue = tempV
									BestMaxCards = linshiCards
									PutCards = true
								}

							}

						}

					} else if length == 3 {
						for j := 0; j < len(tempArr)-2; j++ {
							for k := j + 1; k < len(tempArr); k++ {
								for l := k + 1; l < len(tempArr); l++ {
									arr := copyNumEle(tempArr, 1)                                         //复制一份
									linshiCards := copyNumEle(tempCars, 1)                                //也复制一份
									linshiCards = append(linshiCards, tempArr[j], tempArr[k], tempArr[l]) //带上腿

									arr = deleEleFromArr(arr, []int{tempArr[j], tempArr[k], tempArr[l]})

									tempV := get_HandCardValue(arr, nil)

									if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
										BestHandCardValue = tempV
										BestMaxCards = linshiCards
										PutCards = true
									}
								}
							}

						}
					} else if length == 4 {
						for j := 0; j < len(tempArr)-3; j++ {
							for k := j + 1; k < len(tempArr); k++ {
								for l := k + 1; l < len(tempArr); l++ {
									for m := l + 1; m < len(tempArr); m++ {
										arr := copyNumEle(tempArr, 1)                                                     //复制一份
										linshiCards := copyNumEle(tempCars, 1)                                            //也复制一份
										linshiCards = append(linshiCards, tempArr[j], tempArr[k], tempArr[l], tempArr[m]) //带上腿

										arr = deleEleFromArr(arr, []int{tempArr[j], tempArr[k], tempArr[l], tempArr[m]})

										tempV := get_HandCardValue(arr, nil)

										if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
											BestHandCardValue = tempV
											BestMaxCards = linshiCards
											PutCards = true
										}
									}

								}
							}

						}
					}

				}
			}
			if PutCards {
				return BestMaxCards
			}

			//考虑出炸弹
			if len(selfp.Jishu.FourArr) > 0 {
				for i := 0; i < len(selfp.Jishu.FourArr); i++ {
					v := selfp.Jishu.FourArr[i]
					tempCars := getCardsFromValues(self, []int{v, v, v, v})
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}
				if PutCards == true {
					return BestMaxCards
				}
			}
			//考虑王炸
			if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
				if BestHandCardValue.SumValue > 20 {
					return []int{551, 552}
				}
			}
			//不出
			return []int{}

		} else if outp.Feiji.Type == FEIJI_DUI {
			//看看是不是正好大于并且一手出掉了
			if selfp.Paixing == PAIXING_FEIJI && selfp.Feiji != nil && selfp.Feiji.Type == FEIJI_DUI {
				if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
					return self
				}
			} else if selfp.Paixing == PAIXING_ZHADAN {
				return self
			} else if selfp.Paixing == PAIXING_DUIWANG {
				return self
			}

			//同类型牌压制
			// 2 3 4就好了
			//同类型牌压制
			BestHandCardValue := get_HandCardValue(self, selfp)
			BestHandCardValue.NeedRound += 1
			//暂存最佳的牌号
			var BestMaxCards []int
			//是否出牌的标志
			PutCards := false
			//验证顺子的标志
			prov := 0
			//顺子起点
			start_i := 0
			//顺子终点
			end_i := 0
			//顺子长度
			length := outp.Feiji.Length

			threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)
			sort.Sort(IntSlice(threeArr))
			tempSelf := copyNumEle(self, 1)

			for i := outg.MaxCard - length + 2; i < 15; i++ {
				if IndexOf(threeArr, i) != -1 {
					prov++
				} else {
					prov = 0
				}

				if prov >= length {
					end_i = i
					start_i = i - length + 1

					tempValues := []int{}

					for j := start_i; j < end_i; j++ {
						tempValues = append(tempValues, j, j, j)
					}

					tempCars := getCardsFromValues(self, tempValues)
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					sort.Sort(IntSlice(tempArr))
					tempJishu := GetJishuArrData(tempArr)
					twoArr := append(tempJishu.TwoArr, tempJishu.ThreeArr...)
					twoArr = append(twoArr, tempJishu.FourArr...)

					if length == 2 {
						for j := 0; j < len(twoArr)-1; j++ {
							for k := j + 1; k < len(twoArr); k++ {

								arr := copyNumEle(tempArr, 1)                                       //复制一份
								values := copyNumEle(tempValues, 1)                                 //也复制一份
								values = append(values, twoArr[j], twoArr[j], twoArr[k], twoArr[k]) //带上腿
								tempCars = getCardsFromValues(self, values)
								linshiCards := getCardsFromValues(tempArr, []int{twoArr[j], twoArr[j], twoArr[k], twoArr[k]})
								arr = deleEleFromArr(arr, linshiCards)
								tempV := get_HandCardValue(arr, nil)

								if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
									BestHandCardValue = tempV
									BestMaxCards = tempCars
									PutCards = true
								}

							}

						}

					} else if length == 3 {
						for j := 0; j < len(twoArr)-2; j++ {
							for k := j + 1; k < len(twoArr); k++ {
								for l := k + 1; l < len(twoArr); l++ {
									arr := copyNumEle(tempArr, 1)                                                             //复制一份
									values := copyNumEle(tempValues, 1)                                                       //也复制一份
									values = append(values, twoArr[j], twoArr[j], twoArr[k], twoArr[k], twoArr[l], twoArr[l]) //带上腿
									tempCars = getCardsFromValues(self, values)
									linshiCards := getCardsFromValues(tempArr, []int{twoArr[j], twoArr[j], twoArr[k], twoArr[k], twoArr[l], twoArr[l]})
									arr = deleEleFromArr(arr, linshiCards)

									tempV := get_HandCardValue(arr, nil)

									if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
										BestHandCardValue = tempV
										BestMaxCards = tempCars
										PutCards = true
									}
								}
							}

						}
					}

				}
			}
			if PutCards {
				return BestMaxCards
			}

			//考虑出炸弹
			if len(selfp.Jishu.FourArr) > 0 {
				for i := 0; i < len(selfp.Jishu.FourArr); i++ {
					v := selfp.Jishu.FourArr[i]
					tempCars := getCardsFromValues(self, []int{v, v, v, v})
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}
				if PutCards == true {
					return BestMaxCards
				}
			}
			//考虑王炸
			if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
				if BestHandCardValue.SumValue > 20 {
					return []int{551, 552}
				}
			}
			//不出
			return []int{}

		}
	} else if outp.Paixing == PAIXING_SIDAIER && outp.Sier != nil {

		if outp.Sier.Type == SIDAIER_DAN {
			//看看是不是正好大于并且一手出掉了
			if selfp.Paixing == PAIXING_SIDAIER && selfp.Sier != nil {
				if selfg.MaxCard > outg.MaxCard && selfp.Sier.Type == SIDAIER_DAN {
					return self
				}
			} else if selfp.Paixing == PAIXING_ZHADAN {
				return self
			} else if selfp.Paixing == PAIXING_DUIWANG {
				return self
			}

			//同类型牌压制
			BestHandCardValue := get_HandCardValue(self, selfp)
			BestHandCardValue.NeedRound += 1
			//暂存最佳的牌号
			var BestMaxCards []int
			//是否出牌的标志
			PutCards := false
			//顺带出去的牌
			//			tmp_1 := 0

			tempSelf := copyNumEle(self, 1)

			fourArr := copyNumEle(selfp.Jishu.FourArr, 1)
			oneArr := copyNumEle(selfp.Jishu.ValueArr, 1)

			if BestHandCardValue.SumValue <= 14 {

				for i := 0; i < len(fourArr); i++ {
					if fourArr[i] > outg.MaxCard { //大于就好
						v := fourArr[i]
						for j := 0; j < len(oneArr)-1; j++ {
							for k := j + 1; k < len(oneArr); k++ {
								if oneArr[j] != v && oneArr[k] != v {
									tempValues := []int{v, v, v, v, oneArr[j], oneArr[k]}
									tempCars := getCardsFromValues(self, tempValues)
									tempArr := copyNumEle(tempSelf, 1)
									tempArr = deleEleFromArr(tempArr, tempCars) //去掉

									tempV := get_HandCardValue(tempArr, nil)

									if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
										BestHandCardValue = tempV
										BestMaxCards = tempCars
										PutCards = true
									}
								}

							}
						}

					}
				}
				if PutCards {
					return BestMaxCards
				}
			}

			//考虑出炸弹
			if len(selfp.Jishu.FourArr) > 0 {
				for i := 0; i < len(selfp.Jishu.FourArr); i++ {
					v := selfp.Jishu.FourArr[i]
					tempCars := getCardsFromValues(self, []int{v, v, v, v})
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}
				if PutCards == true {
					return BestMaxCards
				}
			}
			//考虑王炸
			if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
				if BestHandCardValue.SumValue > 20 {
					return []int{551, 552}
				}
			}
			//不出
			return []int{}

		} else if outp.Sier.Type == SIDAIER_DUI {
			//看看是不是正好大于并且一手出掉了
			if selfp.Paixing == PAIXING_SIDAIER && selfp.Sier != nil {
				if selfg.MaxCard > outg.MaxCard && selfp.Sier.Type == SIDAIER_DUI {
					return self
				}
			} else if selfp.Paixing == PAIXING_ZHADAN {
				return self
			} else if selfp.Paixing == PAIXING_DUIWANG {
				return self
			}
			//同类型牌压制
			BestHandCardValue := get_HandCardValue(self, selfp)
			BestHandCardValue.NeedRound += 1
			//暂存最佳的牌号
			var BestMaxCards []int
			//是否出牌的标志
			PutCards := false
			//顺带出去的牌
			//			tmp_1 := 0

			tempSelf := copyNumEle(self, 1)

			fourArr := copyNumEle(selfp.Jishu.FourArr, 1)
			twoArr := append(selfp.Jishu.TwoArr, selfp.Jishu.ThreeArr...)

			if BestHandCardValue.SumValue <= 14 {
				for i := 0; i < len(fourArr); i++ {
					if fourArr[i] > outg.MaxCard { //大于就好
						v := fourArr[i]
						for j := 0; j < len(twoArr)-1; j++ {
							for k := j + 1; k < len(twoArr); k++ {
								if twoArr[j] != v && twoArr[k] != v {
									tempValues := []int{v, v, v, v, twoArr[j], twoArr[j], twoArr[k], twoArr[k]}
									tempCars := getCardsFromValues(self, tempValues)
									tempArr := copyNumEle(tempSelf, 1)
									tempArr = deleEleFromArr(tempArr, tempCars) //去掉

									tempV := get_HandCardValue(tempArr, nil)

									if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
										BestHandCardValue = tempV
										BestMaxCards = tempCars
										PutCards = true
									}
								}

							}
						}

					}
				}
				if PutCards {
					return BestMaxCards
				}
			}

			//考虑出炸弹
			if len(selfp.Jishu.FourArr) > 0 {
				for i := 0; i < len(selfp.Jishu.FourArr); i++ {
					v := selfp.Jishu.FourArr[i]
					tempCars := getCardsFromValues(self, []int{v, v, v, v})
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}
				if PutCards == true {
					return BestMaxCards
				}
			}
			//考虑王炸
			if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
				if BestHandCardValue.SumValue > 20 {
					return []int{551, 552}
				}
			}
			//不出
			return []int{}

		}
	} else if outp.Paixing == PAIXING_ZHADAN {
		return GetBigthan(self, outCards) //我是地主我怕谁

	} else if outp.Paixing == PAIXING_DUIWANG {
		return []int{}
	}

	//检查玩家是不是正好可以一手出掉且大于当前出的牌

	return retArr
}

//地主上家跟牌  outIdx -1上家出的牌 1下家出的牌  上家是盟友(在地主牌很少的时候就开始作弊了，因为其实这个时候正常基本可以判断出地主的牌，哈哈哈哈)
func DizhuSGen(self []int, preCards []int, nexCards []int, other []int, selfStep [][]int,
	preStep [][]int, nexStep [][]int, outIdx int, outCards []int) []int {
	var retArr = make([]int, 0)
	if checkDuiwangAndOne(self) == true { //大王加一手牌
		return []int{551, 552}
	}

	//	preNum := len(preCards)
	nexNum := len(nexCards) //地主的牌，哈哈哈，这个玩意其实有点作弊了，但是我觉得作弊不明显就好了

	outp := GetPaixing(outCards)
	selfp := GetPaixing(self)
	outg := get_GroupData(outp)
	selfg := get_GroupData(selfp)

	//如果地主大不了这个牌，我就不要
	if outIdx == -1 {
		dizhuBiger := GetBigthan(nexCards, outCards) //地主要不起的牌，我肯定不要
		if len(dizhuBiger) == 0 {
			return []int{}
		}
	}

	if outp.Paixing == PAIXING_NON {
		return []int{}
	} else if outp.Paixing == PAIXING_DAN { //单张
		if selfp.Paixing == PAIXING_DAN {
			if selfg.MaxCard > outg.MaxCard {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		//后面遍历的时候，遍历的牌都要大过这张
		intentCard := GetPockValue(self[0]) - 1
		mustBig := false

		if outIdx == -1 { //盟友出的牌
			if nexNum == 1 {
				//地主就剩一张了，考虑一下，
				if GetPockValue(outCards[0]) >= GetPockValue(nexCards[0]) { //这个肯定不要了
					return []int{}
				}
				//现在肯定要要滴
				intentCard = GetPockValue(nexCards[0])
				mustBig = true
			} else { //地主家不是一张
				if outg.MaxCard >= 13 { //除了可以跑路的，不用王压盟友
					return []int{}
				} else {
					intentCard = 12
				}
			}

		}

		//单张牌这里是需要顶门的，不然盟友会骂我

		BestHandCardValue := get_HandCardValue(self, selfp)
		//我们认为不出牌的话会让对手一个轮次，即加一轮（权值减少7）便于后续的对比参考。
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false

		//同类型压制 地主只用考虑自己，不用考虑是不是其他人
		tempSelf := copyNumEle(self, 1)
		for i := 0; i < len(tempSelf); i++ {
			must := true
			if mustBig == false {
				if GetPockValue(tempSelf[i]) <= 14 {
					must = true
				} else {
					must = false
				}
			}

			if must && GetPockValue(tempSelf[i]) > outg.MaxCard && GetPockValue(tempSelf[i]) >= intentCard {
				//除去这张牌并且求取临时价值

				tempArr := copyNumEle(tempSelf, 1)
				tempArr = deleEleFromArr(tempArr, []int{tempSelf[i]})

				tempV := get_HandCardValue(tempArr, nil)
				//选取总权值-轮次*7值最高的策略  因为我们认为剩余的手牌需要n次控手的机会才能出完，若轮次牌型很大（如炸弹） 则其-7的价值也会为正
				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
					BestHandCardValue = tempV
					BestMaxCards = []int{tempSelf[i]}
					PutCards = true
				}

			}
		}
		if PutCards == true {
			return BestMaxCards
		}

		//考虑拆单张来顶住
		if mustBig == false { //不是一定要大过，这个时候，有限考虑14以下的
			tempSelf := copyNumEle(self, 1)
			for i := 0; i < len(tempSelf); i++ {
				if GetPockValue(tempSelf[i]) > outg.MaxCard {
					//除去这张牌并且求取临时价值
					tempArr := copyNumEle(tempSelf, 1)
					tempArr = deleEleFromArr(tempArr, []int{tempSelf[i]})

					tempV := get_HandCardValue(tempArr, nil)
					//选取总权值-轮次*7值最高的策略  因为我们认为剩余的手牌需要n次控手的机会才能出完，若轮次牌型很大（如炸弹） 则其-7的价值也会为正
					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
						BestHandCardValue = tempV
						BestMaxCards = []int{tempSelf[i]}
						PutCards = true
					}

				}
			}
			if PutCards == true {
				return BestMaxCards
			}
		}

		if outIdx == -1 {
			if mustBig == false {
				return []int{} //盟友的牌不用炸弹和对王
			}
		}

		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_DUIZI {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_DUIZI {
			if selfg.MaxCard > outg.MaxCard {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		if outIdx == -1 {
			if outg.MaxCard >= 11 {
				return []int{}
			}
		}

		//同类型压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		tempSelf := copyNumEle(self, 1)
		twoArr := append(selfp.Jishu.TwoArr, selfp.Jishu.ThreeArr...)
		twoArr = append(twoArr, selfp.Jishu.FourArr...)

		for i := 0; i < len(twoArr); i++ {
			if twoArr[i] > outp.Jishu.TwoArr[0] {
				v := twoArr[i]
				tempCars := getCardsFromValues(self, []int{v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}
			}
		}
		if PutCards == true {
			return BestMaxCards
		}

		if outIdx == -1 {
			return []int{}
		}

		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_SAN {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_SAN {
			if selfg.MaxCard > outg.MaxCard {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		if outIdx == -1 {
			if outg.MaxCard >= 11 { //盟友3个j以上就不要
				return []int{}
			}
		}

		//同类型压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		tempSelf := copyNumEle(self, 1)
		threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)

		for i := 0; i < len(threeArr); i++ {
			if threeArr[i] > outp.Jishu.ThreeArr[0] {
				v := threeArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}
			}
		}
		if PutCards == true {
			return BestMaxCards
		}

		if outIdx == -1 {
			return []int{}
		}

		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_SANYI {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_SANYI {
			if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		if outIdx == -1 {
			if outg.MaxCard >= 11 { //盟友3个k以上就不要 毕竟可以带一张哈
				return []int{}
			}
		}

		//同类型牌压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		//顺带出去的牌
		//		tmp_1 := 0

		tempSelf := copyNumEle(self, 1)

		threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)
		oneArr := append(selfp.Jishu.OneArr, selfp.Jishu.TwoArr...)
		oneArr = append(oneArr, threeArr...)

		for i := 0; i < len(threeArr); i++ {
			if threeArr[i] > outg.MaxCard { //大于就好
				v := threeArr[i]
				for j := 0; j < len(oneArr); j++ {
					if oneArr[j] != v {
						tempValues := []int{v, v, v, oneArr[j]}
						tempCars := getCardsFromValues(self, tempValues)
						tempArr := copyNumEle(tempSelf, 1)
						tempArr = deleEleFromArr(tempArr, tempCars) //去掉

						tempV := get_HandCardValue(tempArr, nil)

						if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
							BestHandCardValue = tempV
							BestMaxCards = tempCars
							PutCards = true
						}
					}
				}

			}
		}
		if PutCards {
			return BestMaxCards
		}

		if outIdx == -1 {
			return []int{}
		}

		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_SANER {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_SANER {
			if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		if outIdx == -1 {
			if outg.MaxCard >= 11 {
				return []int{}
			}
		}
		//同类型牌压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		//顺带出去的牌
		//		tmp_1 := 0

		tempSelf := copyNumEle(self, 1)

		threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)
		twoArr := append(selfp.Jishu.TwoArr, threeArr...)

		for i := 0; i < len(threeArr); i++ {
			if threeArr[i] > outg.MaxCard { //大于就好
				v := threeArr[i]
				for j := 0; j < len(twoArr); j++ {
					if twoArr[j] != v {
						tempValues := []int{v, v, v, twoArr[j], twoArr[j]}
						tempCars := getCardsFromValues(self, tempValues)
						tempArr := copyNumEle(tempSelf, 1)
						tempArr = deleEleFromArr(tempArr, tempCars) //去掉

						tempV := get_HandCardValue(tempArr, nil)

						if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
							BestHandCardValue = tempV
							BestMaxCards = tempCars
							PutCards = true
						}
					}
				}

			}
		}
		if PutCards {
			return BestMaxCards
		}

		if outIdx == -1 {
			return []int{}
		}

		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_SHUNZI {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_SHUNZI {
			if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		if outIdx == -1 {
			return []int{}
		}

		//同类型牌压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		tempSelf := copyNumEle(self, 1)

		oneArr := append(selfp.Jishu.OneArr, selfp.Jishu.TwoArr...)
		oneArr = append(oneArr, selfp.Jishu.ThreeArr...)
		oneArr = append(oneArr, selfp.Jishu.FourArr...)

		sort.Sort(IntSlice(oneArr))

		//验证顺子的标志
		prov := 0
		//顺子起点
		start_i := 0
		//顺子终点
		end_i := 0
		//顺子长度
		length := outg.Count

		for i := outg.MaxCard - length + 2; i < 15; i++ {
			if IndexOf(oneArr, i) != -1 {
				prov++
			} else {
				prov = 0
			}
			if prov >= length {
				end_i = i
				start_i = i - length + 1

				tempValues := []int{}

				for j := start_i; j <= end_i; j++ {
					tempValues = append(tempValues, j)
				}

				tempCars := getCardsFromValues(self, tempValues)
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}

		}

		if PutCards {
			return BestMaxCards
		}

		if outIdx == -1 {
			return []int{}
		}

		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_LIANDUI {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_LIANDUI {
			if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		if outIdx == -1 {
			return []int{}

		}

		//同类型牌压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		tempSelf := copyNumEle(self, 1)

		twoArr := append(selfp.Jishu.TwoArr, selfp.Jishu.ThreeArr...)
		twoArr = append(twoArr, selfp.Jishu.FourArr...)

		sort.Sort(IntSlice(twoArr))

		//验证顺子的标志
		prov := 0
		//顺子起点
		start_i := 0
		//顺子终点
		end_i := 0
		//顺子长度
		length := outg.Count / 2

		for i := outg.MaxCard - length + 2; i < 15; i++ {
			if IndexOf(twoArr, i) != -1 {
				prov++
			} else {
				prov = 0
			}
			if prov >= length {
				end_i = i
				start_i = i - length + 1

				tempValues := []int{}

				for j := start_i; j <= end_i; j++ {
					tempValues = append(tempValues, j, j)
				}

				tempCars := getCardsFromValues(self, tempValues)
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}

		}

		if PutCards {
			return BestMaxCards
		}

		if outIdx == -1 {
			return []int{}
		}
		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_FEIJI && outp.Feiji != nil {
		if outp.Feiji.Type == FEIJI_NON {
			//看看是不是正好大于并且一手出掉了
			if selfp.Paixing == PAIXING_FEIJI && selfp.Feiji != nil && selfp.Feiji.Type == FEIJI_NON {
				if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
					return self
				}
			} else if selfp.Paixing == PAIXING_ZHADAN {
				return self
			} else if selfp.Paixing == PAIXING_DUIWANG {
				return self
			}

			if outIdx == -1 {
				return []int{}
			}

			//同类型牌压制
			BestHandCardValue := get_HandCardValue(self, selfp)
			BestHandCardValue.NeedRound += 1
			//暂存最佳的牌号
			var BestMaxCards []int
			//是否出牌的标志
			PutCards := false
			tempSelf := copyNumEle(self, 1)

			threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)

			sort.Sort(IntSlice(threeArr))

			//验证顺子的标志
			prov := 0
			//顺子起点
			start_i := 0
			//顺子终点
			end_i := 0
			//顺子长度
			length := outg.Count / 3

			for i := outg.MaxCard - length + 2; i < 15; i++ {
				if IndexOf(threeArr, i) != -1 {
					prov++
				} else {
					prov = 0
				}
				if prov >= length {
					end_i = i
					start_i = i - length + 1

					tempValues := []int{}

					for j := start_i; j <= end_i; j++ {
						tempValues = append(tempValues, j, j, j)
					}

					tempCars := getCardsFromValues(self, tempValues)
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}

			}

			if PutCards {
				return BestMaxCards
			}
			if outIdx == -1 {
				return []int{}
			}
			//考虑出炸弹
			if len(selfp.Jishu.FourArr) > 0 {
				for i := 0; i < len(selfp.Jishu.FourArr); i++ {
					v := selfp.Jishu.FourArr[i]
					tempCars := getCardsFromValues(self, []int{v, v, v, v})
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}
				if PutCards == true {
					return BestMaxCards
				}
			}
			//考虑王炸
			if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
				if BestHandCardValue.SumValue > 20 {
					return []int{551, 552}
				}
			}
			//不出
			return []int{}

		} else if outp.Feiji.Type == FEIJI_DAN {
			//看看是不是正好大于并且一手出掉了
			if selfp.Paixing == PAIXING_FEIJI && selfp.Feiji != nil && selfp.Feiji.Type == FEIJI_DAN {
				if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
					return self
				}
			} else if selfp.Paixing == PAIXING_ZHADAN {
				return self
			} else if selfp.Paixing == PAIXING_DUIWANG {
				return self
			}

			if outIdx == -1 {
				return []int{}
			}

			//同类型牌压制
			// 2 3 4就好了
			//同类型牌压制
			BestHandCardValue := get_HandCardValue(self, selfp)
			BestHandCardValue.NeedRound += 1
			//暂存最佳的牌号
			var BestMaxCards []int
			//是否出牌的标志
			PutCards := false
			//验证顺子的标志
			prov := 0
			//顺子起点
			start_i := 0
			//顺子终点
			end_i := 0
			//顺子长度
			length := outp.Feiji.Length

			threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)
			sort.Sort(IntSlice(threeArr))
			tempSelf := copyNumEle(self, 1)

			for i := outg.MaxCard - length + 2; i < 15; i++ {
				if IndexOf(threeArr, i) != -1 {
					prov++
				} else {
					prov = 0
				}

				if prov >= length {
					end_i = i
					start_i = i - length + 1

					tempValues := []int{}

					for j := start_i; j < end_i; j++ {
						tempValues = append(tempValues, j, j, j)
					}

					tempCars := getCardsFromValues(self, tempValues)
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					if length == 2 {
						for j := 0; j < len(tempArr)-1; j++ {
							for k := j + 1; k < len(tempArr); k++ {
								arr := copyNumEle(tempArr, 1)                             //复制一份
								linshiCards := copyNumEle(tempCars, 1)                    //也复制一份
								linshiCards = append(linshiCards, tempArr[j], tempArr[k]) //带上腿

								arr = deleEleFromArr(arr, []int{tempArr[j], tempArr[k]})

								tempV := get_HandCardValue(arr, nil)

								if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
									BestHandCardValue = tempV
									BestMaxCards = linshiCards
									PutCards = true
								}

							}

						}

					} else if length == 3 {
						for j := 0; j < len(tempArr)-2; j++ {
							for k := j + 1; k < len(tempArr); k++ {
								for l := k + 1; l < len(tempArr); l++ {
									arr := copyNumEle(tempArr, 1)                                         //复制一份
									linshiCards := copyNumEle(tempCars, 1)                                //也复制一份
									linshiCards = append(linshiCards, tempArr[j], tempArr[k], tempArr[l]) //带上腿

									arr = deleEleFromArr(arr, []int{tempArr[j], tempArr[k], tempArr[l]})

									tempV := get_HandCardValue(arr, nil)

									if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
										BestHandCardValue = tempV
										BestMaxCards = linshiCards
										PutCards = true
									}
								}
							}

						}
					} else if length == 4 {
						for j := 0; j < len(tempArr)-3; j++ {
							for k := j + 1; k < len(tempArr); k++ {
								for l := k + 1; l < len(tempArr); l++ {
									for m := l + 1; m < len(tempArr); m++ {
										arr := copyNumEle(tempArr, 1)                                                     //复制一份
										linshiCards := copyNumEle(tempCars, 1)                                            //也复制一份
										linshiCards = append(linshiCards, tempArr[j], tempArr[k], tempArr[l], tempArr[m]) //带上腿

										arr = deleEleFromArr(arr, []int{tempArr[j], tempArr[k], tempArr[l], tempArr[m]})

										tempV := get_HandCardValue(arr, nil)

										if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
											BestHandCardValue = tempV
											BestMaxCards = linshiCards
											PutCards = true
										}
									}

								}
							}

						}
					}

				}
			}
			if PutCards {
				return BestMaxCards
			}

			if outIdx == -1 {
				return []int{}
			}

			//考虑出炸弹
			if len(selfp.Jishu.FourArr) > 0 {
				for i := 0; i < len(selfp.Jishu.FourArr); i++ {
					v := selfp.Jishu.FourArr[i]
					tempCars := getCardsFromValues(self, []int{v, v, v, v})
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}
				if PutCards == true {
					return BestMaxCards
				}
			}
			//考虑王炸
			if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
				if BestHandCardValue.SumValue > 20 {
					return []int{551, 552}
				}
			}
			//不出
			return []int{}

		} else if outp.Feiji.Type == FEIJI_DUI {
			//看看是不是正好大于并且一手出掉了
			if selfp.Paixing == PAIXING_FEIJI && selfp.Feiji != nil && selfp.Feiji.Type == FEIJI_DUI {
				if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
					return self
				}
			} else if selfp.Paixing == PAIXING_ZHADAN {
				return self
			} else if selfp.Paixing == PAIXING_DUIWANG {
				return self
			}

			if outIdx == -1 {
				return []int{}
			}

			//同类型牌压制
			// 2 3 4就好了
			//同类型牌压制
			BestHandCardValue := get_HandCardValue(self, selfp)
			BestHandCardValue.NeedRound += 1
			//暂存最佳的牌号
			var BestMaxCards []int
			//是否出牌的标志
			PutCards := false
			//验证顺子的标志
			prov := 0
			//顺子起点
			start_i := 0
			//顺子终点
			end_i := 0
			//顺子长度
			length := outp.Feiji.Length

			threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)
			sort.Sort(IntSlice(threeArr))
			tempSelf := copyNumEle(self, 1)

			for i := outg.MaxCard - length + 2; i < 15; i++ {
				if IndexOf(threeArr, i) != -1 {
					prov++
				} else {
					prov = 0
				}

				if prov >= length {
					end_i = i
					start_i = i - length + 1

					tempValues := []int{}

					for j := start_i; j < end_i; j++ {
						tempValues = append(tempValues, j, j, j)
					}

					tempCars := getCardsFromValues(self, tempValues)
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					sort.Sort(IntSlice(tempArr))
					tempJishu := GetJishuArrData(tempArr)
					twoArr := append(tempJishu.TwoArr, tempJishu.ThreeArr...)
					twoArr = append(twoArr, tempJishu.FourArr...)

					if length == 2 {
						for j := 0; j < len(twoArr)-1; j++ {
							for k := j + 1; k < len(twoArr); k++ {

								arr := copyNumEle(tempArr, 1)                                       //复制一份
								values := copyNumEle(tempValues, 1)                                 //也复制一份
								values = append(values, twoArr[j], twoArr[j], twoArr[k], twoArr[k]) //带上腿
								tempCars = getCardsFromValues(self, values)
								linshiCards := getCardsFromValues(tempArr, []int{twoArr[j], twoArr[j], twoArr[k], twoArr[k]})
								arr = deleEleFromArr(arr, linshiCards)
								tempV := get_HandCardValue(arr, nil)

								if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
									BestHandCardValue = tempV
									BestMaxCards = tempCars
									PutCards = true
								}

							}

						}

					} else if length == 3 {
						for j := 0; j < len(twoArr)-2; j++ {
							for k := j + 1; k < len(twoArr); k++ {
								for l := k + 1; l < len(twoArr); l++ {
									arr := copyNumEle(tempArr, 1)                                                             //复制一份
									values := copyNumEle(tempValues, 1)                                                       //也复制一份
									values = append(values, twoArr[j], twoArr[j], twoArr[k], twoArr[k], twoArr[l], twoArr[l]) //带上腿
									tempCars = getCardsFromValues(self, values)
									linshiCards := getCardsFromValues(tempArr, []int{twoArr[j], twoArr[j], twoArr[k], twoArr[k], twoArr[l], twoArr[l]})
									arr = deleEleFromArr(arr, linshiCards)

									tempV := get_HandCardValue(arr, nil)

									if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
										BestHandCardValue = tempV
										BestMaxCards = tempCars
										PutCards = true
									}
								}
							}

						}
					}

				}
			}
			if PutCards {
				return BestMaxCards
			}

			if outIdx == -1 {
				return []int{}
			}

			//考虑出炸弹
			if len(selfp.Jishu.FourArr) > 0 {
				for i := 0; i < len(selfp.Jishu.FourArr); i++ {
					v := selfp.Jishu.FourArr[i]
					tempCars := getCardsFromValues(self, []int{v, v, v, v})
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}
				if PutCards == true {
					return BestMaxCards
				}
			}
			//考虑王炸
			if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
				if BestHandCardValue.SumValue > 20 {
					return []int{551, 552}
				}
			}
			//不出
			return []int{}

		}
	} else if outp.Paixing == PAIXING_SIDAIER && outp.Sier != nil {

		if outp.Sier.Type == SIDAIER_DAN {
			//看看是不是正好大于并且一手出掉了
			if selfp.Paixing == PAIXING_SIDAIER && selfp.Sier != nil {
				if selfg.MaxCard > outg.MaxCard && selfp.Sier.Type == SIDAIER_DAN {
					return self
				}
			} else if selfp.Paixing == PAIXING_ZHADAN {
				return self
			} else if selfp.Paixing == PAIXING_DUIWANG {
				return self
			}

			if outIdx == -1 { //盟友肯定想跑了
				return []int{}
			}

			//同类型牌压制
			BestHandCardValue := get_HandCardValue(self, selfp)
			BestHandCardValue.NeedRound += 1
			//暂存最佳的牌号
			var BestMaxCards []int
			//是否出牌的标志
			PutCards := false
			//顺带出去的牌
			//			tmp_1 := 0

			tempSelf := copyNumEle(self, 1)

			fourArr := copyNumEle(selfp.Jishu.FourArr, 1)
			oneArr := copyNumEle(selfp.Jishu.ValueArr, 1)

			if BestHandCardValue.SumValue <= 14 {

				for i := 0; i < len(fourArr); i++ {
					if fourArr[i] > outg.MaxCard { //大于就好
						v := fourArr[i]
						for j := 0; j < len(oneArr)-1; j++ {
							for k := j + 1; k < len(oneArr); k++ {
								if oneArr[j] != v && oneArr[k] != v {
									tempValues := []int{v, v, v, v, oneArr[j], oneArr[k]}
									tempCars := getCardsFromValues(self, tempValues)
									tempArr := copyNumEle(tempSelf, 1)
									tempArr = deleEleFromArr(tempArr, tempCars) //去掉

									tempV := get_HandCardValue(tempArr, nil)

									if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
										BestHandCardValue = tempV
										BestMaxCards = tempCars
										PutCards = true
									}
								}

							}
						}

					}
				}
				if PutCards {
					return BestMaxCards
				}
			}

			//考虑出炸弹
			if len(selfp.Jishu.FourArr) > 0 {
				for i := 0; i < len(selfp.Jishu.FourArr); i++ {
					v := selfp.Jishu.FourArr[i]
					tempCars := getCardsFromValues(self, []int{v, v, v, v})
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}
				if PutCards == true {
					return BestMaxCards
				}
			}
			//考虑王炸
			if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
				if BestHandCardValue.SumValue > 20 {
					return []int{551, 552}
				}
			}
			//不出
			return []int{}

		} else if outp.Sier.Type == SIDAIER_DUI {
			//看看是不是正好大于并且一手出掉了
			if selfp.Paixing == PAIXING_SIDAIER && selfp.Sier != nil {
				if selfg.MaxCard > outg.MaxCard && selfp.Sier.Type == SIDAIER_DUI {
					return self
				}
			} else if selfp.Paixing == PAIXING_ZHADAN {
				return self
			} else if selfp.Paixing == PAIXING_DUIWANG {
				return self
			}

			if outIdx == -1 {
				return []int{}
			}

			//同类型牌压制
			BestHandCardValue := get_HandCardValue(self, selfp)
			BestHandCardValue.NeedRound += 1
			//暂存最佳的牌号
			var BestMaxCards []int
			//是否出牌的标志
			PutCards := false
			//顺带出去的牌
			//			tmp_1 := 0

			tempSelf := copyNumEle(self, 1)

			fourArr := copyNumEle(selfp.Jishu.FourArr, 1)
			twoArr := append(selfp.Jishu.TwoArr, selfp.Jishu.ThreeArr...)

			if BestHandCardValue.SumValue <= 14 {
				for i := 0; i < len(fourArr); i++ {
					if fourArr[i] > outg.MaxCard { //大于就好
						v := fourArr[i]
						for j := 0; j < len(twoArr)-1; j++ {
							for k := j + 1; k < len(twoArr); k++ {
								if twoArr[j] != v && twoArr[k] != v {
									tempValues := []int{v, v, v, v, twoArr[j], twoArr[j], twoArr[k], twoArr[k]}
									tempCars := getCardsFromValues(self, tempValues)
									tempArr := copyNumEle(tempSelf, 1)
									tempArr = deleEleFromArr(tempArr, tempCars) //去掉

									tempV := get_HandCardValue(tempArr, nil)

									if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
										BestHandCardValue = tempV
										BestMaxCards = tempCars
										PutCards = true
									}
								}

							}
						}

					}
				}
				if PutCards {
					return BestMaxCards
				}
			}

			//考虑出炸弹
			if len(selfp.Jishu.FourArr) > 0 {
				for i := 0; i < len(selfp.Jishu.FourArr); i++ {
					v := selfp.Jishu.FourArr[i]
					tempCars := getCardsFromValues(self, []int{v, v, v, v})
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}
				if PutCards == true {
					return BestMaxCards
				}
			}
			//考虑王炸
			if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
				if BestHandCardValue.SumValue > 20 {
					return []int{551, 552}
				}
			}
			//不出
			return []int{}

		}
	} else if outp.Paixing == PAIXING_ZHADAN {
		if outIdx == -1 {
			return []int{}
		} else {
			return GetBigthan(self, outCards)
		}

	} else if outp.Paixing == PAIXING_DUIWANG {
		return []int{}
	}

	//检查玩家是不是正好可以一手出掉且大于当前出的牌

	return retArr
}

//地主下家跟牌  outIdx -1上家出的牌 1下家出的牌  outCards   outidx == 1 为盟友
func DizhuXGen(self []int, preCards []int, nexCards []int, other []int, selfStep [][]int,
	preStep [][]int, nexStep [][]int, outIdx int, outCards []int) []int {

	preNum := len(preCards)
	nexNum := len(nexCards)

	if outIdx == 1 { //盟友一张，地主不要，当然不能要咯
		if nexNum == 1 {
			return []int{}
		}
	}

	var retArr = make([]int, 0)

	if checkDuiwangAndOne(self) == true { //大王加一手牌
		return []int{551, 552}
	}

	outp := GetPaixing(outCards)
	selfp := GetPaixing(self)
	outg := get_GroupData(outp)
	selfg := get_GroupData(selfp)

	otherp := GetPaixing(other)

	if outp.Paixing == PAIXING_NON {
		return []int{}
	} else if outp.Paixing == PAIXING_DAN { //单张
		if selfp.Paixing == PAIXING_DAN {

			if selfg.MaxCard > outg.MaxCard {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		if outIdx == 1 { //盟友的牌，地主都不要了
			if preNum != 1 { //地主只剩一张了，考虑到地主的上家可能是顶着不让地主出
				//砸锅卖铁也要要
				if outg.MaxCard <= 10 { //10地主都要不到了，我也不接
					return []int{}
				} else if outg.MaxCard >= 16 { //2我也不要
					return []int{}
				}
			}
		}

		BestHandCardValue := get_HandCardValue(self, selfp)
		//我们认为不出牌的话会让对手一个轮次，即加一轮（权值减少7）便于后续的对比参考。
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false

		//同类型压制 地主只用考虑自己，不用考虑是不是其他人
		tempSelf := copyNumEle(self, 1)
		for i := 0; i < len(tempSelf); i++ {
			if GetPockValue(tempSelf[i]) > outg.MaxCard {
				//除去这张牌并且求取临时价值
				tempArr := copyNumEle(tempSelf, 1)
				tempArr = deleEleFromArr(tempArr, []int{tempSelf[i]})

				tempV := get_HandCardValue(tempArr, nil)
				//选取总权值-轮次*7值最高的策略  因为我们认为剩余的手牌需要n次控手的机会才能出完，若轮次牌型很大（如炸弹） 则其-7的价值也会为正
				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
					BestHandCardValue = tempV
					BestMaxCards = []int{tempSelf[i]}
					PutCards = true
				}

			}
		}
		if PutCards == true {
			return BestMaxCards

		}

		b, zhaArr := checkZhadanAndOne(self, selfp) //如果去除炸弹剩一手，妈的，直接炸了
		if b == true {
			if preNum <= 4 || len(otherp.Jishu.FourArr) <= 0 || otherp.Jishu.FourArr[len(otherp.Jishu.FourArr)-1] < GetPockValue(zhaArr[0]) {
				return zhaArr
			}

		}

		if outIdx == 1 {
			return []int{}
		}

		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_DUIZI {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_DUIZI {
			if selfg.MaxCard > outg.MaxCard {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		//尽量拿过来哈

		//同类型压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		tempSelf := copyNumEle(self, 1)
		twoArr := append(selfp.Jishu.TwoArr, selfp.Jishu.ThreeArr...)
		twoArr = append(twoArr, selfp.Jishu.FourArr...)

		for i := 0; i < len(twoArr); i++ {
			if twoArr[i] > outp.Jishu.TwoArr[0] {
				v := twoArr[i]
				tempCars := getCardsFromValues(self, []int{v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}
			}
		}
		if PutCards == true {
			return BestMaxCards
		}

		b, zhaArr := checkZhadanAndOne(self, selfp) //如果去除炸弹剩一手，妈的，直接炸了
		if b == true {
			if preNum <= 4 || len(otherp.Jishu.FourArr) <= 0 || otherp.Jishu.FourArr[len(otherp.Jishu.FourArr)-1] < GetPockValue(zhaArr[0]) {
				return zhaArr
			}

		}

		if outIdx == 1 {
			return []int{}
		}

		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_SAN {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_SAN {
			if selfg.MaxCard > outg.MaxCard {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		//同类型压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		tempSelf := copyNumEle(self, 1)
		threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)

		for i := 0; i < len(threeArr); i++ {
			if threeArr[i] > outp.Jishu.ThreeArr[0] {
				v := threeArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}
			}
		}
		if PutCards == true {
			return BestMaxCards
		}

		b, zhaArr := checkZhadanAndOne(self, selfp) //如果去除炸弹剩一手，妈的，直接炸了
		if b == true {
			if preNum <= 4 || len(otherp.Jishu.FourArr) <= 0 || otherp.Jishu.FourArr[len(otherp.Jishu.FourArr)-1] < GetPockValue(zhaArr[0]) {
				return zhaArr
			}

		}

		if outIdx == 1 {
			return []int{}
		}

		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_SANYI {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_SANYI {
			if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		//同类型牌压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		//顺带出去的牌
		//		tmp_1 := 0

		tempSelf := copyNumEle(self, 1)

		threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)
		oneArr := append(selfp.Jishu.OneArr, selfp.Jishu.TwoArr...)
		oneArr = append(oneArr, threeArr...)

		for i := 0; i < len(threeArr); i++ {
			if threeArr[i] > outg.MaxCard { //大于就好
				v := threeArr[i]
				for j := 0; j < len(oneArr); j++ {
					if oneArr[j] != v {
						tempValues := []int{v, v, v, oneArr[j]}
						tempCars := getCardsFromValues(self, tempValues)
						tempArr := copyNumEle(tempSelf, 1)
						tempArr = deleEleFromArr(tempArr, tempCars) //去掉

						tempV := get_HandCardValue(tempArr, nil)

						if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
							BestHandCardValue = tempV
							BestMaxCards = tempCars
							PutCards = true
						}
					}
				}

			}
		}
		if PutCards {
			return BestMaxCards
		}

		b, zhaArr := checkZhadanAndOne(self, selfp) //如果去除炸弹剩一手，妈的，直接炸了
		if b == true {
			if preNum <= 4 || len(otherp.Jishu.FourArr) <= 0 || otherp.Jishu.FourArr[len(otherp.Jishu.FourArr)-1] < GetPockValue(zhaArr[0]) {
				return zhaArr
			}

		}

		if outIdx == 1 {
			return []int{}
		}

		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_SANER {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_SANER {
			if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		//同类型牌压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		//顺带出去的牌
		//		tmp_1 := 0

		tempSelf := copyNumEle(self, 1)

		threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)
		twoArr := append(selfp.Jishu.TwoArr, threeArr...)

		for i := 0; i < len(threeArr); i++ {
			if threeArr[i] > outg.MaxCard { //大于就好
				v := threeArr[i]
				for j := 0; j < len(twoArr); j++ {
					if twoArr[j] != v {
						tempValues := []int{v, v, v, twoArr[j], twoArr[j]}
						tempCars := getCardsFromValues(self, tempValues)
						tempArr := copyNumEle(tempSelf, 1)
						tempArr = deleEleFromArr(tempArr, tempCars) //去掉

						tempV := get_HandCardValue(tempArr, nil)

						if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
							BestHandCardValue = tempV
							BestMaxCards = tempCars
							PutCards = true
						}
					}
				}

			}
		}
		if PutCards {
			return BestMaxCards
		}

		b, zhaArr := checkZhadanAndOne(self, selfp) //如果去除炸弹剩一手，妈的，直接炸了
		if b == true {
			if preNum <= 4 || len(otherp.Jishu.FourArr) <= 0 || otherp.Jishu.FourArr[len(otherp.Jishu.FourArr)-1] < GetPockValue(zhaArr[0]) {
				return zhaArr
			}

		}

		if outIdx == 1 {
			return []int{}
		}

		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_SHUNZI {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_SHUNZI {
			if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		//同类型牌压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		tempSelf := copyNumEle(self, 1)

		oneArr := append(selfp.Jishu.OneArr, selfp.Jishu.TwoArr...)
		oneArr = append(oneArr, selfp.Jishu.ThreeArr...)
		oneArr = append(oneArr, selfp.Jishu.FourArr...)

		sort.Sort(IntSlice(oneArr))

		//验证顺子的标志
		prov := 0
		//顺子起点
		start_i := 0
		//顺子终点
		end_i := 0
		//顺子长度
		length := outg.Count

		for i := outg.MaxCard - length + 2; i < 15; i++ {
			if IndexOf(oneArr, i) != -1 {
				prov++
			} else {
				prov = 0
			}
			if prov >= length {
				end_i = i
				start_i = i - length + 1

				tempValues := []int{}

				for j := start_i; j <= end_i; j++ {
					tempValues = append(tempValues, j)
				}

				tempCars := getCardsFromValues(self, tempValues)
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}

		}

		if PutCards {
			return BestMaxCards
		}
		b, zhaArr := checkZhadanAndOne(self, selfp) //如果去除炸弹剩一手，妈的，直接炸了
		if b == true {
			if preNum <= 4 || len(otherp.Jishu.FourArr) <= 0 || otherp.Jishu.FourArr[len(otherp.Jishu.FourArr)-1] < GetPockValue(zhaArr[0]) {
				return zhaArr
			}

		}

		if outIdx == 1 {
			return []int{}
		}

		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_LIANDUI {
		//看看是不是正好大于并且一手出掉了
		if selfp.Paixing == PAIXING_LIANDUI {
			if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
				return self
			}
		} else if selfp.Paixing == PAIXING_ZHADAN {
			return self
		} else if selfp.Paixing == PAIXING_DUIWANG {
			return self
		}

		//同类型牌压制
		BestHandCardValue := get_HandCardValue(self, selfp)
		BestHandCardValue.NeedRound += 1
		//暂存最佳的牌号
		var BestMaxCards []int
		//是否出牌的标志
		PutCards := false
		tempSelf := copyNumEle(self, 1)

		twoArr := append(selfp.Jishu.TwoArr, selfp.Jishu.ThreeArr...)
		twoArr = append(twoArr, selfp.Jishu.FourArr...)

		sort.Sort(IntSlice(twoArr))

		//验证顺子的标志
		prov := 0
		//顺子起点
		start_i := 0
		//顺子终点
		end_i := 0
		//顺子长度
		length := outg.Count / 2

		for i := outg.MaxCard - length + 2; i < 15; i++ {
			if IndexOf(twoArr, i) != -1 {
				prov++
			} else {
				prov = 0
			}
			if prov >= length {
				end_i = i
				start_i = i - length + 1

				tempValues := []int{}

				for j := start_i; j <= end_i; j++ {
					tempValues = append(tempValues, j, j)
				}

				tempCars := getCardsFromValues(self, tempValues)
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}

		}

		if PutCards {
			return BestMaxCards
		}

		b, zhaArr := checkZhadanAndOne(self, selfp) //如果去除炸弹剩一手，妈的，直接炸了
		if b == true {
			if preNum <= 4 || len(otherp.Jishu.FourArr) <= 0 || otherp.Jishu.FourArr[len(otherp.Jishu.FourArr)-1] < GetPockValue(zhaArr[0]) {
				return zhaArr
			}

		}

		if outIdx == 1 {
			return []int{}
		}
		//考虑出炸弹
		if len(selfp.Jishu.FourArr) > 0 {
			for i := 0; i < len(selfp.Jishu.FourArr); i++ {
				v := selfp.Jishu.FourArr[i]
				tempCars := getCardsFromValues(self, []int{v, v, v, v})
				tempArr := copyNumEle(tempSelf, 1)

				tempArr = deleEleFromArr(tempArr, tempCars)

				tempV := get_HandCardValue(tempArr, nil)

				if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
					BestHandCardValue = tempV
					BestMaxCards = tempCars
					PutCards = true
				}

			}
			if PutCards == true {
				return BestMaxCards
			}
		}
		//考虑王炸
		if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
			if BestHandCardValue.SumValue > 20 {
				return []int{551, 552}
			}
		}
		//不出
		return []int{}

	} else if outp.Paixing == PAIXING_FEIJI && outp.Feiji != nil {
		if outp.Feiji.Type == FEIJI_NON {
			//看看是不是正好大于并且一手出掉了
			if selfp.Paixing == PAIXING_FEIJI && selfp.Feiji != nil && selfp.Feiji.Type == FEIJI_NON {
				if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
					return self
				}
			} else if selfp.Paixing == PAIXING_ZHADAN {
				return self
			} else if selfp.Paixing == PAIXING_DUIWANG {
				return self
			}

			//同类型牌压制
			BestHandCardValue := get_HandCardValue(self, selfp)
			BestHandCardValue.NeedRound += 1
			//暂存最佳的牌号
			var BestMaxCards []int
			//是否出牌的标志
			PutCards := false
			tempSelf := copyNumEle(self, 1)

			threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)

			sort.Sort(IntSlice(threeArr))

			//验证顺子的标志
			prov := 0
			//顺子起点
			start_i := 0
			//顺子终点
			end_i := 0
			//顺子长度
			length := outg.Count / 3

			for i := outg.MaxCard - length + 2; i < 15; i++ {
				if IndexOf(threeArr, i) != -1 {
					prov++
				} else {
					prov = 0
				}
				if prov >= length {
					end_i = i
					start_i = i - length + 1

					tempValues := []int{}

					for j := start_i; j <= end_i; j++ {
						tempValues = append(tempValues, j, j, j)
					}

					tempCars := getCardsFromValues(self, tempValues)
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}

			}

			if PutCards {
				return BestMaxCards
			}

			b, zhaArr := checkZhadanAndOne(self, selfp) //如果去除炸弹剩一手，妈的，直接炸了
			if b == true {
				if preNum <= 4 || len(otherp.Jishu.FourArr) <= 0 || otherp.Jishu.FourArr[len(otherp.Jishu.FourArr)-1] < GetPockValue(zhaArr[0]) {
					return zhaArr
				}

			}

			if outIdx == 1 {
				return []int{}
			}

			//考虑出炸弹
			if len(selfp.Jishu.FourArr) > 0 {
				for i := 0; i < len(selfp.Jishu.FourArr); i++ {
					v := selfp.Jishu.FourArr[i]
					tempCars := getCardsFromValues(self, []int{v, v, v, v})
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}
				if PutCards == true {
					return BestMaxCards
				}
			}
			//考虑王炸
			if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
				if BestHandCardValue.SumValue > 20 {
					return []int{551, 552}
				}
			}
			//不出
			return []int{}

		} else if outp.Feiji.Type == FEIJI_DAN {
			//看看是不是正好大于并且一手出掉了
			if selfp.Paixing == PAIXING_FEIJI && selfp.Feiji != nil && selfp.Feiji.Type == FEIJI_DAN {
				if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
					return self
				}
			} else if selfp.Paixing == PAIXING_ZHADAN {
				return self
			} else if selfp.Paixing == PAIXING_DUIWANG {
				return self
			}

			//同类型牌压制
			// 2 3 4就好了
			//同类型牌压制
			BestHandCardValue := get_HandCardValue(self, selfp)
			BestHandCardValue.NeedRound += 1
			//暂存最佳的牌号
			var BestMaxCards []int
			//是否出牌的标志
			PutCards := false
			//验证顺子的标志
			prov := 0
			//顺子起点
			start_i := 0
			//顺子终点
			end_i := 0
			//顺子长度
			length := outp.Feiji.Length

			threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)
			sort.Sort(IntSlice(threeArr))
			tempSelf := copyNumEle(self, 1)

			for i := outg.MaxCard - length + 2; i < 15; i++ {
				if IndexOf(threeArr, i) != -1 {
					prov++
				} else {
					prov = 0
				}

				if prov >= length {
					end_i = i
					start_i = i - length + 1

					tempValues := []int{}

					for j := start_i; j < end_i; j++ {
						tempValues = append(tempValues, j, j, j)
					}

					tempCars := getCardsFromValues(self, tempValues)
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					if length == 2 {
						for j := 0; j < len(tempArr)-1; j++ {
							for k := j + 1; k < len(tempArr); k++ {
								arr := copyNumEle(tempArr, 1)                             //复制一份
								linshiCards := copyNumEle(tempCars, 1)                    //也复制一份
								linshiCards = append(linshiCards, tempArr[j], tempArr[k]) //带上腿

								arr = deleEleFromArr(arr, []int{tempArr[j], tempArr[k]})

								tempV := get_HandCardValue(arr, nil)

								if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
									BestHandCardValue = tempV
									BestMaxCards = linshiCards
									PutCards = true
								}

							}

						}

					} else if length == 3 {
						for j := 0; j < len(tempArr)-2; j++ {
							for k := j + 1; k < len(tempArr); k++ {
								for l := k + 1; l < len(tempArr); l++ {
									arr := copyNumEle(tempArr, 1)                                         //复制一份
									linshiCards := copyNumEle(tempCars, 1)                                //也复制一份
									linshiCards = append(linshiCards, tempArr[j], tempArr[k], tempArr[l]) //带上腿

									arr = deleEleFromArr(arr, []int{tempArr[j], tempArr[k], tempArr[l]})

									tempV := get_HandCardValue(arr, nil)

									if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
										BestHandCardValue = tempV
										BestMaxCards = linshiCards
										PutCards = true
									}
								}
							}

						}
					} else if length == 4 {
						for j := 0; j < len(tempArr)-3; j++ {
							for k := j + 1; k < len(tempArr); k++ {
								for l := k + 1; l < len(tempArr); l++ {
									for m := l + 1; m < len(tempArr); m++ {
										arr := copyNumEle(tempArr, 1)                                                     //复制一份
										linshiCards := copyNumEle(tempCars, 1)                                            //也复制一份
										linshiCards = append(linshiCards, tempArr[j], tempArr[k], tempArr[l], tempArr[m]) //带上腿

										arr = deleEleFromArr(arr, []int{tempArr[j], tempArr[k], tempArr[l], tempArr[m]})

										tempV := get_HandCardValue(arr, nil)

										if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
											BestHandCardValue = tempV
											BestMaxCards = linshiCards
											PutCards = true
										}
									}

								}
							}

						}
					}

				}
			}
			if PutCards {
				return BestMaxCards
			}

			b, zhaArr := checkZhadanAndOne(self, selfp) //如果去除炸弹剩一手，妈的，直接炸了
			if b == true {
				if preNum <= 4 || len(otherp.Jishu.FourArr) <= 0 || otherp.Jishu.FourArr[len(otherp.Jishu.FourArr)-1] < GetPockValue(zhaArr[0]) {
					return zhaArr
				}

			}

			if outIdx == 1 {
				return []int{}
			}

			//考虑出炸弹
			if len(selfp.Jishu.FourArr) > 0 {
				for i := 0; i < len(selfp.Jishu.FourArr); i++ {
					v := selfp.Jishu.FourArr[i]
					tempCars := getCardsFromValues(self, []int{v, v, v, v})
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}
				if PutCards == true {
					return BestMaxCards
				}
			}
			//考虑王炸
			if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
				if BestHandCardValue.SumValue > 20 {
					return []int{551, 552}
				}
			}
			//不出
			return []int{}

		} else if outp.Feiji.Type == FEIJI_DUI {
			//看看是不是正好大于并且一手出掉了
			if selfp.Paixing == PAIXING_FEIJI && selfp.Feiji != nil && selfp.Feiji.Type == FEIJI_DUI {
				if selfg.MaxCard > outg.MaxCard && selfg.Count == outg.Count {
					return self
				}
			} else if selfp.Paixing == PAIXING_ZHADAN {
				return self
			} else if selfp.Paixing == PAIXING_DUIWANG {
				return self
			}

			//同类型牌压制
			// 2 3 4就好了
			//同类型牌压制
			BestHandCardValue := get_HandCardValue(self, selfp)
			BestHandCardValue.NeedRound += 1
			//暂存最佳的牌号
			var BestMaxCards []int
			//是否出牌的标志
			PutCards := false
			//验证顺子的标志
			prov := 0
			//顺子起点
			start_i := 0
			//顺子终点
			end_i := 0
			//顺子长度
			length := outp.Feiji.Length

			threeArr := append(selfp.Jishu.ThreeArr, selfp.Jishu.FourArr...)
			sort.Sort(IntSlice(threeArr))
			tempSelf := copyNumEle(self, 1)

			for i := outg.MaxCard - length + 2; i < 15; i++ {
				if IndexOf(threeArr, i) != -1 {
					prov++
				} else {
					prov = 0
				}

				if prov >= length {
					end_i = i
					start_i = i - length + 1

					tempValues := []int{}

					for j := start_i; j < end_i; j++ {
						tempValues = append(tempValues, j, j, j)
					}

					tempCars := getCardsFromValues(self, tempValues)
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					sort.Sort(IntSlice(tempArr))
					tempJishu := GetJishuArrData(tempArr)
					twoArr := append(tempJishu.TwoArr, tempJishu.ThreeArr...)
					twoArr = append(twoArr, tempJishu.FourArr...)

					if length == 2 {
						for j := 0; j < len(twoArr)-1; j++ {
							for k := j + 1; k < len(twoArr); k++ {

								arr := copyNumEle(tempArr, 1)                                       //复制一份
								values := copyNumEle(tempValues, 1)                                 //也复制一份
								values = append(values, twoArr[j], twoArr[j], twoArr[k], twoArr[k]) //带上腿
								tempCars = getCardsFromValues(self, values)
								linshiCards := getCardsFromValues(tempArr, []int{twoArr[j], twoArr[j], twoArr[k], twoArr[k]})
								arr = deleEleFromArr(arr, linshiCards)
								tempV := get_HandCardValue(arr, nil)

								if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
									BestHandCardValue = tempV
									BestMaxCards = tempCars
									PutCards = true
								}

							}

						}

					} else if length == 3 {
						for j := 0; j < len(twoArr)-2; j++ {
							for k := j + 1; k < len(twoArr); k++ {
								for l := k + 1; l < len(twoArr); l++ {
									arr := copyNumEle(tempArr, 1)                                                             //复制一份
									values := copyNumEle(tempValues, 1)                                                       //也复制一份
									values = append(values, twoArr[j], twoArr[j], twoArr[k], twoArr[k], twoArr[l], twoArr[l]) //带上腿
									tempCars = getCardsFromValues(self, values)
									linshiCards := getCardsFromValues(tempArr, []int{twoArr[j], twoArr[j], twoArr[k], twoArr[k], twoArr[l], twoArr[l]})
									arr = deleEleFromArr(arr, linshiCards)

									tempV := get_HandCardValue(arr, nil)

									if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
										BestHandCardValue = tempV
										BestMaxCards = tempCars
										PutCards = true
									}
								}
							}

						}
					}

				}
			}
			if PutCards {
				return BestMaxCards
			}

			b, zhaArr := checkZhadanAndOne(self, selfp) //如果去除炸弹剩一手，妈的，直接炸了
			if b == true {
				if preNum <= 4 || len(otherp.Jishu.FourArr) <= 0 || otherp.Jishu.FourArr[len(otherp.Jishu.FourArr)-1] < GetPockValue(zhaArr[0]) {
					return zhaArr
				}

			}

			if outIdx == 1 {
				return []int{}
			}

			//考虑出炸弹
			if len(selfp.Jishu.FourArr) > 0 {
				for i := 0; i < len(selfp.Jishu.FourArr); i++ {
					v := selfp.Jishu.FourArr[i]
					tempCars := getCardsFromValues(self, []int{v, v, v, v})
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}
				if PutCards == true {
					return BestMaxCards
				}
			}
			//考虑王炸
			if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
				if BestHandCardValue.SumValue > 20 {
					return []int{551, 552}
				}
			}
			//不出
			return []int{}

		}
	} else if outp.Paixing == PAIXING_SIDAIER && outp.Sier != nil {

		if outp.Sier.Type == SIDAIER_DAN {
			//看看是不是正好大于并且一手出掉了
			if selfp.Paixing == PAIXING_SIDAIER && selfp.Sier != nil {
				if selfg.MaxCard > outg.MaxCard && selfp.Sier.Type == SIDAIER_DAN {
					return self
				}
			} else if selfp.Paixing == PAIXING_ZHADAN {
				return self
			} else if selfp.Paixing == PAIXING_DUIWANG {
				return self
			}

			//同类型牌压制
			BestHandCardValue := get_HandCardValue(self, selfp)
			BestHandCardValue.NeedRound += 1
			//暂存最佳的牌号
			var BestMaxCards []int
			//是否出牌的标志
			PutCards := false
			//顺带出去的牌
			//			tmp_1 := 0

			tempSelf := copyNumEle(self, 1)

			fourArr := copyNumEle(selfp.Jishu.FourArr, 1)
			oneArr := copyNumEle(selfp.Jishu.ValueArr, 1)

			if BestHandCardValue.SumValue <= 14 {

				for i := 0; i < len(fourArr); i++ {
					if fourArr[i] > outg.MaxCard { //大于就好
						v := fourArr[i]
						for j := 0; j < len(oneArr)-1; j++ {
							for k := j + 1; k < len(oneArr); k++ {
								if oneArr[j] != v && oneArr[k] != v {
									tempValues := []int{v, v, v, v, oneArr[j], oneArr[k]}
									tempCars := getCardsFromValues(self, tempValues)
									tempArr := copyNumEle(tempSelf, 1)
									tempArr = deleEleFromArr(tempArr, tempCars) //去掉

									tempV := get_HandCardValue(tempArr, nil)

									if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
										BestHandCardValue = tempV
										BestMaxCards = tempCars
										PutCards = true
									}
								}

							}
						}

					}
				}
				if PutCards {
					return BestMaxCards
				}
			}

			b, zhaArr := checkZhadanAndOne(self, selfp) //如果去除炸弹剩一手，妈的，直接炸了
			if b == true {
				if preNum <= 4 || len(otherp.Jishu.FourArr) <= 0 || otherp.Jishu.FourArr[len(otherp.Jishu.FourArr)-1] < GetPockValue(zhaArr[0]) {
					return zhaArr
				}

			}

			if outIdx == 1 {
				return []int{}
			}

			//考虑出炸弹
			if len(selfp.Jishu.FourArr) > 0 {
				for i := 0; i < len(selfp.Jishu.FourArr); i++ {
					v := selfp.Jishu.FourArr[i]
					tempCars := getCardsFromValues(self, []int{v, v, v, v})
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}
				if PutCards == true {
					return BestMaxCards
				}
			}
			//考虑王炸
			if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
				if BestHandCardValue.SumValue > 20 {
					return []int{551, 552}
				}
			}
			//不出
			return []int{}

		} else if outp.Sier.Type == SIDAIER_DUI {
			//看看是不是正好大于并且一手出掉了
			if selfp.Paixing == PAIXING_SIDAIER && selfp.Sier != nil {
				if selfg.MaxCard > outg.MaxCard && selfp.Sier.Type == SIDAIER_DUI {
					return self
				}
			} else if selfp.Paixing == PAIXING_ZHADAN {
				return self
			} else if selfp.Paixing == PAIXING_DUIWANG {
				return self
			}
			//同类型牌压制
			BestHandCardValue := get_HandCardValue(self, selfp)
			BestHandCardValue.NeedRound += 1
			//暂存最佳的牌号
			var BestMaxCards []int
			//是否出牌的标志
			PutCards := false
			//顺带出去的牌
			//			tmp_1 := 0

			tempSelf := copyNumEle(self, 1)

			fourArr := copyNumEle(selfp.Jishu.FourArr, 1)
			twoArr := append(selfp.Jishu.TwoArr, selfp.Jishu.ThreeArr...)

			if BestHandCardValue.SumValue <= 14 {
				for i := 0; i < len(fourArr); i++ {
					if fourArr[i] > outg.MaxCard { //大于就好
						v := fourArr[i]
						for j := 0; j < len(twoArr)-1; j++ {
							for k := j + 1; k < len(twoArr); k++ {
								if twoArr[j] != v && twoArr[k] != v {
									tempValues := []int{v, v, v, v, twoArr[j], twoArr[j], twoArr[k], twoArr[k]}
									tempCars := getCardsFromValues(self, tempValues)
									tempArr := copyNumEle(tempSelf, 1)
									tempArr = deleEleFromArr(tempArr, tempCars) //去掉

									tempV := get_HandCardValue(tempArr, nil)

									if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 {
										BestHandCardValue = tempV
										BestMaxCards = tempCars
										PutCards = true
									}
								}

							}
						}

					}
				}
				if PutCards {
					return BestMaxCards
				}
			}

			b, zhaArr := checkZhadanAndOne(self, selfp) //如果去除炸弹剩一手，妈的，直接炸了
			if b == true {
				if preNum <= 4 || len(otherp.Jishu.FourArr) <= 0 || otherp.Jishu.FourArr[len(otherp.Jishu.FourArr)-1] < GetPockValue(zhaArr[0]) {
					return zhaArr
				}

			}

			if outIdx == 1 {
				return []int{}
			}

			//考虑出炸弹
			if len(selfp.Jishu.FourArr) > 0 {
				for i := 0; i < len(selfp.Jishu.FourArr); i++ {
					v := selfp.Jishu.FourArr[i]
					tempCars := getCardsFromValues(self, []int{v, v, v, v})
					tempArr := copyNumEle(tempSelf, 1)

					tempArr = deleEleFromArr(tempArr, tempCars)

					tempV := get_HandCardValue(tempArr, nil)

					if BestHandCardValue.SumValue-BestHandCardValue.NeedRound*7 <= tempV.SumValue-tempV.NeedRound*7 || tempV.SumValue > 0 {
						BestHandCardValue = tempV
						BestMaxCards = tempCars
						PutCards = true
					}

				}
				if PutCards == true {
					return BestMaxCards
				}
			}
			//考虑王炸
			if IndexOf(self, 551) != -1 && IndexOf(self, 552) != -1 {
				if BestHandCardValue.SumValue > 20 {
					return []int{551, 552}
				}
			}
			//不出
			return []int{}

		}
	} else if outp.Paixing == PAIXING_ZHADAN {
		if outIdx == 1 {
			//不要
			return []int{}
		} else {
			return GetBigthan(self, outCards)
		}

	} else if outp.Paixing == PAIXING_DUIWANG {
		return []int{}

	}

	//检查玩家是不是正好可以一手出掉且大于当前出的牌

	return retArr
}

//获取介于某个牌值区间的牌 不包括小的 包括大的
func getDurationCards(from []int, min int, max int) []int {
	var retArr []int = make([]int, 0)

	if checkDuiwangAndOne(from) == true { //大王加一手牌
		return []int{551, 552}
	}

	return retArr

}

//检查当前手中的牌是否是一对王加一手牌
func checkDuiwangAndOne(from []int) bool {
	var idxX = IndexOf(from, 551)
	var idxD = IndexOf(from, 552)

	if idxX != -1 && idxD != -1 { //有对王
		tempFrom := copyNumEle(from, 1) //不要改变原来的
		tempFrom = deleEleFromArr(tempFrom, []int{551, 552})
		if GetPaixing(tempFrom).Paixing != -1 { //一手就可以出完
			return true
		}
	}
	return false
}

//检查当前手中是不是去除炸弹就一手牌，如果是就他妈要了，后期再升级判断其他的
func checkZhadanAndOne(from []int, fromp *PockPaixing) (bool, []int) {
	if fromp == nil {
		fromp = GetPaixing(from)
	}

	for i := 0; i < len(fromp.Jishu.FourArr); i++ {
		tempFrom := copyNumEle(from, 1)
		values := []int{}
		for j := 0; j <= i; j++ {
			values = append(values, fromp.Jishu.FourArr[j])
			values = append(values, fromp.Jishu.FourArr[j])
			values = append(values, fromp.Jishu.FourArr[j])
			values = append(values, fromp.Jishu.FourArr[j])
		}
		tempCards := getCardsFromValues(tempFrom, values)

		tempFrom = deleEleFromArr(tempFrom, tempCards)

		if GetPaixing(tempFrom).Paixing != PAIXING_NON {
			return true, getCardsFromValues(from, []int{fromp.Jishu.FourArr[0], fromp.Jishu.FourArr[0], fromp.Jishu.FourArr[0], fromp.Jishu.FourArr[0]})
		}

	}

	return false, []int{}

}
