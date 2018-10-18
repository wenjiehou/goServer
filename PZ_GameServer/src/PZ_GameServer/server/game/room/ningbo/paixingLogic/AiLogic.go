package paixingLogic

//对主动出出去的牌做一个优化
func freeDeal(putPaixing *PockPaixing, paixingList []*PockPaixing) []int {
	if putPaixing.Paixing == PAIXING_DAN { //如果是单张的牌型，我们比较一下3+1的屁股和它谁大
		for _, v := range paixingList {
			if v.Paixing == PAIXING_SANYI {
				if v.Jishu.OneArr[0] < putPaixing.Jishu.OneArr[0] {
					putPaixing.Jishu.OneArr = v.Jishu.OneArr
					putPaixing.Cards = getCardsFromValues(v.Cards, v.Jishu.OneArr)
				}
			}
		}
	} else if putPaixing.Paixing == PAIXING_DUIZI {
		for _, v := range paixingList {
			if v.Paixing == PAIXING_SANER {
				if v.Jishu.TwoArr[0] < putPaixing.Jishu.TwoArr[0] {
					putPaixing.Jishu.TwoArr = v.Jishu.TwoArr
					putPaixing.Cards = getCardsFromValues(v.Cards, copyNumEle(v.Jishu.TwoArr, 2))
				}
			}
		}
	} else if putPaixing.Paixing == PAIXING_SANYI {
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

}
