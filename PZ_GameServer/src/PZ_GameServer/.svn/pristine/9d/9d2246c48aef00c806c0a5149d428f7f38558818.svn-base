package statement

import (
	al "PZ_GameServer/common/util/arrayList"
)

const (
	cardLen = 42 //
)

// ------------------------
// 得到胡牌类型
// listCount = 手牌的数量(不包括吃碰杠)
// startCount = 出牌的次数 0=天胡  1=地胡
func (sc *StatementCtl) GetHuType(mjs []int, seatIndex int, room interface{}) (string, int) {

	return "", 0
}

//	mjs[0] = 0 // 万  (0-8)
//	mjs[1] = 0
//	mjs[2] = 0
//	mjs[3] = 0
//	mjs[4] = 0
//	mjs[5] = 0
//	mjs[6] = 3
//	mjs[7] = 0
//	mjs[8] = 0

//	mjs[9] = 0 // 饼 (9-17)
//	mjs[10] = 0
//	mjs[11] = 1
//	mjs[12] = 1
//	mjs[13] = 4
//	mjs[14] = 0
//	mjs[15] = 0
//	mjs[16] = 0
//	mjs[17] = 0

//	mjs[18] = 0 // 条 (18-26)
//	mjs[19] = 0
//	mjs[20] = 0
//	mjs[21] = 0
//	mjs[22] = 0
//	mjs[23] = 0
//	mjs[24] = 0
//	mjs[25] = 0
//	mjs[26] = 0

//	mjs[27] = 0  // 东 风(27-30)
//	mjs[28] = 0  // 南
//	mjs[29] = 0  // 西
//	mjs[30] = 0  // 北
// 											字牌(27-33)
//	mjs[31] = 0  // 中 箭(31-33)
//	mjs[32] = 0  // 发
//	mjs[33] = 0  // 白

//	mjs[34] = 0  // 春   花牌, 各一张
//	mjs[35] = 0  // 夏
//	mjs[36] = 0  // 秋
//	mjs[37] = 0  // 冬
//	mjs[38] = 0  // 梅
//	mjs[39] = 0  // 兰
//	mjs[40] = 0  // 竹
//	mjs[41] = 0  // 菊

//【对对胡】   四副刻子加一对将
func (sc *StatementCtl) F_DuiDuiHu(mjs []int) (string, int) {
	msg := ""
	jiang := 0 // 将
	ke := 0    // 刻
	for i := 0; i < len(mjs); i++ {
		if mjs[i] == 0 {
			continue
		}
		if mjs[i]%2 == 0 {
			jiang++
		} else if mjs[i]%3 == 0 {
			ke++
		} else {
			return msg, 0
		}
	}
	if jiang == 1 && ke >= 1 {
		return "对对胡", 2
	}
	return msg, 0
}

// 【清一色】
func (sc *StatementCtl) F_QingYiSe(mjs []int) (string, int) {

	t := -1 //类型
	for i := 0; i < len(mjs); i++ {
		if mjs[i] == 0 {
			continue
		}
		if i >= 0 && i < 9 && (t < 0 || t == 0) {
			t = 0
		} else if i >= 9 && i < 18 && (t < 0 || t == 1) {
			t = 1
		} else if i >= 18 && i < 27 && (t < 0 || t == 2) {
			t = 2
		} else {
			return "", 0
		}
	}

	if t >= 0 {
		return "清一色", 4
	}

	return "", 0
}

// 【断幺九】2台，没有幺九，没有风，没有中发白。
func (sc *StatementCtl) F_Duan19(mjs []int) (string, int) {

	if mjs[0] > 0 || mjs[8] > 0 || mjs[9] > 0 || mjs[17] > 0 || mjs[18] > 0 || mjs[26] > 0 {
		return "", 0
	}

	for i := 27; i <= 33; i++ {
		if mjs[i] > 0 {
			return "", 0
		}
	}

	return "断幺九", 2
}

//【带幺九】 玩家手牌中，全部是用1的连牌或者9的连牌组成的牌； (123) , (789)
func (sc *StatementCtl) F_Dai19(mjs []int) (string, int) {

	if mjs[3] > 0 || mjs[4] > 0 || mjs[5] > 0 ||
		mjs[12] > 0 || mjs[13] > 0 || mjs[14] > 0 ||
		mjs[21] > 0 || mjs[22] > 0 || mjs[23] > 0 {
		return "", 0 // 判断所有 4,5,6
	}

	yyu := []int{1, 7, 10, 16, 19, 25}
	for y := 0; y < len(yyu); y++ { // 去掉所有 1,2,3   7,8,9
		c := yyu[y]
		h := mjs[c]
		if h > 0 {
			for i := 0; i < h; i++ {
				mjs[c-1]--
				mjs[c]--
				mjs[c+1]--
			}
			if mjs[c-1] != 0 || mjs[c] != 0 || mjs[c+1] != 0 {
				return "", 0
			}
		}
	}

	jiang := 0
	ke := 0
	for i := 0; i < len(mjs); i++ {
		if mjs[i] == 0 {
			continue
		}
		if mjs[i]%2 == 0 {
			jiang++
		} else if mjs[i]%3 == 0 {
			ke++
		} else {
			return "", 0
		}
	}

	if jiang == 1 && ke > 0 {
		return "带幺九", 1
	}
	return "", 0
}

//【七对】 玩家的手牌全部是两张一对的，没有碰过和杠过；
func (sc *StatementCtl) F_QiDui(mjs []int) int {

	duiCount := 0
	for i := 0; i < len(mjs); i++ {
		if mjs[i] == 0 {
			continue
		}
		if mjs[i]%2 != 0 {
			duiCount = 0
			break
		} else {
			duiCount++
		}
	}
	if duiCount == 7 {
		return 1
	}
	return 0
}

//【金钩钓】 玩家和牌时，其他牌都被用作碰牌，杠牌。手牌中只剩下唯一的一张牌，不计对对胡；
func (sc *StatementCtl) F_JinGouGou(list *al.ArrayList) int {
	if list.Count == 1 {
		return 4
	} else {
		return 0
	}
}

// 【4张一样】 获得4张1样的
func (sc *StatementCtl) F_Geng(mjs []int) int {
	geng := 0
	for i := 0; i < len(mjs); i++ {
		if mjs[i] == 0 {
			continue
		}
		if mjs[i]%4 != 0 {
			geng++
		}
	}
	return geng
}
