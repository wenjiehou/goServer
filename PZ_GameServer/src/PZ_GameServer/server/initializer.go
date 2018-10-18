package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	logf "PZ_GameServer/log"
	"PZ_GameServer/model"

	"PZ_GameServer/server/game/room"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var (
	runTimeConfig    *Config
	serverConfigPath = "../config/server.json"
	gameConfigPath   = "../config/game.json"
	configPath       = "../config/createroom.json"
)

func initConfig() error {
	// 服务器配置
	data, err := ioutil.ReadFile(serverConfigPath)
	if err != nil {
		log.Fatal("Can't read config file "+serverConfigPath+"   %v", err)
		return err
	}
	runTimeConfig = new(Config)
	err = json.Unmarshal(data, runTimeConfig)
	fmt.Println("dd:::" + runTimeConfig.GMHttp.Ip + " " + strconv.Itoa(runTimeConfig.GMHttp.Port))
	if err != nil {
		log.Fatal("Read json file error "+serverConfigPath+"   %v", err)
		return err
	}

	// 游戏规则配置
	data, err = ioutil.ReadFile(gameConfigPath)
	if err != nil {
		log.Fatal("Can't read config file "+gameConfigPath+"   %v", err)
		return err
	}
	err = json.Unmarshal(data, &RoomRule)
	if err != nil {
		log.Fatal("Read json file error "+gameConfigPath+"   %v", err)
		return err
	}

	err = room.LoadConfig(configPath)
	if err != nil {
		return err
	}
	// 日志
	logf.InitConfig(runTimeConfig.LogLevel, runTimeConfig.LogPath, runTimeConfig.LogOutput)

	return nil
}

func initDb() error {
	db, err := gorm.Open("mysql", runTimeConfig.Db.String())

	if err != nil {
		return err
	}

	err = db.DB().Ping()
	if err != nil {
		return err
	}

	if err == nil {
		fmt.Println("Connect Database [" + runTimeConfig.Db.Ip + ":" + strconv.Itoa(runTimeConfig.Db.Port) + "] Successed")
	}

	db.DB().SetMaxOpenConns(runTimeConfig.Db.MaxConnect)
	db.DB().SetMaxIdleConns(runTimeConfig.Db.MaxIdle)

	err = db.AutoMigrate(
		model.User{},
		model.BattleRecord{},
		model.ReplayRecord{},
		model.UserItem{},
		model.Mail{},
		model.Notice{},
		model.Order{},
		model.RoomRecord{},
		model.Suggestion{},
		model.LogDiamond{},
		model.LogItem{},
		model.LogLogin{},
		model.Account{},
		model.Room{},
		model.UserDeny{},
	).Error

	if err != nil {
		return err
	}

	db.LogMode(runTimeConfig.Db.Debug)

	model.InitCommonDb(db)

	return nil
}
