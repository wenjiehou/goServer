package paixingLogic

type NiuData struct {
	Cards    []int
	ThreeArr []int
	TwoArr   []int
	NiuNum   int
}

/**请使用排好序的牌传入*/
func GetNiuData(cards []int) *NiuData {
	var niuData = &NiuData{
		NiuNum: 0,
	}
	niuData.Cards = append(cards)
	//一共就五张牌
	var clen = len(cards)
	//	var newArray = make([]int, 0)
	var twoV = 0

	for i := 0; i < clen; i++ {
		for j := i + 1; j < clen; j++ {
			for k := j + 1; k < clen; k++ {
				if (GetPockValueNN(cards[i])+GetPockValueNN(cards[j])+GetPockValueNN(cards[k]))%10 == 0 {
					var threeArr = []int{cards[i], cards[j], cards[k]}
					var twoArr = make([]int, 0)
					//求另外两个
					twoV = 0
					for n := 0; n < clen; n++ {
						if n != i && n != j && n != k {
							twoV += GetPockValueNN(cards[n])
							twoArr = append(twoArr, cards[n])
						}
					}
					var tempV = twoV % 10
					if tempV == 0 {
						tempV = 10
					}

					if tempV > niuData.NiuNum {
						niuData.NiuNum = tempV
						niuData.ThreeArr = threeArr
						niuData.TwoArr = twoArr
					}
				}
			}

		}
	}
	return niuData
}
