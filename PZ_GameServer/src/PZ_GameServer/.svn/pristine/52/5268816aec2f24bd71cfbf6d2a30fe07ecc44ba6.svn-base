package statement

import (
	"fmt"
)

func init() {
	fmt.Println("statement init")
}

//// 这里定义的变量可以在派生类中增加 后期都是读配置表
//type Def struct {
//	SitDown  int // 坐下
//	StandUp  int // 站起来
//	Ready    int // 准备
//	Start    int // 开始
//	Dice     int // 筛子
//	Zhuang   int // 定庄家
//	Direct   int // 定方位
//	Deal     int // 发牌
//	Mo       int // 正常摸牌
//	MoBack   int // 后面摸牌
//	Put      int // 出牌
//	Kong     int // 杠
//	AnKong   int // 暗杠
//	PengKong int // 补杠 // 碰过再杠
//	Peng     int // 碰
//	Chow     int // 吃
//	Ting     int // 听
//	Hu       int // 胡
//	Hu_ZiMo  int // 胡 自摸
//	Exit     int // 用户退出
//	Enter    int // 用户进入
//}

const (
	T_SitDown  = 5   // 坐下
	T_StandUp  = 6   // 站起来
	T_Ready    = 7   // 准备
	T_Start    = 8   // 开始
	T_Dice     = 10  // 筛子
	T_Zhuang   = 15  // 定庄家
	T_Direct   = 16  // 定方位
	T_Deal     = 20  // 发牌
	T_Mo       = 30  // 正常摸牌
	T_MoBack   = 40  // 后面摸牌
	T_Put      = 50  // 出牌
	T_Kong     = 60  // 杠
	T_AnKong   = 70  // 暗杠
	T_PengKong = 80  // 补杠 // 碰过再杠
	T_Peng     = 90  // 碰
	T_Chow     = 100 // 吃
	T_Ting     = 120 // 听
	T_Hu       = 130 // 胡
	T_Hu_ZiMo  = 150 // 胡 自摸
	T_Exit     = 300 // 用户退出
	T_Enter    = 310 // 用户进入
	T_Draw     = 320 // 流局

	T_HuaZhu        = 160  // 花猪
	T_Dajiao        = 165  // 大叫
	T_Change3Card   = 180  // 换3张
	T_Change3CardOK = 185  // 换3张完成
	T_MissType      = 190  // 定缺
	T_DuiDuiHu      = 1010 // 对对胡(都是刻)
	T_QinYiSe       = 1020 // 清一色
	T_Dai19         = 1030 // 带幺玖
	T_QiDui         = 1040 // 七对
	T_JinGouGou     = 1050 // 金钩钩
	T_KongHa        = 1060 // 杠上花
	T_KongPao       = 1070 // 杠上炮
	T_LongQiDui     = 1080 // 龙七对
	T_QinLongQiDui  = 1090 // 清龙七对
	T_QinQiDui      = 1100 // 清七对
	T_Qin19         = 1110 // 清19
	T_QinJinGouGou  = 1120 // 清金钩钩
	T_ShiBaLuoHan   = 1130 // 18罗汉
	T_PingHu        = 1150 // 平胡

)

//// 获得操作描述
//func GetToolDef(id int) ToolDef {
//	var td = ToolDef{ID: id, MSG: ""}
//	if val, ok := DictDef[id]; ok {
//		td.MSG = val.MSG
//	}
//	return td
//}

// 用于显示日志
var (
	DictMsg = map[int]string{
	//		TEnum_SitDown:  "坐下",
	//		TEnum_StandUp:  "站起",
	//		TEnum_Dice:     "筛子",
	//		TEnum_Zhuang:   "定庄家",
	//		TEnum_Deal:     "发牌",
	//		TEnum_Mo:       "摸牌",
	//		TEnum_MoBack:   "杠后抓牌",
	//		TEnum_Put:      "出牌",
	//		TEnum_Kong:     "杠",
	//		TEnum_AnKong:   "暗杠",
	//		TEnum_PengKong: "补杠",
	//		TEnum_Peng:     "碰",
	//		TEnum_Chow:     "吃",
	//		TEnum_Ting:     "听",
	//		TEnum_Hu:       "胡",
	//		TEnum_Hu_ZiMo:  "自摸",

	//		TEnum_Change3Card:   "换3张",
	//		TEnum_Change3CardOK: "换3张完成",
	//		TEnum_MissType:      "定缺",
	//		TEnum_DuiDuiHu:      "对对胡(刻)",
	//		TEnum_QinYiSe:       "清一色",
	//		TEnum_Dai19:         "带幺玖",
	//		TEnum_QiDui:         "七对",
	//		TEnum_JinGouGou:     "金钩钩",
	//		TEnum_KongHa:        "杠上花",
	//		TEnum_KongPao:       "杠上炮",
	//		TEnum_LongQiDui:     "龙七对",
	//		TEnum_QinLongQiDui:  "清龙七对",
	//		TEnum_QinQiDui:      "清七对",
	//		TEnum_Qin19:         "清19",
	//		TEnum_QinJinGouGou:  "清金钩钩",
	//		TEnum_ShiBaLuoHan:   "18罗汉",
	//		TEnum_PingHu:        "平胡",
	}
)
