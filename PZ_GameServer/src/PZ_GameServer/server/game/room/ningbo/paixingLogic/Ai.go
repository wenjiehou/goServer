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
func FreeDizhuOutput(self []int, preCards []int, nexCards []int, other []int,
	preStep [][]int, nexStep [][]int) []int {

	//	preNum := len(preCards)
	//	nexNum := len(nexCards)

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
				if GetPaixing(otherBigCards).Paixing == PAIXING_ZHADAN { //比我大的只有炸弹，这个时候
					putPaixing = bigestPaixing
				}

			}
		} else if len(retList) == 3 { //
			//尽量出一手自己要的回来的
			if lowestPaixing.Paixing == PAIXING_SANER && bigestPaixing.Paixing == PAIXING_SANYI {
				//交换一下屁股出 3+1
				putPaixing = lowestPaixing
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
func FreeDizhuSOutput(self []int, preCards []int, nexCards []int, other []int,
	preStep [][]int, nexStep [][]int) []int {

	//	preNum := len(preCards)
	nexNum := len(nexCards)

	_, retList := getBestList(self)
	lowest := 100
	bigest := -100
	var lowestPaixing *PockPaixing
	//var bigestPaixing *PockPaixing

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
				bigest = tmpg.Value
			}

			paixingList = append(paixingList, tmpp)
		}

		nexBigList := []*PockPaixing{}   //地主能大过的牌
		nexNoBigList := []*PockPaixing{} //地主不能大过的牌

		for i := 0; i < len(paixingList); i++ {
			if len(GetBigthan(nexCards, paixingList[i].Cards)) == 0 {
				nexNoBigList = append(nexNoBigList, paixingList[i])
			} else {
				nexBigList = append(nexBigList, paixingList[i])
			}
		}

		if len(nexBigList) == 0 { //地主都大不过，随便出
			putPaixing = lowestPaixing
		} else if len(nexBigList) == 1 { //地主能大过的只有一手牌，直接跑就好了
			if len(nexNoBigList) > 0 {
				putPaixing = nexNoBigList[0] //出地主要不到的
			} else {
				putPaixing = nexBigList[0]
			}
		} else if len(nexBigList) == 2 { //地主能大过的有两手
			if len(nexNoBigList) >= 1 { //我有一手地主要不到，最好出一个一样的牌型
				for _, p := range nexBigList {
					if putPaixing != nil { //找到了就出来
						break
					}
					for _, np := range nexNoBigList {
						if np.Paixing == p.Paixing {
							if len(GetBigthan(nexCards, p.Cards)) != nexNum { //地主不能一手牌就跑掉了
								putPaixing = p
								break
							}
						}
					}
				}
			}
		}

		if putPaixing == nil {
			//最好出一手价值很低，地主要不到的，但是盟友可以要得到的（注意价值不能太高，不然的话，大牌都出掉了，玩个屁）
			if len(nexNoBigList) > 0 { //有地主大不过的，取最小价值的
				var lower *PockPaixing
				for _, v := range nexNoBigList {
					if len(GetBigthan(preCards, v.Cards)) > 0 {
						if lower == nil {
							lower = v
						} else {
							if get_GroupData(v).Value < get_GroupData(lower).Value {
								lower = v
							}
						}
					}
				}
				if lower != nil { //如果这里的价值特别高，需要考虑一下
					//这个地方因为盟友可以接过去，我们后面来考虑 ，todo
					putPaixing = lower
				}
			}

		}

		if putPaixing != nil {
			//反正一定不能让地主走了
			if putPaixing == nil {
				if len(nexNoBigList) > 0 { //有地主大不过的，取最小价值的
					var lower *PockPaixing
					for _, v := range nexNoBigList {
						if lower == nil {
							lower = v
						} else {
							if get_GroupData(v).Value < get_GroupData(lower).Value {
								lower = v
							}
						}
					}
					if lower != nil { //如果这里的价值特别高，需要考虑一下
						//这里怎么说呢，如果是炸弹啊，对王啊这种，或者小王啊，大王啊，肯定不能出
						if lower.Paixing != PAIXING_ZHADAN && lower.Paixing != PAIXING_DUIWANG { //炸弹和对王出了没有意义，反正走不掉了
							if get_GroupData(lower).Value <= 1 {
								putPaixing = lower
							}
						}

					}
				}
			}
		}

		if putPaixing != nil {
			//出一个地主要得到，但是盟友要得到地主的牌，意思是盟友能接走
			for _, p := range nexBigList {
				nexBigCards := GetBigthan(nexCards, p.Cards)
				if len(nexBigCards) >= 0 {
					if len(nexBigCards) != nexNum && len(GetBigthan(preCards, nexBigCards)) > 0 { //盟友可以接过去
						if p.Paixing == PAIXING_DAN {
							if p.Jishu.OneArr[0] >= 12 {
								putPaixing = p
							}
						} else {
							putPaixing = p
						}

					}
				}
			}
		}

		if putPaixing == nil { //地主能接过去，盟友接不过去了，看看自己能不能接过来
			for _, p := range nexBigList {
				nexBigCards := GetBigthan(nexCards, p.Cards)
				if len(nexBigCards) >= 0 {
					if len(nexBigCards) != nexNum && len(GetBigthan(self, nexBigCards)) > 0 { //自己可以接过来
						if p.Paixing == PAIXING_DAN {
							if p.Jishu.OneArr[0] >= 12 {
								putPaixing = p
							}
						} else {
							putPaixing = p

						}
					}
				}
			}
		}

		//上面表示对家接不上了，常规的，我们出一个不是单张类型最小的牌
		if putPaixing == nil {
			var lower *PockPaixing
			for _, p := range paixingList {
				if len(GetBigthan(nexCards, p.Cards)) != nexNum {
					if p.Paixing != PAIXING_DAN {
						if lower == nil {
							lower = p
						} else {
							if get_GroupData(p).Value < get_GroupData(lower).Value {
								lower = p
							}
						}
					}
				}

			}
			if lower != nil {
				putPaixing = lower
			}
		}

		//这里需要考虑一种情况，我出了一个大牌，没人要得着，然后剩了一对小牌的情况

		if putPaixing == nil { //地主都能大过，我就出最小的
			for i := 11; i >= 2; i-- {
				retCard := GetBigthan(self, []int{i})
				if len(retCard) > 0 {
					return retCard
				}
			}
		}

		if putPaixing == nil {
			putPaixing = lowestPaixing
		}

		if putPaixing.Paixing == PAIXING_SANYI {
			for _, v := range paixingList {
				if v.Paixing == PAIXING_DAN {
					if v.Jishu.OneArr[0] < putPaixing.Jishu.OneArr[0] {
						putPaixing.Jishu.OneArr = v.Jishu.OneArr
						putPaixing.Cards[3] = v.Cards[0]
					}
				}
			}
		} else if putPaixing.Paixing == PAIXING_SANER {
			for _, v := range paixingList {
				if v.Paixing == PAIXING_DUIZI {
					if v.Jishu.TwoArr[0] < putPaixing.Jishu.TwoArr[0] {
						putPaixing.Jishu.TwoArr = v.Jishu.TwoArr
						putPaixing.Cards[3] = v.Cards[0]
						putPaixing.Cards[4] = v.Cards[1]
					}
				}
			}
		}

		return putPaixing.Cards

		//return freeDeal(putPaixing, paixingList)
	}
	return []int{}
}

