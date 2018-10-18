package init

import (
	al "PZ_GameServer/common/util/arrayList"
	xz "PZ_GameServer/server/game/room/xizhou"
	rb "PZ_GameServer/server/game/roombase"
	st "PZ_GameServer/server/game/statement"
	"encoding/json"
	"io/ioutil"
	"log"
)

type InitData struct {
	CurIndex   int   //当前轮到的玩家
	SiteIndex  int   //座位
	Direct     int   //方位
	CurCard    int   //当前出的牌
	Feng       int   //圈风
	List       []int //手牌
	Chow       []int //吃牌
	Peng       []int //碰牌
	Kong       []int //杠牌
	KongType   []int //杠牌类型0=明杠 1=暗杠 2=碰杠
	Hua        []int //花牌
	Hu         int   //胡牌
	Ting       int   //是否听牌
	ChengBao   []int //承包
	StartIndex int   //开始的牌的位置
	CurMJIndex int   //当前拿牌的位置 (断线重连)
	EndBlank   int   //结尾拿掉的牌 (杠后从结尾拿掉的牌)
}

func GetHuId() int {
	return Config.Hu
}

func GetSiteIndex() int {
	return Config.SiteIndex
}
func GetCard(id int) *rb.Card {

	//  ID     int    // id
	//  Type   int    // 类型 w=0 b=1 t=2
	//  Num    int    // 字数
	//  TIndex int    // 碰吃家的座位Index
	//  KType  int    // 0=明杠 1=暗杠 2=碰杠 状态类型
	//  MSG    string // 说明

	c := rb.Card{ID: id}
	c.Type, c.Num = st.GetMjTypeNum(id)
	c.MSG = st.GetMjNameForIndex(id)
	return &c

}

var Config InitData

func init() {
	data := ReadFile()
	Config = ConfigParse(data)
	// log.Printf("config data is :%d\n", Config.CurIndex)

}

func ReadFile() []byte {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
		return nil
	}
	// log.Printf("Data read: %s\n", data)

	return data
}

func ConfigParse(data []byte) InitData {
	m := InitData{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		log.Printf("err: %s\n", err)

	}
	return m
}

func InitRoom(index int) xz.RoomXiZhou {
	r := xz.RoomXiZhou{}
	// r.SeatLen = 4
	r.Rules.SeatLen = 4
	r.Init()
	r.StartIndex = Config.StartIndex
	r.CurMJIndex = Config.CurMJIndex
	r.EndBlank = Config.EndBlank
	r.CurIndex = Config.CurIndex
	r.AllCardLength = 144
	r.CurCard = GetCard(Config.Hu)
	r.Seats = make([]*rb.SeatBase, 4)

	r.ChengBao[index].Seat = Config.ChengBao
	r.ChengBao[index].Index = index

	UserCard := InitCards()
	seatBase := rb.SeatBase{}
	seatBase.Cards = &UserCard
	seatBase.Ting = Config.Ting
	seatBase.Direct = Config.Direct
	r.Seats[index] = &seatBase

	return r

}

func InitCards() rb.UserCard {
	userCard := rb.UserCard{
		List: al.New(),
		Kong: al.New(),
		Peng: al.New(),
		Chow: al.New(),
		Out:  al.New(),
		Hua:  al.New(),
		Hu:   al.New(),
	}
	for _, v := range Config.List {
		c := GetCard(v)
		userCard.List.Add(c)
	}
	for k, v := range Config.Kong {
		c := GetCard(v)
		c.Status = Config.KongType[k]
		userCard.Kong.Add(c)
	}
	for _, v := range Config.Peng {
		c := GetCard(v)
		userCard.Peng.Add(c)
	}
	for _, v := range Config.Chow {
		c := GetCard(v)
		userCard.Chow.Add(c)
	}
	for _, v := range Config.Hua {
		c := GetCard(v)
		userCard.Hua.Add(c)
	}

	userCard.Hu.Add(GetCard(Config.Hu))

	return userCard
}

