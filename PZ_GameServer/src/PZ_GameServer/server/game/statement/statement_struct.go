package statement

import (
	al "PZ_GameServer/common/util/arrayList"
)

//结算基类接口
type IStatement interface {
	Init()                                                  // 初始化
	AddRecord(record OnceRecord)                            // 添加记录
	AddTool(toolType int, index int, tindex int, val []int) // 添加操作
	AddListCard(index int, listcard []int)                  // 添加初始手牌(13张)
	AddType(toolType int, msg string)                       // 添加类型
	GetMsg(toolType int) string                             //
	Get(index int) *OnceRecord                              // 得到上一步操作
	GetForType(tooltype int) *OnceRecord                    // 得到上一步操作(Type)
	GetForTypes(tooltypes []int) *OnceRecord                //
	GetTypeCount(uid int, tooltype int) int                 // 得到操作类型
	ToolCalc() *OnceRecord                                  // 操作结算
	CheckHu(pM []int) int                                   // 判断是否胡牌
	FanCalc(seatIndex int, args ...interface{}) TotalResult // 算番
	CalcTotal()                                             // 计算统计结果
}

// 算番 预处理结构,用于节省计算量
type PreFanInit struct {
	all_mjs    []int // 全部麻将(包括吃,碰,杠)
	types      []int // 全部类型
	list       []int
	chow       []int
	peng       []int
	kong       []int
	hua        []int
	hu         []int
	startIndex int
}

// 结算控制器
type StatementCtl struct {
	GameType  int            // 麻将类型
	BaseScore int            // 基础分
	IDs       []string       // User ID
	Names     []string       //
	Score     []int          // 分数(开始都是0)
	CurHu     []int          // 当前胡家的牌型, 包括被胡的牌. 胡牌后, 填充此数组, 用来算番
	Record    *al.ArrayList  // 记录
	RoomRef   *interface{}   // 房间指针
	Types     map[int]string // 类型
	FanType   int            // 算番类型(有的麻将一个番型,有多种番数(用户可以选择))
	Count     int            // RecordCount
	SiChuan   bool           //
}

// 单次记录 (操作和结果分开, 在逻辑层可以分开处理)
type OnceRecord struct {
	Index  int
	Tool   *OnceTool   // 操作
	Result *OnceResult // 结果
}

// 算番结果
type FanResult struct {
	Type       int
	Msg        string
	Tai        int
	SpecialSid int //用于特别的算番记录的sid，例如 拉杠胡，送杠胡，还杠胡，其它用不到的传-1
}

// 一次操作记录
type OnceTool struct {
	Index    int      // 操作者(发起者)
	TIndex   int      // 目标索引
	ToolType int      // 操作类型 (碰 杠  碰后杠 胡牌)
	Val      []int    // 值(换3张, 定缺...)
	MSG      []string // 自定义描述消息
}

// 一次结果(统计结果)
type OnceResult struct {
	Type        []int    // 结算类型 (对对胡, 刮风, 下雨...)
	Score       []int    // 获得的分数  (按座位顺序)
	Val         []int    // 值
	MSG         []string // 自定义描述消息
	UpdateScore bool     // 是否即时更新分数
}

// 统计结果
type TotalResult struct {
	TotalScore  []int32      // 总计分数
	TotalMsg    []string     //
	PeifuCount  []int32      // 赔付
	TotalResult []OnceResult // 统计结果
	Winner      int32        // 赢家
	TotalTai    int32        // 台数统计
	Attached    string       //附加信息
}
