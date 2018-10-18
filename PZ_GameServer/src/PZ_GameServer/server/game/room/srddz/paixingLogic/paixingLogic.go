package paixingLogic

import (
	"fmt"
	"math"
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

//扑克牌型的定义
const (
	//单张
	PAIXING_DAN = 1
	//对子
	PAIXING_DUIZI = 2
	//三个
	PAIXING_SAN = 3
	//三个带2
	PAIXING_SANER = 4
	//连对
	PAIXING_LIANDUI = 5
	//顺子
	PAIXING_SHUNZI = 6
	//飞机
	PAIXING_FEIJI = 7
	//炸弹
	PAIXING_ZHADAN = 8
	//对王
	PAIXING_TIANZHA = 9
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
	return 1
	//	if t == PAIXING_ZHADAN {
	//		return 2
	//	} else if t == PAIXING_DUIWANG {
	//		return 4
	//	} else {
	//		return 1
	//	}
}
