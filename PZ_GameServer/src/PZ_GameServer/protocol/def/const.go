package def

const (
	KongTypeMing = 1 //明杠
	KongTypeAn   = 2 //暗杠
	KongTypePeng = 3 //碰后杠
)

const (
	Tool_Hu   = 0
	Tool_Kong = 1
	Tool_Peng = 2
	Tool_Chow = 3
	Tool_Put  = 4
	Tool_Pass = 5
)

const (
	DoTypePengNumber = 3 //碰牌数量
	DoTypeKongNumber = 4 //杠牌数量
)

const (
	Voice = iota + 1 //语音
	Text             //文本
)

//const (
//	XiangShanDrawCount = 16  //西周流局牌数
//	KickTimeDuration   = 180 //可以被踢出的时间（秒）
//	WaitStartTime      = 15  //最大等待开局时间(分钟)
//	MaxOfflineTime     = 3   //最大离线时间(分钟)
//  VoteTimeOut        = 60  //投票超时时间
//)

const (
	XiangShanDrawCount = 16      //西周流局牌数
	KickTimeDuration   = 1800000 //180 //可以被踢出的时间（秒）
	JinBiKickTime      = 1800000 //30s 金币场可以被踢出的掉线时间
	WaitStartTime      = 900     //最大等待开局时间(秒)
	MaxOfflineTime     = 120     //最大离线时间(分钟)180
	VoteTimeOut        = 30      //投票超时时间(秒)
)

const (
	Online = iota + 1
	Offline
	Kick    //踢掉了，在线状态
	OffKick //踢掉了，掉线状态
)

const (
	QuitTimeTicker = iota + 1 //退出定时器
)
