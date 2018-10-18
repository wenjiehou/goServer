package paixingLogic

import (
	"sort"
)

const (
	FEIJI_NON int = 0
	FEIJI_DAN int = 1
	FEIJI_DUI int = 2
)

type FeijiData struct {
	Length   int //飞机的长度
	Type     int //飞机的类型
	BigValue int //飞机的最大值
}

//获得飞机牌型
func GetFeijiData(jishuData *JishuArrData) *FeijiData {
	var feijiData *FeijiData
	if len(jishuData.ThreeArr) == 0 {
		return nil
	}

	tempjishuData := &JishuArrData{}
	tempjishuData.Clone(jishuData)

	if len(tempjishuData.FourArr) == 0 {
		return preDeajishuData(tempjishuData)
	} else if len(jishuData.FourArr) == 1 {
		tempjishuData.Clone(jishuData)
		tempjishuData.ThreeArr = append(tempjishuData.ThreeArr, jishuData.FourArr[0])
		sort.Sort(IntSlice(tempjishuData.ThreeArr))

		tempjishuData.OneArr = append(tempjishuData.OneArr, jishuData.FourArr[0])
		sort.Sort(IntSlice(tempjishuData.OneArr))

		feijiData = preDeajishuData(tempjishuData)
		if feijiData != nil {
			return feijiData
		}

		//拆成22
		tempjishuData.Clone(jishuData)
		tempjishuData.TwoArr = append(tempjishuData.TwoArr, jishuData.FourArr[0], jishuData.FourArr[0])
		sort.Sort(IntSlice(tempjishuData.TwoArr))

		return preDeajishuData(tempjishuData)

	} else if len(jishuData.FourArr) >= 2 { //1313 只能拆1313 带的类型要相同
		tempjishuData.Clone(jishuData)

		for _, v := range jishuData.FourArr {
			tempjishuData.ThreeArr = append(tempjishuData.ThreeArr, v)
			tempjishuData.OneArr = append(tempjishuData.OneArr, v)
		}

		sort.Sort(IntSlice(tempjishuData.ThreeArr))
		return preDeajishuData(tempjishuData)
	}

	return nil
}

func preDeajishuData(jishuData *JishuArrData) *FeijiData {
	tempjishuData := &JishuArrData{}
	tempjishuData.Clone(jishuData)
	if len(tempjishuData.ThreeArr) < 2 {
		return nil
	}

	if CheckValueLian(tempjishuData.ThreeArr) == true {
		return getFeijiWithNoFour(tempjishuData)
	} else { //不连续就要拿出一个不连续的出来
		for i := 0; i < len(tempjishuData.ThreeArr); i++ {
			if i == 0 {
				if tempjishuData.ThreeArr[i] != tempjishuData.ThreeArr[i+1]-1 {
					tempjishuData.OneArr = append(tempjishuData.OneArr, tempjishuData.ThreeArr[i], tempjishuData.ThreeArr[i], tempjishuData.ThreeArr[i])
					sort.Sort(IntSlice(tempjishuData.OneArr)) //对单张进行重新排序
					tempjishuData.ThreeArr = append(tempjishuData.ThreeArr[:i], tempjishuData.ThreeArr[i+1:]...)
					break
				}

			} else {
				if tempjishuData.ThreeArr[i] != tempjishuData.ThreeArr[i-1]+1 {
					tempjishuData.OneArr = append(tempjishuData.OneArr, tempjishuData.ThreeArr[i], tempjishuData.ThreeArr[i], tempjishuData.ThreeArr[i])
					sort.Sort(IntSlice(tempjishuData.OneArr)) //对单张进行重新排序
					tempjishuData.ThreeArr = append(tempjishuData.ThreeArr[:i], tempjishuData.ThreeArr[i+1:]...)
					break
				}

			}
		}
		return getFeijiWithNoFour(tempjishuData)
	}

	return nil

}

/**请使用排好序的牌传入,四个的计数必须为零*/
func getFeijiWithNoFour(jishuData *JishuArrData) *FeijiData {

	lenThree := len(jishuData.ThreeArr)
	lenTwo := len(jishuData.TwoArr)
	lenOne := len(jishuData.OneArr)

	if lenThree < 2 {
		return nil
	}

	var feijiData *FeijiData

	var threeLian = CheckValueLian(jishuData.ThreeArr)

	if threeLian {
		if lenTwo == 0 && lenOne == 0 && lenThree > 1 {
			feijiData = &FeijiData{}
			feijiData.BigValue = jishuData.ThreeArr[lenThree-1]
			feijiData.Length = lenThree
			feijiData.Type = FEIJI_NON
			return feijiData
		} else if lenThree == lenOne+2*lenTwo {
			feijiData = &FeijiData{}
			feijiData.BigValue = jishuData.ThreeArr[lenThree-1]
			feijiData.Length = lenThree
			feijiData.Type = FEIJI_DAN
			return feijiData
		} else if lenThree == lenTwo && lenOne == 0 {
			feijiData = &FeijiData{}
			feijiData.BigValue = jishuData.ThreeArr[lenThree-1]
			feijiData.Length = lenThree
			feijiData.Type = FEIJI_DUI
			return feijiData
		}
	}

	return nil

}
