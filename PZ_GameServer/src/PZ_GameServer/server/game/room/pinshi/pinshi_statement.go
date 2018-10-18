/*
 * +-----------------------------------------------
 * | ningbo_statemnet.go
 * +-----------------------------------------------
 * | Version: 1.0
 * +-----------------------------------------------
 * | Context: 拼十结算 GameType : 3000
 * +-----------------------------------------------
 */
package pinshi

import (
	al "PZ_GameServer/common/util/arrayList"
	st "PZ_GameServer/server/game/statement"
	"strconv"
)

const (
	//数值空间 GameType + ID
	// 例子 : GameType = 3000, QuanID = 1 ,T_Quan = 3001
	T_Quan = 3001 // 圈
)

func GetPinshiStatement(mjType int32, bscore int, uid []string, room interface{}) *Pinshi_Statement {
	smt := Pinshi_Statement{}
	smt.BaseCtl.GameType = int(mjType)
	smt.BaseCtl.IDs = uid
	smt.BaseCtl.BaseScore = bscore
	smt.BaseCtl.Record = al.New()
	smt.BaseCtl.Types = make(map[int]string)
	smt.BaseCtl.RoomRef = &room

	smt.Init()
	return &smt
}

type Pinshi_Statement struct {
	BaseCtl st.StatementCtl
}

func (s *Pinshi_Statement) Init() {
	s.BaseCtl.Init()
	s.AddType(T_Quan, "定圈")
}

func (s *Pinshi_Statement) AddListCard(index int, listcard []int) {
	s.BaseCtl.AddListCard(index, listcard)
}

func (s *Pinshi_Statement) AddRecord(record st.OnceRecord) {
	s.BaseCtl.AddRecord(record)
}

func (s *Pinshi_Statement) AddTool(toolType int, index int, tindex int, val []int) {
	s.BaseCtl.AddTool(toolType, index, tindex, val)
}

func (s *Pinshi_Statement) AddType(toolType int, msg string) {
	s.BaseCtl.AddType(toolType, msg)
}

func (s *Pinshi_Statement) GetMsg(toolType int) string {
	return s.BaseCtl.GetMsg(toolType)
}

func (s *Pinshi_Statement) Get(index int) *st.OnceRecord {
	return s.BaseCtl.GetPrvTool(index)
}

func (s *Pinshi_Statement) GetForType(tooltype int) *st.OnceRecord {
	return s.BaseCtl.GetPrvToolForType(tooltype)
}

func (s *Pinshi_Statement) GetForTypes(tooltypes []int) *st.OnceRecord {
	return nil
}

func (s *Pinshi_Statement) GetTypeCount(uid int, tooltype int) int {
	return 0
}

func (s *Pinshi_Statement) ToolCalc() *st.OnceRecord {
	return nil
}

func (s *Pinshi_Statement) CalcTotal() {

}

// 判断是否胡牌
func (s *Pinshi_Statement) CheckHu(pM []int) int {
	result := s.BaseCtl.CheckHu(pM)
	//fmt.Println("判断胡牌 ", result, pM)
	return result
}

// 计算承包关系
// 返回: 承包关系的seatIndex ,  分数
func (s *Pinshi_Statement) CheckChengBao(seatIndex int) ([]int, []int) {

	room := (*s.BaseCtl.RoomRef).(*RoomPinshi)
	chengbao := make([]int, room.Rules.SeatLen)
	fen := make([]int, room.Rules.SeatLen)

	// 获得承包关系
	for i := 0; i < room.Rules.SeatLen; i++ {
		v := room.ChengBao[seatIndex].Seat[i]
		if v >= 3 && v < 6 {
			chengbao[i] = 1

			room.Seats[i].Accumulation.PayCount++
		} else if v >= 6 {
			chengbao[i] = 2
			room.Seats[i].Accumulation.PayCount++
		}
		//测试
		// if v >= 2 && v < 4 {
		// 	chengbao[i] = 1
		// } else if v >= 4 {
		// 	chengbao[i] = 2
		// }
	}

	// 判断是否自摸或点炮 A和B为承包关系
	if room.CurIndex == seatIndex {
		//自摸  如果A自摸，B支付分数给A，此时其余两家不必支付积分
		for i := 0; i < room.Rules.SeatLen; i++ {
			if chengbao[i] > 0 {
				fen[i] = 5 * chengbao[i]
			}
		}

	} else {
		// 点炮
		if chengbao[room.CurIndex] > 0 { //  B给A点炮，那么B支付积分给A
			//有承包关系的点炮
			fen[room.CurIndex] = 2 * 2 * chengbao[room.CurIndex] //总台数*2（承包）*2（点炮）
		} else {
			//C点炮给A，那么C支付积分给A，B虽然没有点炮，B也得支付与C相同的分数给A
			fen[room.CurIndex] = 2 //* chengbao[room.CurIndex] // 点炮的人
			// for i := 0; i < room.Rules.SeatLen; i++ {
			// 	if chengbao[i] > 0 {
			// 		fen[i] = 2 * chengbao[i] // 承包的人
			// 	}
			// }
		}
		// 多种承包关系 A承包B，B承包C
		for i := 0; i < room.Rules.SeatLen; i++ {
			if chengbao[i] > 0 && i != room.CurIndex {
				fen[i] = 2 * chengbao[i] // 承包的人
			}
		}
	}
	return chengbao, fen
}

