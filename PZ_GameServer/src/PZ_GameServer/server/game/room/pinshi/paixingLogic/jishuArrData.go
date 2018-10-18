package paixingLogic

import (
	"sort"
)

type JishuArrData struct {
	ValueArr  []int
	TypeArr   []int
	NValueArr []int
	IsHuaArr  []bool

	// 5 6 7 8需要再扩展
	FourArr  []int
	ThreeArr []int
	TwoArr   []int
	OneArr   []int
}

func GetJishuArrData(cards []int) *JishuArrData {

	jishuData := &JishuArrData{
		ValueArr:  make([]int, 0),
		TypeArr:   make([]int, 0),
		NValueArr: make([]int, 0),
		IsHuaArr:  make([]bool, 0),

		FourArr:  make([]int, 0),
		ThreeArr: make([]int, 0),
		TwoArr:   make([]int, 0),
		OneArr:   make([]int, 0),
	}

	clen := len(cards)
	for i := 0; i < clen; i++ {
		jishuData.ValueArr = append(jishuData.ValueArr, GetPockValue(cards[i]))
		jishuData.TypeArr = append(jishuData.TypeArr, GetPockType(cards[i]))
		jishuData.NValueArr = append(jishuData.NValueArr, GetPockValueNN(cards[i]))
		jishuData.IsHuaArr = append(jishuData.IsHuaArr, CheckIsHua(cards[i]))
	}
	valueArr := make([]int, len(jishuData.ValueArr))
	for i, v := range jishuData.ValueArr {
		valueArr[i] = v
	}

	for i := 0; i < len(valueArr); i++ {
		num := 1
		tempV := valueArr[i]
		valueArr = append(valueArr[:i], valueArr[i+1:]...)
		i -= 1
		idx := IndexOf(valueArr, tempV)
		for {
			if idx == -1 {
				break
			}
			valueArr = append(valueArr[:idx], valueArr[idx+1:]...)
			num += 1
			idx = IndexOf(valueArr, tempV)
		}
		switch num {
		case 1:
			jishuData.OneArr = append(jishuData.OneArr, tempV)
		case 2:
			jishuData.TwoArr = append(jishuData.TwoArr, tempV)
		case 3:
			jishuData.ThreeArr = append(jishuData.ThreeArr, tempV)
		case 4:
			jishuData.FourArr = append(jishuData.FourArr, tempV)
		}

	}

	return jishuData
}

func IndexOf(cards []int, ele int) int {
	for i, v := range cards {
		if v == ele {
			return i
		}
	}
	return -1
}

//判断值是否连续
func CheckValueLian(valueArr []int) bool {
	len := len(valueArr)
	for i := 0; i < len-1; i++ {
		if valueArr[i+1]-valueArr[i] != 1 {
			return false
		}
	}
	return true
}

/**将一个数组中的14 16 转化成 1 2 并且进行排序*/
func ConversValueArr(arr []int) {
	clen := len(arr)

	for i := 0; i < clen; i++ {
		if arr[i] == 14 {
			arr[i] = 1
		} else if arr[i] == 16 {
			arr[i] = 2
		}
	}
	sort.Sort(IntSlice(arr))
}
