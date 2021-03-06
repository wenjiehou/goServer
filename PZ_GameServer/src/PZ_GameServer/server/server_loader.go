package server

import (
	"PZ_GameServer/model"
	"PZ_GameServer/net/network/gmHttp"
	"PZ_GameServer/server/game"
	"PZ_GameServer/server/game/roombase"
	"strconv"
	"time"
)

func Start() {
	err := initConfig()
	if err != nil {
		panic(err)
	}

	err = initDb() //连接数据库
	if err != nil {
		panic(err)
	}

	//初始化配置
	game.InitConfig(runTimeConfig.Debug)

	//初始化基础数据
	ServerInit()
	go gmHttp.StartHttp(runTimeConfig.GMHttp.Ip, runTimeConfig.GMHttp.Port)

	//初始化Redis数据库 IP , pwd, maxIdle, maxActive, idleTimeOut int
	roombase.Redis_InitRedisDb(runTimeConfig.Redis.Ip+":"+strconv.Itoa(runTimeConfig.Redis.Port), runTimeConfig.Redis.Pwd, runTimeConfig.Redis.MaxIdle, runTimeConfig.Redis.MaxActive, runTimeConfig.Redis.TimeOut)
	time.Sleep(1 * time.Second)
	roombase.Redis_ClearRedis()

	//初始化定时任务
	game.InitCrontab()

	//开启服务器
	game.StartWsServer(runTimeConfig.WSAddr, runTimeConfig.MaxConnNum)
}

// 服务器初始化 Init
func ServerInit() {
	model.GetUserModel().SetUserRoomIdToZero() //房间信息在内存，服务器重启，玩家room_id清0
}

func Stop() {

}