// 胡牌前的算番
// 判断是否一台起胡
func (s *Pinshi_Statement) CheckCanWin(seatIndex int, args ...interface{}) bool {
	room := (*s.BaseCtl.RoomRef).(*RoomPinshi) //房间
	uc := room.Seats[seatIndex].Cards

	if uc.Hua.Count > 0 {
		return true //花牌
	}

	if uc.Kong.Count == 0 && uc.Peng.Count == 0 && uc.Chow.Count == 0 {
		return true //门清
	}

	if uc.List.Count == 1 {
		return true // 大吊车
	}

	if s.F_ZiMo(seatIndex).Tai > 0 {
		return true // 自摸
	}

	if s.F_Feng(seatIndex).Tai > 0 {
		return true // 风
	}

	if s.F_ZhongFaBai(seatIndex).Tai > 0 {
		return true // 中发白
	}

	if s.F_TianHu(seatIndex).Tai > 0 {
		return true
	}
	if s.F_DiHu(seatIndex).Tai > 0 {
		return true
	}
	if s.F_LaGangHu(seatIndex).Tai > 0 {
		return true
	}
	if s.F_SongGangHu(seatIndex).Tai > 0 {
		return true
	}
	if s.F_KangShangKaiHua(seatIndex).Tai > 0 {
		return true
	}
	if s.F_HaiDiLaoYue(seatIndex).Tai > 0 {
		return true
	}
	if s.F_BianDao(seatIndex).Tai > 0 {
		return true
	}
	if s.F_QianDao(seatIndex).Tai > 0 {
		return true
	}
	if s.F_DanDiao(seatIndex).Tai > 0 {
		return true
	}
	if s.F_DuiDao(seatIndex).Tai > 0 {
		return true
	}

	if s.F_Duan19(seatIndex).Tai > 0 {
		return true
	}
	if s.F_MenQing(seatIndex).Tai > 0 {
		return true
	}
	if s.F_DaDiaoChe(seatIndex).Tai > 0 {
		return true
	}
	if s.F_HunYiSe(seatIndex).Tai > 0 {
		return true
	}
	if s.F_DuiDuiHu(seatIndex).Tai > 0 {
		return true
	}
	if s.F_QuanShunZi(seatIndex).Tai > 0 {
		return true
	}
	if s.F_QinYiSe(seatIndex).Tai > 0 {
		return true
	}
	if s.F_FengYiSe(seatIndex).Tai > 0 {
		return true
	}
	if s.F_BanGao(seatIndex).Tai > 0 {
		return true
	}
	if s.F_DaSiXi(seatIndex).Tai > 0 {
		return true
	}

	return false

}