// type RoomXiZhou struct {
// 	rb.RoomBase

// 	ChengBao []ChengBaoSeat

// 	FengQuan int // 风圈(0-3)东南西北

// 	Status int
// }

// 房间基础信息
// type RoomBase struct {
// 	IRoom

// 	RID             int             // 房间号
// 	PWD             string          // 房间密码
// 	CreateUID       string          // 创建房间的用户 (房主)(根据PayType, 如果一人付,退出房间后依然是房主付)
// 	CreateTime      int64           // 房间创建时间
// 	Type            int32           // 类型(玩法类型) 类型详见 Proto文件
// 	Level           int             // 等级
// 	State           chan int        // 状态机
// 	Rules           RoomRule        // 规则列表
// 	chanRoom        chan int        // 操作通道
// 	RoundCount      int             // 局数
// 	BigRoundCount   int             // 大圈数
// 	StartTime       int64           // 游戏开始时间
// 	RoundTime       int64           // 当前局开始时间
// 	TotalTime       int64           // 总计时间(从开局开始)
// 	IsRun           bool            // 是否游戏进行中
// 	Dict            int             // 骰子
// 	ZhuangIndex     int             // 庄家的位置
// 	CurIndex        int             // 当前轮到的玩家
// 	CurToolIndex    int             // 当前等待操作玩家的Index
// 	SeatLen         int             // 房间内座位数量 2,3,4
// 	Seats           []*SeatBase     // 房间内座位
// 	WatchSeats      []*SeatBase     // 观察者信息
// 	AllCards        []Card          // 全部牌
// 	Ticker          *time.Ticker    // 计时器
// 	TimeOutCB       reflect.Value   // 超时回调
// 	TimeOutCBParam  []reflect.Value // 超时回调参数
// 	NeedWaitTool    *al.ArrayList   // 等待操作的过程 0=胡牌 1=杠 2=碰 3=吃
// 	WaitPutTimeOut  int             // 出牌超时时间
// 	WaitToolTimeOut int             // 操作超时时间
// 	WaitTimeCount   int             // 等待计时时间
// 	CurCard         *Card           // 当前出的牌(用于吃碰杠胡)
// 	AllCardLength   int             // 全部牌的数量
// 	CurMJIndex      int             // 当前拿牌的位置 (断线重连)
// 	StartIndex      int             // 开始的牌的位置 (可以算出总共拿了多少张)
// 	EndBlank        int             // 结尾拿掉的牌 (杠后从结尾拿掉的牌)
// 	StlCtrl         interface{}     // 记录结算
// }
//座位基础信息
// type SeatBase struct {
// 	UID             string     // 用户uid
// 	Index           int        // (0-3) 座位顺序
// 	Direct          int        // 方位(0-3) 东南西北
// 	State           int        // 状态
// 	IsReady         bool       // 是否准备
// 	IsZhuang        bool       // 是否是庄家
// 	IsNeedTool      bool       // 需要等待操作
// 	IsCanWin        bool       //
// 	IsCanPeng       bool       //
// 	IsCanKong       bool       //
// 	IsCanChow       bool       //
// 	IsDefeat        bool       // 认输
// 	Ting            int        // 是否听牌  >0  听牌
// 	Disband         int        // -1未操作, 0=反对  1=同意
// 	User            *user.User // 用户
// 	Cards           *UserCard  // 用户的所有牌(手牌, 出牌, 吃, 碰, 杠, 胡, 花)
// 	CreateTimeStamp int64      //秒
// }

/**
* huindex 胡牌位置id
* huCard  胡牌牌id
**/

// 牌的列表
// type UserCard struct {
// 	List *al.ArrayList // 牌的列表
// 	Kong *al.ArrayList // 杠的牌
// 	Peng *al.ArrayList // 碰的牌
// 	Chow *al.ArrayList // 吃的牌
// 	Out  *al.ArrayList // 打出的牌
// 	Hua  *al.ArrayList // 花牌
// 	Hu   *al.ArrayList // 胡的牌
// }
