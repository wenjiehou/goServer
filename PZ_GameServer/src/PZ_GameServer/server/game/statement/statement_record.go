package statement

import (
	"PZ_GameServer/protocol/pb"
)

//麻将牌型的定义
const (
	W = 0 // 万字牌 1-9   36
	B = 1 // 饼字牌 1-9   36
	T = 2 // 条字牌 1-9   36
	F = 3 // 风字牌 1=东 南 西 北 (逆时针)   16
	J = 4 // 箭字牌 1=中 2=发 3=白   12
	H = 5 // 花牌 0=春 1=夏 2=秋 3=冬 4=梅 5=兰 6=竹 7=菊  8
)

//	mjs[0] = 0 // 万
//	mjs[1] = 0
//	mjs[2] = 0
//	mjs[3] = 0
//	mjs[4] = 0
//	mjs[5] = 0
//	mjs[6] = 3
//	mjs[7] = 0
//	mjs[8] = 0

//	mjs[9] = 0 // 饼
//	mjs[10] = 0
//	mjs[11] = 1
//	mjs[12] = 1
//	mjs[13] = 4
//	mjs[14] = 0
//	mjs[15] = 0
//	mjs[16] = 0
//	mjs[17] = 0

//	mjs[18] = 0 // 条
//	mjs[19] = 0
//	mjs[20] = 0
//	mjs[21] = 0
//	mjs[22] = 0
//	mjs[23] = 0
//	mjs[24] = 0
//	mjs[25] = 0
//	mjs[26] = 0

//	mjs[27] = 0  // 东   风
//	mjs[28] = 0  // 南
//	mjs[29] = 0  // 西
//	mjs[30] = 0  // 北

//	mjs[31] = 0  // 中   箭
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

// 总计 144 张牌

// 得到上一个Tool记录
func (s *StatementCtl) GetPrvTool(index int) *OnceRecord {
	pIndex := s.Record.Count - 1 - index
	if pIndex < 0 || s.Record == nil || s.Record.Count == 0 {
		return nil
	}
	if *s.Record.Index(pIndex) != nil {
		return (*s.Record.Index(pIndex)).(*OnceRecord)
	} else {
		return nil
	}
}

// 得到当前和上一个Tool
func (s *StatementCtl) GetPrvToolForType(getType int) *OnceRecord {

	var tool1 *OnceRecord = nil
	var preTool *OnceRecord
	for i := s.Record.Count - 1; i >= 0; i-- {
		if *s.Record.Index(i) != nil {
			preTool = (*s.Record.Index(i)).(*OnceRecord)
			if preTool.Tool == nil {
				break
			}
			toolType := preTool.Tool.ToolType
			switch getType {
			case int(mjgame.MsgID_MSG_Kong):
				if toolType == T_Kong || toolType == T_PengKong || toolType == T_AnKong {
					tool1 = preTool
					break
				}
			case int(mjgame.MsgID_MSG_ACKBC_GetCard):
				if toolType == T_Mo || toolType == T_MoBack {
					tool1 = preTool
					break
				}
			case T_Put:
				if toolType == T_Put {
					tool1 = preTool
					break
				}
			case T_Peng:
				if toolType == T_Peng {
					tool1 = preTool
					break
				}
			case T_Chow:
				if toolType == T_Chow {
					tool1 = preTool
					break
				}
			default:
				if toolType == getType {
					tool1 = preTool
					break
				}
			}
		}

	}
	return tool1
}