// 算番 (自定义判断)
func (s *Pinshi_Statement) FanCalc(seatIndex int, args ...interface{}) st.TotalResult {

	//fmt.Println(" 开始算番 >")
	SpecialSid := -1 //用于特别的算番记录的uid，例如 拉杠胡，送杠胡，还杠胡，其它用不到-1
	attached := ""   //用于拉杠胡，送杠胡，还杠胡附加json信息
	gangIndex := -1
	gangCid := -1
	room := (*s.BaseCtl.RoomRef).(*RoomPinshi) //房间

	if room.Rules.BaseScore < 1 { // 底分最低为1分
		room.Rules.BaseScore = 1
	}

	var totalresult st.TotalResult
	totalresult.TotalMsg = make([]string, room.Rules.SeatLen)

	fanList := al.New()
	cb, cbfeng := s.CheckChengBao(seatIndex) //承包关系

	zm := s.F_ZiMo(seatIndex)
	fanList.Add(zm)                         // 自摸
	fanList.Add(s.F_Hua(seatIndex))         // 花
	fanList.Add(s.F_Feng(seatIndex))        // 风
	fanList.Add(s.F_ZhongFaBai(seatIndex))  // 中发白
	fanList.Add(s.F_TianHu(seatIndex))      // 天胡
	fanList.Add(s.F_DiHu(seatIndex))        // 地胡
	fanList.Add(s.F_HaiDiLaoYue(seatIndex)) // 海底捞月
	fanList.Add(s.F_BianDao(seatIndex))     // 边倒
	fanList.Add(s.F_QianDao(seatIndex))     // 嵌到
	fanList.Add(s.F_DanDiao(seatIndex))     // 单吊
	fanList.Add(s.F_DaDiaoChe(seatIndex))   // 大吊车
	fanList.Add(s.F_DuiDao(seatIndex))      // 对到
	fanList.Add(s.F_Duan19(seatIndex))      // 段19
	fanList.Add(s.F_HunYiSe(seatIndex))     // 混一色
	fanList.Add(s.F_DuiDuiHu(seatIndex))    // 对对胡
	qsz := s.F_QuanShunZi(seatIndex)
	fanList.Add(qsz)                            // 全顺子
	fanList.Add(s.F_QinYiSe(seatIndex))         // 清一色
	fanList.Add(s.F_FengYiSe(seatIndex))        // 风一色
	fanList.Add(s.F_BanGao(seatIndex))          // 板高
	fanList.Add(s.F_DaSiXi(seatIndex))          // 大四喜
	fanList.Add(s.F_MenQing(seatIndex))         // 门清
	fanList.Add(s.F_LaGangHu(seatIndex))        // 拉杠胡
	fanList.Add(s.F_SongGangHu(seatIndex))      // 松杠胡
	fanList.Add(s.F_KangShangKaiHua(seatIndex)) // 杠上开花
	fanList.Add(s.F_HuanGangHu(seatIndex))      // 还杠胡

	msg1 := ""
	for i := 0; i < fanList.Count; i++ {
		if *fanList.Index(i) != nil {
			f := (*fanList.Index(i)).(*st.FanResult)
			msg1 += f.Msg
		}

	}
	fanList2 := s.F_RepeatCalc(*fanList) // 重复牌型计算

	msg2 := ""
	for i := 0; i < fanList2.Count; i++ {
		if *fanList2.Index(i) != nil {
			f := (*fanList2.Index(i)).(*st.FanResult)
			msg2 += f.Msg
			//		if (f.Type == 3007 || f.Type == 3006 || f.Type == 3025) && f.Tai > 0 {
			//			SpecialSid = f.SpecialSid
			//		}
		}

	}
	msg := ""
	tai := 0
	for i := 0; i < fanList2.Count; i++ {
		if *fanList2.Index(i) != nil {
			f := (*fanList2.Index(i)).(*st.FanResult)
			if f.Msg != "" {
				if f.Type == 3001 {
					msg += f.Msg
				} else {
					msg += (f.Msg + "(" + strconv.Itoa(f.Tai) + "台) ")
				}
				tai += f.Tai
				if f.Type == 3006 {
					SpecialSid, gangIndex, gangCid = s.GetLaGangHuInfo(seatIndex)
				} else if f.Type == 3007 {
					SpecialSid, gangIndex, gangCid = s.GetSongGangHuInfo(seatIndex)
				} else if f.Type == 3025 {
					SpecialSid, gangIndex, gangCid = s.GetHuanGangHuInfo(seatIndex)
				}
			}
		}

	}

	if SpecialSid >= 0 {
		attached = "{\"index\":" + strconv.Itoa(gangIndex) + ",\"cid\":" + strconv.Itoa(gangCid) + "}"
	}

	peifu := make([]int32, room.Rules.SeatLen) //赔付

	ischengbao := false
	strchengbao := ""
	chengbaofeng := make([]int32, room.Rules.SeatLen)
	for i := 0; i < len(cbfeng); i++ {
		if cb[i] > 0 { // 有承包关系
			ischengbao = true
			strchengbao += room.Seats[i].User.NickName + " "
		}
		//拉杠胡，送杠胡，还杠胡  构成承包且对关系人按照自摸算番
		if SpecialSid == i {
			ischengbao = true
			if cb[SpecialSid] == 1 || cb[SpecialSid] == 2 {
				chengbaofeng[SpecialSid] = int32(10 * tai * -room.Rules.BaseScore)
			} else {
				chengbaofeng[SpecialSid] = int32(5 * tai * -room.Rules.BaseScore)
				strchengbao += room.Seats[SpecialSid].User.NickName + " "
			}
		} else {
			if SpecialSid >= 0 {
				if cb[i] == 1 {
					chengbaofeng[i] = int32(5 * tai * -room.Rules.BaseScore)
				} else if cb[i] == 2 {
					chengbaofeng[i] = int32(10 * tai * -room.Rules.BaseScore)
				}
			} else {
				chengbaofeng[i] = int32(cbfeng[i] * tai * -room.Rules.BaseScore)
			}

		}
	}

	if !ischengbao { // 没有承包关系
		if zm.Tai > 0 { // 自摸
			for i := 0; i < room.Rules.SeatLen; i++ {
				if i != seatIndex {
					chengbaofeng[i] = int32(tai * -room.Rules.BaseScore)
				}
			}
		} else { // 点炮
			chengbaofeng[room.CurIndex] = int32(tai*-room.Rules.BaseScore) * 2
		}
	} else {
		msg += "承包(" + strchengbao + ")"
	}

	totalfen := 0
	for i := 0; i < len(chengbaofeng); i++ {
		totalfen += int(chengbaofeng[i])
	}

	chengbaofeng[seatIndex] = int32(-totalfen)
	totalresult.TotalScore = chengbaofeng
	totalresult.PeifuCount = peifu
	totalresult.TotalMsg[seatIndex] = msg
	totalresult.TotalTai = int32(tai)
	totalresult.Winner = int32(seatIndex)
	totalresult.Attached = attached //附加信息 暂时只有拉，送 还杠胡使用
	return totalresult
}

