package server

import "fmt"

// 配置文件类

//GameServer
type Config struct {
	Debug       bool   //是否调试
	LogLevel    string //Log等级
	LogPath     string //Log路径
	LogOutput   bool   //是否输出到文件
	WSAddr      string //WebSocket路径
	TCPAddr     string //
	MaxConnNum  int    //最大连接数
	ConsolePort int
	ProfilePath string
	LenStackBuf uint32

	// gate conf
	PendingWriteNum int
	MaxMsgLen       int
	HTTPTimeout     int
	LenMsgLen       int
	LittleEndian    bool

	// Redis conf
	Redis RedisDb

	GMHttp GMHttpDb

	// db conf
	DBAddr   string
	DB_Token string

	// DB
	Db DataBase
}

type DataBase struct {
	Type       string
	Ip         string
	Port       int
	User       string
	Pwd        string
	TimeOut    int
	MaxConnect int
	MaxIdle    int
	Name       string
	Charset    string
	Debug      bool
}

type RedisDb struct {
	Type      string
	MaxIdle   int
	MaxActive int
	TimeOut   int
	Ip        string
	Port      int
	User      string
	Pwd       string
}

type GMHttpDb struct {
	Ip   string
	Port int
}

//房间规则
var RoomRule struct {
	DiscardTimeOut     int // 打牌的最大时间
	CheckTimeOut       int // 吃胡碰杠的判断时间
	Change3CardTimeOut int // 换3张的时间
	MissTypeTimeOut    int // 定缺的时间
}

func (d *DataBase) String() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		d.User, d.Pwd, d.Ip, d.Port, d.Name, d.Charset)
}
