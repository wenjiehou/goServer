/*
 * +-----------------------------------------------
 * | xizhou_statemnet_fan.go
 * +-----------------------------------------------
 * | Version: 1.0
 * +-----------------------------------------------
 * | Context: 象山麻将算番 GameType : 3000
 * +-----------------------------------------------
 */
package xiangshan

import (
	al "PZ_GameServer/common/util/arrayList"
	"PZ_GameServer/protocol/def"
	rb "PZ_GameServer/server/game/roombase"
	st "PZ_GameServer/server/game/statement"
	"strconv"
)

/*

   默认清12混8
 		   GameType + Index
【花】2台     	3001
【风】2台	     	3002
【中发白】25台  	3003
【天胡】24台    	3004
【地胡】16台	   3005
【拉扛胡】1台	   3006
【送杠胡】0台	   3007
【杠上开花】1台   3008
【海底捞月】1台   3009
【边倒】1台	   3010
【嵌倒】1台	   3011
【单吊】1台	   3012
【对倒】1台	   3013
【断幺九】2台	   3014 // 没有
【自摸】1台	   3015
【门清】1台	   3016
【大吊车】8台	   3017
【混一色】8台	   3018
【对对胡】8台	   3019
【全顺子】1台	   3020
【清一色】12台	   3021
【风一色】40台	   3022
【板高】6台	   3023 // 没有
【大四喜】50台	   3024
*/

