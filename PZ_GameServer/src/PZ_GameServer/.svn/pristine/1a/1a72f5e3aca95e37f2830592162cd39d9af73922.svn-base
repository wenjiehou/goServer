// 血流麻将  GameType : MsgID : 2000
package room

import (
	"PZ_GameServer/protocol/pb"
	rb "PZ_GameServer/server/game/roombase"
	//"PZ_GameServer/server/router"
	"PZ_GameServer/server/user"
)

const (
	XueLiu = int32(mjgame.MsgID_GTYPE_SiChuan_XueLiu)
)

// 规则
var XueLiu_RoomRule = rb.RoomRule{

	Create_NeedDiamond: 0,         // 创建房间需要的钻石
	Play_NeedDiamond:   100,       // 开始玩 需要的钻石(总数)()
	SeatLen:            4,         // 座位数量  2, 3, 4
	DefCardLen:         13,        // 默认手牌数量 13
	AllCardLen:         108,       //
	Card_W:             1,         // 默认带万 1代表一次
	Card_B:             1,         // 默认带饼
	Card_T:             1,         // 默认带条
	Card_F:             0,         // 默认带风
	Card_J:             0,         // 默认带箭 (中发白)
	Card_H:             0,         // 默认带花
	Card_Else:          []int{},   // 特殊牌
	CanLaiZi:           0,         // 赖子数量
	CanPeng:            1,         // 可以碰
	CanChow:            0,         // 可以吃
	CanKong:            1,         // 可以直杠
	CanAnKong:          1,         // 可以暗杠
	CanPengKong:        1,         // 可以碰杠
	CanTing:            1,         // 可以听
	CanWin:             1,         // 可以直胡
	CanZiMo:            1,         // 可以自摸胡
	MaxWinCount:        0,         // 最大胡牌数量<0为不限次数, 大众麻将为1
	MaxTime:            0,         // 最大时间(<=0 为不限时间)
	MaxRound:           0,         // 最大局数(<=0 为不限局数)
	BaseScore:          1,         // 基本分
	MaxTai:             0,         // 封顶台数(<=0 为不限)
	WinNeedTai:         0,         // 胡牌最小台数 (<=0 为不限)
	Rules:              []int32{}, // 全部规则(包含特殊规则)
}

type RoomXueLiu struct {
	rb.RoomBase
}

func (r *RoomXueLiu) Create(rid int, t int32, user *user.User, rule *rb.RoomRule) {
	r.RoomBase.Create(rid, t, user.ID, rule)
}

// 绑定路由
func XueLiu_Init() {
	//router.Bind(int32(mjgame.MsgID_MSG_Chow), XueLiu, &mjgame.ACKBC_Chow{}, XueLiu_Chow) //玩家准备

}

func XueLiu_Chow(args ...interface{}) {

}