//获取拉杠胡的杠牌以及杠牌座位
func (s *Pinshi_Statement) GetLaGangHuInfo(seatIndex int) (int, int, int) {
	specialSid := -1
	gangIndex := -1
	gangCid := -1
	room := (*s.BaseCtl.RoomRef).(*RoomPinshi)
	if room.CurIndex != seatIndex {
		or := s.Get(1)
		if or != nil && or.Tool != nil && or.Tool.ToolType == st.T_PengKong {
			specialSid = room.CurIndex
			gangCid = or.Tool.Val[0]
			gangIndex = or.Tool.Index

		}
	}
	return specialSid, gangIndex, gangCid
}

//获取送杠胡的杠牌以及杠牌座位
func (s *Pinshi_Statement) GetSongGangHuInfo(seatIndex int) (int, int, int) {
	specialSid := -1
	gangIndex := -1
	gangCid := -1
	room := (*s.BaseCtl.RoomRef).(*RoomPinshi)
	if room.CurIndex == seatIndex {
		//自摸
		for i := 1; i < 9; i++ {
			or1 := s.Get(i)
			if or1 != nil && or1.Tool.Index == room.CurIndex {
				if or1.Tool.ToolType == st.T_MoBack {
					continue
				} else if (or1.Tool.ToolType == st.T_Kong || or1.Tool.ToolType == st.T_PengKong) && or1.Tool.TIndex != room.CurIndex {

					specialSid = or1.Tool.TIndex
					gangCid = or1.Tool.Val[0]
					gangIndex = or1.Tool.Index
					break
				}
			} else {
				break
			}
		}
	}
	return specialSid, gangIndex, gangCid
}

////获取还杠胡的杠牌以及杠牌座位
func (s *Pinshi_Statement) GetHuanGangHuInfo(seatIndex int) (int, int, int) {
	specialSid := -1
	gangIndex := -1
	gangCid := -1
	room := (*s.BaseCtl.RoomRef).(*RoomPinshi)
	if room.CurIndex != seatIndex {
		for i := 3; i < 12; i++ {
			or := s.Get(i)
			if or != nil && or.Tool.Index == room.CurIndex {
				if or.Tool.ToolType == st.T_MoBack {
					continue
				} else if or.Tool.ToolType == st.T_Kong || or.Tool.ToolType == st.T_AnKong || or.Tool.ToolType == st.T_PengKong {
					if or.Tool.TIndex == seatIndex {
						specialSid = room.CurIndex
						gangCid = or.Tool.Val[0]
						gangIndex = or.Tool.Index
						break
					}
				}

			} else {
				break
			}
		}
	}
	return specialSid, gangIndex, gangCid
}