// 【花】2台，自己风位上的花（A是东风位，那么春，梅就是正花，每张2台，春梅齐全是 4台）
//  东 = 春 梅  正花 (2台)
//  南 = 夏 兰
//  西 = 秋 竹
//  北 = 冬 菊
//  不是正花1台
//  (8花是覆盖关系)
func (sc *XiangShan_Statement) F_Hua(seatIndex int) *st.FanResult {
	// directIndex  0 - 3   东 南 西 北
	// 0 - 7  春夏秋冬梅兰菊竹
	tp := 3001       //
	msg := ""        //
	tai := 0         // 台数
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	huaList := room.Seats[seatIndex].Cards.Hua
	directIndex := room.Seats[seatIndex].Direct // 方位
	hua := new([8]int)
	for i := 0; i < huaList.Count; i++ {
		if *huaList.Index(i) != nil {
			card := (*huaList.Index(i)).(*rb.Card)
			hua[card.ID-34]++
		}

	}

	if hua[0] > 0 && hua[1] > 0 && hua[2] > 0 && hua[3] > 0 && hua[4] > 0 && hua[5] > 0 && hua[6] > 0 && hua[7] > 0 {
		if sc.BaseCtl.FanType == Qin10Hun6 {
			msg = "八花(15台)"
			tai += 15
		} else {
			msg = "八花(25台)"
			tai += 25 // 8花
		}
		return &st.FanResult{tp, msg, tai, SpecialSid}
	}

	if hua[0] > 0 && hua[1] > 0 && hua[2] > 0 && hua[3] > 0 {
		if sc.BaseCtl.FanType == Qin10Hun6 {
			msg += "四花(8台)"
			tai += 8
		} else {
			msg += "四花(12台)"
			tai += 12 // 4花
		}

		//tai += 12 // 4花
		//msg += "四花(12台)"
		hua[0] = 0
		hua[1] = 0
		hua[2] = 0
		hua[3] = 0
	}
	if hua[4] > 0 && hua[5] > 0 && hua[6] > 0 && hua[7] > 0 {
		if sc.BaseCtl.FanType == Qin10Hun6 {
			msg += "四花(8台)"
			tai += 8
		} else {
			msg += "四花(12台)"
			tai += 12 // 4花
		}

		hua[4] = 0
		hua[5] = 0
		hua[6] = 0
		hua[7] = 0
	}

	huaCheck := []int{-1, -1}
	switch directIndex {
	case 0: // 东 梅
		huaCheck[0] = 0
		huaCheck[1] = 4
	case 1: // 南 兰
		huaCheck[0] = 1
		huaCheck[1] = 5
	case 2: // 西 菊
		huaCheck[0] = 2
		huaCheck[1] = 7
	case 3: // 北 竹
		huaCheck[0] = 3
		huaCheck[1] = 6
	}

	zhenghua := false
	zhenghuatai := 0
	yehua := false
	yehuatai := 0
	for i := 0; i < 8; i++ {
		if huaCheck[0] == i || huaCheck[1] == i {
			if hua[i] > 0 {
				zhenghua = true
				hua[i] = hua[i] * 2
				zhenghuatai += hua[i]
			}
		} else {
			if hua[i] > 0 {
				yehua = true
				yehuatai++
			}
		}
		tai += hua[i]
	}

	if zhenghua {
		msg += "正花(" + strconv.Itoa(zhenghuatai) + "台)"
	}
	if yehua {
		msg += " 野花(" + strconv.Itoa(yehuatai) + "台)"
	}

	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【风】2台，碰出、暗刻、杠牌 与 圈风+自身方位 同时符合才算正风。例如，当圈风为东风圈时，A的位置如果是东方位，那么东风碰出、暗刻、杠都算正风。
//
func (sc *XiangShan_Statement) F_Feng(seatIndex int) *st.FanResult {
	tp := 3002
	tai := 0 //台数
	msg := ""
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	FenIndex := room.Seats[seatIndex].Direct // 方位
	QuanFen := room.FengQuan                 // 圈风

	mjs := new([42]int) // 麻将
	for i := 0; i < room.Seats[seatIndex].Cards.List.Count; i++ {
		if *room.Seats[seatIndex].Cards.List.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.List.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Peng.Count; i++ {
		if *room.Seats[seatIndex].Cards.Peng.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Peng.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Kong.Count; i++ {
		if *room.Seats[seatIndex].Cards.Kong.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Kong.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}
	if seatIndex != room.CurIndex && room.CurCard != nil {
		mjs[room.CurCard.ID]++
	}
	// 27,28,29,30 东南西北
	// QuanFeng  0 - 3 东南西北
	feng := new([4]int)
	//fmt.Println(QuanFen+27, mjs[QuanFen+27])
	if mjs[QuanFen+27] >= 3 {
		feng[QuanFen]++
	}
	if mjs[FenIndex+27] >= 3 {
		feng[FenIndex]++
	}
	for i := 0; i < 4; i++ {
		if feng[i] == 2 {
			msg += "正风"
			tai += 2
		}
		if feng[i] == 1 {
			msg += "位风"
			tai += 1
		}
	}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【中发白】 (中发白碰出，暗刻,杠。每刻/杠1台。)
// 【中发白全】25台，同时全部碰出，暗刻，杠
func (sc *XiangShan_Statement) F_ZhongFaBai(seatIndex int) *st.FanResult {
	tp := 3003 //
	tai := 0
	msg := ""
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	mjs := new([42]int)
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	for i := 0; i < room.Seats[seatIndex].Cards.List.Count; i++ {
		if *room.Seats[seatIndex].Cards.List.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.List.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Peng.Count; i++ {
		if *room.Seats[seatIndex].Cards.Peng.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Peng.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Kong.Count; i++ {
		if *room.Seats[seatIndex].Cards.Kong.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Kong.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}

	if seatIndex != room.CurIndex && room.CurCard != nil {
		mjs[room.CurCard.ID]++
	}

	if mjs[31] >= 3 {
		tai++
		msg += "中"
	}
	if mjs[32] >= 3 {
		tai++
		msg += "发"
	}
	if mjs[33] >= 3 {
		tai++
		msg += "白"
	}
	if mjs[31] >= 3 && mjs[32] >= 3 && mjs[33] >= 3 {
		tai = 25
		msg = "大三元"
		if sc.BaseCtl.FanType == Qin10Hun6 {
			tai = 15
		}
	}

	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【天胡】 24台，庄家起手胡牌；
func (sc *XiangShan_Statement) F_TianHu(seatIndex int) *st.FanResult {
	tp := 3004 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	huaList := room.Seats[seatIndex].Cards.Hua
	if room.CurMJIndex-room.StartIndex-huaList.Count == 1 && seatIndex == room.CurIndex {
		msg = "天胡"
		tai = 24
		if sc.BaseCtl.FanType == Qin10Hun6 {
			tai = 18
		}
	}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【地胡】 16台，庄家打出的那张牌有人胡掉；
func (sc *XiangShan_Statement) F_DiHu(seatIndex int) *st.FanResult {
	tp := 3005 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	for _, v := range room.Seats {
		if !v.IsPutCard {
			msg = "地胡"
			tai = 16
			if sc.BaseCtl.FanType == Qin10Hun6 {
				tai = 12
			}
			break
		}
	}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【拉扛胡】1台  碰后自己暗杠(注意: 没有补牌), 别人胡这张杠的牌 . .   按照"自摸"算
func (sc *XiangShan_Statement) F_LaGangHu(seatIndex int) *st.FanResult {
	tp := 3006 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	laGangHuToolFlag := 0
	if sc.GetIsDoHu() {
		laGangHuToolFlag = 1
	}
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	if room.CurIndex != seatIndex {
		or := sc.Get(laGangHuToolFlag)
		if or != nil && or.Tool != nil && or.Tool.ToolType == st.T_PengKong {
			SpecialSid = room.CurIndex
			msg = "拉扛胡"
			tai = 1
		}
	}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【送杠胡】1台(杠后打的牌, 别人胡的这张牌)
func (sc *XiangShan_Statement) F_SongGangHu(seatIndex int) *st.FanResult {
	tp := 3007 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	songGangHuToolFlag := 0
	if sc.GetIsDoHu() {
		songGangHuToolFlag = 1
	}
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	if room.CurIndex == seatIndex {
		//自摸
		for i := songGangHuToolFlag; i < songGangHuToolFlag+8; i++ {
			or1 := sc.Get(i)
			if or1 != nil && or1.Tool.Index == room.CurIndex {
				if or1.Tool.ToolType == st.T_MoBack {
					continue
				} else if (or1.Tool.ToolType == st.T_Kong || or1.Tool.ToolType == st.T_PengKong) && or1.Tool.TIndex != room.CurIndex {
					msg = "送杠胡"
					tai = 1
					SpecialSid = or1.Tool.TIndex
					break
				}
			} else {
				break
			}
		}
	}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【杠上开花】1台，暗杠或先碰后杠杠牌后（包括花牌杠），从牌墙最后摸一张牌，胡牌；(补杠胡牌)
func (sc *XiangShan_Statement) F_KangShangKaiHua(seatIndex int) *st.FanResult {
	tp := 3008 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	kangShangKaiHuaToolFlag := 0
	if sc.GetIsDoHu() {
		kangShangKaiHuaToolFlag = 1
	}
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	if room.CurIndex == seatIndex {
		or1 := sc.Get(kangShangKaiHuaToolFlag)
		if or1 != nil && or1.Tool.Index == room.CurIndex {
			if or1.Tool.ToolType == st.T_Kong || or1.Tool.ToolType == st.T_AnKong || or1.Tool.ToolType == st.T_PengKong || or1.Tool.ToolType == st.T_MoBack {
				msg = "杠上开花"
				tai = 1
			}
		}

	}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【海底捞月】1台，摸最后一张牌的玩家自摸；(最后八墩牌)
func (sc *XiangShan_Statement) F_HaiDiLaoYue(seatIndex int) *st.FanResult {
	tp := 3009 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)

	if room.CurIndex != seatIndex {
		return &st.FanResult{tp, "", 0, SpecialSid}
	}

	// fmt.Println("room.AllCardLength ", room.AllCardLength, room.CurMJIndex)
	// sxleft := room.AllCardLength - room.CurMJIndex
	// sxLast := room.EndBlank % 2
	// if sxleft-sxLast == 16 {
	// 	msg = "海底捞月"
	// 	tai = 1
	// }
	overCount := room.AllCardLength - room.EndBlank - room.CurMJIndex - def.XiangShanDrawCount
	if overCount == 0 {
		msg = "海底捞月"
		tai = 1
	}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【边倒】1台  (1,2,3) (7,8,9)
// 【边倒，嵌倒，单吊，对倒】
// 1台，边倒： 1、2，胡出3，只能胡一张牌 。嵌倒：3、5，胡4，只能胡一张牌 。单吊：比如1、1、1、4，胡4，只能胡一张牌。对倒：比如 ABC AA BB  只能胡 AA BB；
func (sc *XiangShan_Statement) F_BianDao(seatIndex int) *st.FanResult {
	// 边倒只有3, 7可以
	tp := 3010
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	mjs := make([]int, 42)
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)

	if room.CurCard == nil {
		return &st.FanResult{tp, "", 0, SpecialSid}
	}

	curid := room.CurCard.ID

	// if room.Seats[seatIndex].Ting == 1 {
	if room.CurCard.Num == 2 || room.CurCard.Num == 6 {

		for i := 0; i < room.Seats[seatIndex].Cards.List.Count; i++ {
			if *room.Seats[seatIndex].Cards.List.Index(i) != nil {
				card := (*room.Seats[seatIndex].Cards.List.Index(i)).(*rb.Card)
				mjs[card.ID]++
			}

		}

		if seatIndex != room.CurIndex {
			mjs[curid]++
		}

		if room.CurCard.Num == 2 {
			mjs[curid]--
			mjs[curid-1]--
			mjs[curid-2]--
		}
		if room.CurCard.Num == 6 {
			mjs[curid]--
			mjs[curid+1]--
			mjs[curid+2]--
		}

		for i := 0; i < len(mjs); i++ {
			if mjs[i] < 0 {
				return &st.FanResult{tp, "", 0, SpecialSid}
			}
		}

		hu := sc.BaseCtl.CheckHu(mjs)

		if hu > 0 {
			msg = "边倒"
			tai = 1
		}
	}
	// }
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

//【嵌倒】1台 嵌倒：3、5，胡4，只能胡一张牌
func (sc *XiangShan_Statement) F_QianDao(seatIndex int) *st.FanResult {
	// 嵌倒 2345678
	tp := 3011 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	mjs := make([]int, 42)
	//if room.Seats[seatIndex].Ting == 1 {
	if room.CurCard.Num > 0 && room.CurCard.Num < 8 {
		for i := 0; i < room.Seats[seatIndex].Cards.List.Count; i++ {
			if *room.Seats[seatIndex].Cards.List.Index(i) != nil {
				card := (*room.Seats[seatIndex].Cards.List.Index(i)).(*rb.Card)
				mjs[card.ID]++
			}

		}

		if seatIndex != room.CurIndex {
			mjs[room.CurCard.ID]++
		}

		mjs[room.CurCard.ID-1]--
		mjs[room.CurCard.ID+1]--
		mjs[room.CurCard.ID]--
		for i := 0; i < len(mjs); i++ {
			if mjs[i] < 0 {
				return &st.FanResult{tp, "", 0, SpecialSid}
			}
		}
		hu := sc.BaseCtl.CheckHu(mjs)
		if hu > 0 {
			msg = "嵌倒"
			tai = 1
		}
	}
	//}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 11123

//【单吊】1台  单吊(将牌)：比如1、1、1、4，胡4，只能胡一张牌, (只有针对将牌)
func (sc *XiangShan_Statement) F_DanDiao(seatIndex int) *st.FanResult {
	tp := 3012 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	curid := room.CurCard.ID
	mjs := make([]int, 42)
	//if room.Seats[seatIndex].Ting == 1 {
	for i := 0; i < room.Seats[seatIndex].Cards.List.Count; i++ {
		if *room.Seats[seatIndex].Cards.List.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.List.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}

	}

	if seatIndex != room.CurIndex {
		mjs[curid]++
	}

	if mjs[curid] == 1 {
		return &st.FanResult{tp, "", 0, SpecialSid}
	}

	//如果是单吊就返回
	// if room.Seats[seatIndex].Cards.List.Count < 3 {
	// 	return &st.FanResult{tp, "", 0}
	// }
	mjs[curid] -= 2

	hu := sc.BaseCtl.CheckHu(mjs)

	if hu > 0 {
		msg = "单吊"
		tai = 1
	}
	//}

	return &st.FanResult{tp, msg, tai, SpecialSid}
}

//【对倒】1台  对倒(将牌)：比如 ABC AA BB  只能胡 AA BB；
func (sc *XiangShan_Statement) F_DuiDao(seatIndex int) *st.FanResult {
	tp := 3013 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	mjs := make([]int, 42)
	//if room.Seats[seatIndex].Ting == 2 {
	for i := 0; i < room.Seats[seatIndex].Cards.List.Count; i++ {
		if *room.Seats[seatIndex].Cards.List.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.List.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}

	}
	if seatIndex != room.CurIndex {
		mjs[room.CurCard.ID]++
	}

	mjs[room.CurCard.ID] -= 3
	if mjs[room.CurCard.ID] < 0 {
		return &st.FanResult{tp, "", 0, SpecialSid}
	}
	hu := sc.BaseCtl.CheckHu(mjs)
	if hu > 0 {
		msg = "对倒"
		tai = 1
	}
	//}

	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【断幺九】2台，没有幺九，没有风，没有中发白
func (sc *XiangShan_Statement) F_Duan19(seatIndex int) *st.FanResult {

	tp := 3014
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	//	mjs := make([]int, 42)
	//	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	//	for i := 0; i < room.Seats[seatIndex].Cards.List.Count; i++ {
	//		card := (*room.Seats[seatIndex].Cards.List.Index(i)).(*rb.Card)
	//		mjs[card.ID]++
	//	}
	//	for i := 0; i < room.Seats[seatIndex].Cards.Chow.Count; i++ {
	//		card := (*room.Seats[seatIndex].Cards.Chow.Index(i)).(*rb.Card)
	//		mjs[card.ID]++
	//	}
	//	for i := 0; i < room.Seats[seatIndex].Cards.Peng.Count; i++ {
	//		card := (*room.Seats[seatIndex].Cards.Peng.Index(i)).(*rb.Card)
	//		mjs[card.ID]++
	//	}
	//	for i := 0; i < room.Seats[seatIndex].Cards.Kong.Count; i++ {
	//		card := (*room.Seats[seatIndex].Cards.Kong.Index(i)).(*rb.Card)
	//		mjs[card.ID]++
	//	}

	//	if seatIndex != room.CurIndex {
	//		mjs[room.CurCard.ID]++
	//	}

	//	msg, tai = sc.BaseCtl.F_Duan19(mjs)

	//	if tai > 0 {
	//		tai = 2
	//	}

	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【自摸】1台，
func (sc *XiangShan_Statement) F_ZiMo(seatIndex int) *st.FanResult {
	tp := 3015 //
	tai := 0
	msg := ""
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	if room.CurIndex == seatIndex {
		msg = "自摸"
		tai = 1
	}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【门清】1台，没吃,没碰，没杠（暗杠算门清）
func (sc *XiangShan_Statement) F_MenQing(seatIndex int) *st.FanResult {
	tp := 3016 //
	tai := 0
	msg := ""
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	if room.Seats[seatIndex].Cards.Peng.Count == 0 &&
		room.Seats[seatIndex].Cards.Chow.Count == 0 {

		notAnKong := false
		for i := 0; i < room.Seats[seatIndex].Cards.Kong.Count; i++ {
			if *room.Seats[seatIndex].Cards.Kong.Index(i) != nil {
				card := (*room.Seats[seatIndex].Cards.Kong.Index(i)).(*rb.Card)
				if card.Status == 0 || card.Status == 2 { //0=明杠 1=暗杠 2=碰杠
					notAnKong = true
					break
				}
			}
		}

		if !notAnKong {
			tai = 1
			msg = "门清"
		}
	}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【大吊车】8台，手上只剩一张牌，
func (sc *XiangShan_Statement) F_DaDiaoChe(seatIndex int) *st.FanResult {
	tp := 3017 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	if room.CurIndex == seatIndex {
		if room.Seats[seatIndex].Cards.List.Count == 2 {
			msg = "大吊车"
			tai = 8
			if sc.BaseCtl.FanType == Qin10Hun6 {
				tai = 6
			}
		}
	} else {
		if room.Seats[seatIndex].Cards.List.Count == 1 {
			msg = "大吊车"
			tai = 8
			if sc.BaseCtl.FanType == Qin10Hun6 {
				tai = 6
			}
		}
	}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【混一色】8台，胡牌的牌由一色牌和字牌组成；(清一色+字牌)
func (sc *XiangShan_Statement) F_HunYiSe(seatIndex int) *st.FanResult {
	tp := 3018 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	mjs := make([]int, 6) // 麻将类型
	for i := 0; i < room.Seats[seatIndex].Cards.List.Count; i++ {
		if *room.Seats[seatIndex].Cards.List.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.List.Index(i)).(*rb.Card)
			mjs[card.Type] = 1
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Chow.Count; i++ {
		if *room.Seats[seatIndex].Cards.Chow.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Chow.Index(i)).(*rb.Card)
			mjs[card.Type] = 1
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Peng.Count; i++ {
		if *room.Seats[seatIndex].Cards.Peng.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Peng.Index(i)).(*rb.Card)
			mjs[card.Type] = 1
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Kong.Count; i++ {
		if *room.Seats[seatIndex].Cards.Kong.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Kong.Index(i)).(*rb.Card)
			mjs[card.Type] = 1
		}
	}

	if seatIndex != room.CurIndex {
		mjs[room.CurCard.Type] = 1
	}
	wbt := 0
	for i := 0; i < 3; i++ {
		wbt += mjs[i]
	}

	fj := 0
	if mjs[3] == 1 || mjs[4] == 1 {
		fj = 1
	}

	if wbt == 1 && fj == 1 {
		msg = "混一色"
		tai = 8
		//		if len(room.Rules.Rules) >= 4 {
		//			switch room.Rules.Rules[3] {
		//			case 128:
		//				tai = 8
		//				break
		//			case 106:
		//				tai = 6
		//				break
		//			default:
		//				tai = 8
		//				break
		//			}
		//		} else {
		//			tai = 8
		//		}
		if sc.BaseCtl.FanType == Qin10Hun6 {
			tai = 6
		}
	}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【对对胡】8台，胡出牌的牌型全由对子，杠，碰组成（没有顺子）；碰 杠 不算
func (sc *XiangShan_Statement) F_DuiDuiHu(seatIndex int) *st.FanResult {
	tp := 3019       //
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	if room.Seats[seatIndex].Cards.Chow.Count > 0 { // 有吃则直接返回
		return &st.FanResult{tp, "", 0, SpecialSid}
	}

	mjs := make([]int, 42) // 麻将
	for i := 0; i < room.Seats[seatIndex].Cards.List.Count; i++ {
		if *room.Seats[seatIndex].Cards.List.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.List.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}

	for i := 0; i < room.Seats[seatIndex].Cards.Peng.Count; i++ {
		if *room.Seats[seatIndex].Cards.Peng.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Peng.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}

	if seatIndex != room.CurIndex {
		mjs[room.CurCard.ID]++
	}

	msg, tai := sc.BaseCtl.F_DuiDuiHu(mjs)
	if tai > 0 {
		tai = 8
		if sc.BaseCtl.FanType == Qin10Hun6 {
			tai = 6
		}
	}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【全顺子】 1台 除一对将外，全部都是顺子，不能有刻子。  "如果有风位上的对子或中发白的对子也不能胡"
func (sc *XiangShan_Statement) F_QuanShunZi(seatIndex int) *st.FanResult {
	tp := 3020 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	FenIndex := room.Seats[seatIndex].Direct // 方位
	QuanFen := room.FengQuan                 // 圈风
	shunzi := 0

	if room.Seats[seatIndex].Cards.Peng.Count > 0 || room.Seats[seatIndex].Cards.Kong.Count > 0 {
		return &st.FanResult{tp, "", 0, SpecialSid}
	}

	mjs := make([]int, 42)
	for i := 0; i < room.Seats[seatIndex].Cards.List.Count; i++ {
		if *room.Seats[seatIndex].Cards.List.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.List.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Chow.Count; i++ {
		if *room.Seats[seatIndex].Cards.Chow.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Chow.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}
	if seatIndex != room.CurIndex {
		mjs[room.CurCard.ID]++
	}

	//风或中白发判断是否全顺字
	for k, v := range mjs {
		//如果有风或箭刻字直接返回
		if k >= 27 && k <= 33 {
			if v >= 3 {
				return &st.FanResult{tp, "", 0, SpecialSid}
			}
		}
		//风位上的对子或中发白的对子不能胡
		if k == 31 || k == 32 || k == 33 || k == FenIndex+27 || k == QuanFen+27 {
			if v != 0 {
				return &st.FanResult{tp, "", 0, SpecialSid}
			}
		}
	}
	//去掉将对判断是否为顺子
	for k1, v1 := range mjs {
		if v1 >= 2 {
			mjs[k1] -= 2
			shunzi = sc.BaseCtl.CheckQuanShunZi(mjs)
			if shunzi == 1 {
				break
			} else {
				mjs[k1] += 2
			}
		}
	}
	if shunzi > 0 {
		msg = "全顺子"
		tai = 1
	}

	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【清一色】12台，胡牌的牌全由一色牌组成；
func (sc *XiangShan_Statement) F_QinYiSe(seatIndex int) *st.FanResult {
	tp := 3021             //
	SpecialSid := -1       //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	mjs := make([]int, 42) // 麻将
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	for i := 0; i < room.Seats[seatIndex].Cards.List.Count; i++ {
		if *room.Seats[seatIndex].Cards.List.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.List.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Peng.Count; i++ {
		if *room.Seats[seatIndex].Cards.Peng.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Peng.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Kong.Count; i++ {
		if *room.Seats[seatIndex].Cards.Kong.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Kong.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Chow.Count; i++ {
		if *room.Seats[seatIndex].Cards.Chow.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Chow.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}
	if seatIndex != room.CurIndex {
		mjs[room.CurCard.ID]++
	}

	msg, tai := sc.BaseCtl.F_QingYiSe(mjs)
	if tai > 0 {
		tai = 12
		//		if len(room.Rules.Rules) >= 4 {
		//			switch room.Rules.Rules[3] {
		//			case 128:
		//				tai = 12
		//				break
		//			case 106:
		//				tai = 10
		//				break
		//			default:
		//				tai = 12
		//				break
		//			}
		//		} else {
		//			tai = 12
		//		}
		msg = "清一色"
		if sc.BaseCtl.FanType == Qin10Hun6 {
			tai = 10
		}
	}

	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【风一色】40台，胡牌的牌全由风牌构成，西周麻将的风牌包括 东南西北中发白
func (sc *XiangShan_Statement) F_FengYiSe(seatIndex int) *st.FanResult {
	tp := 3022 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	mjs := make([]int, 6)
	for i := 0; i < room.Seats[seatIndex].Cards.List.Count; i++ {
		if *room.Seats[seatIndex].Cards.List.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.List.Index(i)).(*rb.Card)
			mjs[card.Type] = 1
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Chow.Count; i++ {
		if *room.Seats[seatIndex].Cards.Chow.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Chow.Index(i)).(*rb.Card)
			mjs[card.Type] = 1
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Peng.Count; i++ {
		if *room.Seats[seatIndex].Cards.Peng.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Peng.Index(i)).(*rb.Card)
			mjs[card.Type] = 1
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Kong.Count; i++ {
		if *room.Seats[seatIndex].Cards.Kong.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Kong.Index(i)).(*rb.Card)
			mjs[card.Type] = 1
		}
	}
	if seatIndex != room.CurIndex {
		mjs[room.CurCard.Type] = 1
	}

	if mjs[0] == 1 || mjs[1] == 1 || mjs[2] == 1 {
	} else {
		msg = "风一色"
		tai = 40
		if sc.BaseCtl.FanType == Qin10Hun6 {
			tai = 20
		}
	}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【板高（2顺子）】6台，2副一模一样的顺子，如1条2条3条1条2条3条
// 【太板高（3顺子）】25台，3副一模一样的顺子，如1条2条3条1条2条3条1条2条3条
// 【太太板高（4顺子）】50台，4副一模一样的顺子，如1条2条3条1条2条3条1条2条3条1条2条3条
// 【双板高（2+2顺子）】15台，2副一模一样的顺子+另外2副一模一样的顺子，如1条2条3条1条2条3条 + 2万3万4万2万3万4万
func (sc *XiangShan_Statement) F_BanGao(seatIndex int) *st.FanResult {
	tp := 3023 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【大四喜】50台，随意一对将+东风刻+南风刻+西风刻+北风刻。
func (sc *XiangShan_Statement) F_DaSiXi(seatIndex int) *st.FanResult {
	tp := 3024 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	if room.Seats[seatIndex].Cards.Chow.Count > 0 { // 有吃则直接返回
		return &st.FanResult{tp, "", 0, SpecialSid}
	}

	mjs := new([42]int) // 麻将
	for i := 0; i < room.Seats[seatIndex].Cards.List.Count; i++ {
		if *room.Seats[seatIndex].Cards.List.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.List.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Peng.Count; i++ {
		if *room.Seats[seatIndex].Cards.Peng.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Peng.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}
	for i := 0; i < room.Seats[seatIndex].Cards.Kong.Count; i++ {
		if *room.Seats[seatIndex].Cards.Kong.Index(i) != nil {
			card := (*room.Seats[seatIndex].Cards.Kong.Index(i)).(*rb.Card)
			mjs[card.ID]++
		}
	}
	if seatIndex != room.CurIndex {
		mjs[room.CurCard.ID]++
	}
	if mjs[27] >= 3 && mjs[28] >= 3 && mjs[29] >= 3 && mjs[30] >= 3 {
		msg = "大四喜"
		tai = 50
		if sc.BaseCtl.FanType == Qin10Hun6 {
			tai = 20
		}
	}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 【还杠胡】(杠后打的牌, 别人胡的这张牌) (没有台, 如果没有台则不能胡)
func (sc *XiangShan_Statement) F_HuanGangHu(seatIndex int) *st.FanResult {
	tp := 3025 //
	msg := ""
	tai := 0
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
	huanGangHuToolFlag := 2
	if sc.GetIsDoHu() {
		huanGangHuToolFlag = 3
	}
	room := (*sc.BaseCtl.RoomRef).(*RoomXiangshan)
	if room.CurIndex != seatIndex {
		for i := huanGangHuToolFlag; i < huanGangHuToolFlag+8; i++ {
			or := sc.Get(i)
			if or != nil && or.Tool.Index == room.CurIndex {
				if or.Tool.ToolType == st.T_MoBack {
					continue
				} else if or.Tool.ToolType == st.T_Kong || or.Tool.ToolType == st.T_AnKong || or.Tool.ToolType == st.T_PengKong {
					if or.Tool.TIndex == seatIndex {
						SpecialSid = room.CurIndex
						msg = "还杠胡"
						tai = 1
					}
				}

			} else {
				break
			}
		}
	}
	return &st.FanResult{tp, msg, tai, SpecialSid}
}

// 重复关系计算
func (sc *XiangShan_Statement) F_RepeatCalc(fanlist al.ArrayList) al.ArrayList {

	fanMap := make(map[int]int)
	resultMap := make(map[int]*st.FanResult)
	for i := 0; i < fanlist.Count; i++ {
		if *fanlist.Index(i) != nil {
			fanresult := (*fanlist.Index(i)).(*st.FanResult)
			t := fanresult.Type
			if fanresult.Tai > 0 {
				fanMap[t]++
				resultMap[t] = fanresult
			}
		}
	}

	//天胡不与地胡共存
	if fanMap[3004] > 0 {
		fanMap[3005] = 0
	}

	//天胡不与自摸共存
	if fanMap[3004] > 0 || fanMap[3007] > 0 {
		fanMap[3015] = 0
	}
	//   边，嵌，单钓(吊)，对倒 全顺字 只能存在一个
	isTrue := false
	for _, v := range []int{3010, 3011, 3012, 3013, 3020} {
		if isTrue {
			fanMap[v] = 0
			continue
		}
		if fanMap[v] > 0 {
			isTrue = true
		}
	}
	// if fanMap[3010] > 0 || fanMap[3011] > 0 || fanMap[3012] > 0 || fanMap[3013] > 0 { //  边，嵌，单钓(吊)，对倒
	// 	fanMap[3020] = 0
	// 	fanMap[3012] = 0
	// }

	if fanMap[3022] > 0 { // 风一色
		fanMap[3019] = 0
	}
	if fanMap[3024] > 0 { // 大四喜
		fanMap[3019] = 0
	}

	allfan := al.New()
	for i, v := range fanMap {
		if v > 0 {
			allfan.Add(resultMap[i])
		}
	}
	return *allfan
}

func (sc *XiangShan_Statement) GetIsDoHu() bool {
	or := sc.Get(0)
	if or != nil && or.Tool != nil && (or.Tool.ToolType == st.T_Hu || or.Tool.ToolType == st.T_Hu_ZiMo) {
		return true
	}
	return false
}