//自己出的，随便出 地主的下家
func FreeDizhuXOutput(self []int, preCards []int, nexCards []int, other []int,
	preStep [][]int, nexStep [][]int) []int {

	//	preNum := len(preCards)
	//	nexNum := len(nexCards)

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

		//下面这一部分是帮助盟友出牌
		_, nexList := getBestList(nexCards)
		if len(nexList) == 1 { //妈的只有一手了，我要尽量满足他
			nexPaixing := GetPaixing(nexList[0])
			selfp := GetPaixing(self)
			if nexPaixing.Paixing == PAIXING_DAN {
				for _, c := range self {
					if GetPockValue(c) < nexPaixing.Jishu.OneArr[0] {
						return []int{c}
					}
				}
			} else if nexPaixing.Paixing == PAIXING_DUIZI {
				for i := nexPaixing.Jishu.TwoArr[0] + 1; i <= 16; i++ {
					if i == 17 {
						continue
					}
					if IndexOf(selfp.Jishu.TwoArr, i) != -1 || IndexOf(selfp.Jishu.ThreeArr, i) != -1 || IndexOf(selfp.Jishu.FourArr, i) != -1 {
						return getCardsFromValues(self, []int{i, i})
					}
				}
			} else if nexPaixing.Paixing == PAIXING_SAN { //三个
				for i := nexPaixing.Jishu.ThreeArr[0] + 1; i <= 16; i++ {
					if i == 17 {
						continue
					}
					if IndexOf(selfp.Jishu.ThreeArr, i) != -1 || IndexOf(selfp.Jishu.FourArr, i) != -1 {
						return getCardsFromValues(self, []int{i, i, i})
					}
				}
			} else if nexPaixing.Paixing == PAIXING_SANYI {
				for i := nexPaixing.Jishu.ThreeArr[0] + 1; i <= 16; i++ {
					if i == 17 {
						continue
					}
					if IndexOf(selfp.Jishu.ThreeArr, i) != -1 || IndexOf(selfp.Jishu.FourArr, i) != -1 {
						tempCards := getCardsFromValues(self, []int{i, i, i})
						tempSelf := copyNumEle(self, 1)
						tempSelf = deleEleFromArr(tempSelf, tempCards)
						if len(tempSelf) > 0 {
							return append(tempCards, tempSelf[0])
						}
					}
				}
			} else if nexPaixing.Paixing == PAIXING_SANER {
				for i := nexPaixing.Jishu.ThreeArr[0] + 1; i <= 16; i++ {
					if i == 17 {
						continue
					}
					if IndexOf(selfp.Jishu.ThreeArr, i) != -1 || IndexOf(selfp.Jishu.FourArr, i) != -1 {
						tempCards := getCardsFromValues(self, []int{i, i, i})
						tempSelf := copyNumEle(self, 1)
						tempSelf = deleEleFromArr(tempSelf, tempCards)
						tempSelfj := GetJishuArrData(tempSelf)

						if len(tempSelfj.TwoArr) > 0 {
							return append(tempCards, getCardsFromValues(tempSelf, []int{tempSelfj.TwoArr[0], tempSelfj.TwoArr[0]})...)
						}

						if len(tempSelfj.ThreeArr) > 0 {
							return append(tempCards, getCardsFromValues(tempSelf, []int{tempSelfj.ThreeArr[0], tempSelfj.ThreeArr[0]})...)
						}

						if len(tempSelfj.FourArr) > 0 {
							return append(tempCards, getCardsFromValues(tempSelf, []int{tempSelfj.FourArr[0], tempSelfj.FourArr[0]})...)
						}

					}
				}
			} else if nexPaixing.Paixing == PAIXING_LIANDUI {

			} else if nexPaixing.Paixing == PAIXING_FEIJI {

			}

		}

		preBigList := []*PockPaixing{}   //地主能大过的牌
		preNoBigList := []*PockPaixing{} //地主不能大过的牌

		for i := 0; i < len(paixingList); i++ {
			if len(GetBigthan(preCards, paixingList[i].Cards)) == 0 {
				preNoBigList = append(preNoBigList, paixingList[i])
			} else {
				preBigList = append(preBigList, paixingList[i])
			}
		}

		if len(preBigList) == 0 { //随便出
			putPaixing = lowestPaixing
		} else if len(preBigList) == 1 { //地主只能要的起一手，俺就把这手最后出
			if len(preNoBigList) > 0 {
				putPaixing = preNoBigList[0] //出地主要不到的
			} else {
				putPaixing = preBigList[0]
			}
		} else if len(preBigList) == 2 {

		}

		if putPaixing == nil { //前面没有取到就取小的
			putPaixing = bigestPaixing
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

	//	preNum := len(preCards)d
	nexNum := len(nexCards) //地主的牌，哈哈哈，这个玩意其实有点作弊了，但是我觉得作弊不明显就好了

	outp := GetPaixing(outCards)
	selfp := GetPaixing(self)
	outg := get_GroupData(outp)
	selfg := get_GroupData(selfp)

	selfBigCards := GetBigthan(self, outCards)

	if len(selfBigCards) == len(self) { //自己能跑的，赶紧跑
		return self
	}
	//如果地主大不了这个牌，我就不要
	if outIdx == -1 {
		dizhuBiger := GetBigthan(nexCards, outCards) //地主要不起的牌，我肯定不要
		dizhuBigerP := GetPaixing(dizhuBiger)
		if len(dizhuBiger) == 0 {
			if len(selfBigCards) > 0 { //我能要，如果我要了只剩一手，并且我要了，地主要不到，我肯定要
				tempSelf := copyNumEle(self, 1)
				tempSelf = deleEleFromArr(tempSelf, selfBigCards)

				if GetPaixing(tempSelf).Paixing != PAIXING_NON { //我能跑，所以我跑了
					return selfBigCards
				}
			}
			return []int{}
		} else if dizhuBigerP.Paixing == PAIXING_ZHADAN || dizhuBigerP.Paixing == PAIXING_DUIWANG {
			return []int{}
		}
	} //这一段待完善

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

		} else {
			mustBig = true //地主出的一定要接
			if nexNum == 1 {
				intentCard = nexCards[0] //和地主剩余的一样大就好了
			} else { //地主家不是一张
				intentCard = 12
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
		} else {
			//能跑就跑吧
			tempRet := GetBigthan(self, outCards)
			if len(tempRet) == 1 {
				return tempRet
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

					for j := start_i; j <= end_i; j++ {
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

					for j := start_i; j <= end_i; j++ {
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
	//	nexNum := len(nexCards)

	var retArr = make([]int, 0)

	if checkDuiwangAndOne(self) == true { //大王加一手牌
		return []int{551, 552}
	}

	outp := GetPaixing(outCards)
	selfp := GetPaixing(self)
	outg := get_GroupData(outp)
	selfg := get_GroupData(selfp)

	otherp := GetPaixing(other)

	if outIdx == 1 { //盟友一张，地主不要，当然不能要咯
		//如果盟友只剩一手牌，就不要
		_, nexList := getBestList(nexCards)
		if len(nexList) == 1 { //盟友就剩一手了
			return []int{}
		} else if len(nexList) == 2 { //如果盟友剩两手牌，大的大于地主，也不要
			for i := 0; i < len(nexList); i++ {
				if len(GetBigthan(preCards, nexList[i])) == 0 {
					return []int{}
				}
			}
		} else { //如果下家只有一手牌地主可以要得起，我们也不要
			preBigNexNum := 0
			for i := 0; i < len(nexList); i++ {
				if len(GetBigthan(preCards, nexList[i])) > 0 {
					preBigNexNum += 1
				}
			}

			if preBigNexNum <= 1 { //地主只有一手大过玩家的，不要
				return []int{}
			}
		}

	}

	//如果我的下家能要得起，并且不是炸弹，我就不考虑炸弹了
	var needZhadan bool = true
	if len(GetBigthan(nexCards, outCards)) > 0 {
		needZhadan = false
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

		if needZhadan == false {
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

		if needZhadan == false {
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

		if needZhadan == false {
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

		if needZhadan == false {
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

		if needZhadan == false {
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

		if needZhadan == false {
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

		if needZhadan == false {
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

			if needZhadan == false {
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

					for j := start_i; j <= end_i; j++ {
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

			if needZhadan == false {
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

					for j := start_i; j <= end_i; j++ {
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

			if needZhadan == false {
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

			if needZhadan == false {
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

			if needZhadan == false {
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
