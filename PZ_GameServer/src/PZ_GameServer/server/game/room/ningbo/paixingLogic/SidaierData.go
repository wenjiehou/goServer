package paixingLogic

const (
	SIDAIER_DAN int = 1
	SIDAIER_DUI int = 2
)

type SidaierData struct {
	Value int //比较的最大值
	Type  int //类型
}

func GetSidaierData(jishuData *JishuArrData) *SidaierData {
	var si *SidaierData
	lenFour := len(jishuData.FourArr)
	lenThree := len(jishuData.ThreeArr)
	lenTwo := len(jishuData.TwoArr)
	lenOne := len(jishuData.OneArr)

	if lenFour == 1 { //只有一个四个
		if lenThree == 0 { //三个必须没有
			if lenOne == 2 && lenTwo == 0 { //一个有两张
				si = &SidaierData{}
				si.Value = jishuData.FourArr[0]
				si.Type = SIDAIER_DAN
				return si
			} else if lenOne == 0 && lenTwo == 1 {
				si = &SidaierData{}
				si.Value = jishuData.FourArr[0]
				si.Type = SIDAIER_DAN
				return si
			} else if lenOne == 0 && lenTwo == 2 {
				si = &SidaierData{}
				si.Value = jishuData.FourArr[0]
				si.Type = SIDAIER_DUI
				return si
			}
		}

	} else if lenFour == 2 { //四带sige
		if lenThree == 0 && lenOne == 0 && lenTwo == 0 {
			si = &SidaierData{}
			si.Value = jishuData.FourArr[1]
			si.Type = SIDAIER_DUI
			return si
		}
	}

	return nil
}
