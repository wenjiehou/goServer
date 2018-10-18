package game

import (
	"PZ_GameServer/protocol/pb"
	room "PZ_GameServer/server/game/room"
	//"PZ_GameServer/server/game/room/xiangshan"
	//"PZ_GameServer/server/game/room/xizhou"
	"PZ_GameServer/model"
	"PZ_GameServer/server/user"
	"encoding/json"
	"reflect"
	"strconv"
)

//服务器在线总人数
func getTotalUserOnline(args ...interface{}) {
	//m := args[0].(*mjgame.MessageJson) 保留参数
	a := args[1].(*user.User)
	totalUsers := len(GServer.CheckUserList)
	totalRooms := len(room.RoomList)
	runRooms := 0

	//@andy
	GServer.mux.Lock()
	userlist, _ := json.Marshal(GServer.CheckUserList)
	GServer.mux.Unlock()
	userlist = []byte("test set is null")
	str, _ := json.Marshal(map[string]string{
		"totalUsers": strconv.Itoa(totalUsers),
		"totalRooms": strconv.Itoa(totalRooms),
		"runRooms":   strconv.Itoa(runRooms),
		"userlist":   string(userlist),
	})
	a.SendMessage(mjgame.MsgID_MSG_ACK_MessageJson,
		&mjgame.ACK_MessageJson{JSON: string(str)})

}

func GetNewRoomId(args ...interface{}) {
	a := args[1].(*user.User)
	//	totalRooms := len(room.RoomList)
	//	NewRoomID := 0
	//	var tempCreatRoomTime int64
	//	if totalRooms > 0 {
	//		for _, v := range room.RoomList {
	//			if tempCreatRoomTime < v.CreateTime {
	//				tempCreatRoomTime = v.CreateTime
	//				NewRoomID = v.RoomId
	//			}
	//		}
	//	}
	NewRoomID := room.NewRoomId()
	str, _ := json.Marshal(map[string]string{
		"NewRoomID": strconv.Itoa(NewRoomID),
	})
	a.SendMessage(mjgame.MsgID_MSG_ACK_MessageJson,
		&mjgame.ACK_MessageJson{JSON: string(str)})
}

func GetRoomRecord(args ...interface{}) {
	m := args[0].(*mjgame.MessageJson)
	a := args[1].(*user.User)
	params := &struct {
		Rid int
	}{}

	errstr := json.Unmarshal([]byte(m.GetJSON()), params)
	if errstr != nil {
		a.SendMessage(mjgame.MsgID_MSG_ACK_MSG, &mjgame.ACK_MSG{Type: -1, MSG: "查看失败，查询参数错误:" + errstr.Error()})
	}
	if params == nil {
		a.SendMessage(mjgame.MsgID_MSG_ACK_MSG, &mjgame.ACK_MSG{Type: -1, MSG: "查看失败，查询房间id错误"})
	}
	//roomHandle, ok := room.RoomList[params.Rid]
	roomHandle, ok := room.GetLockRoomHandle(params.Rid)

	if params.Rid <= 0 || !ok {
		a.SendMessage(mjgame.MsgID_MSG_ACK_MSG, &mjgame.ACK_MSG{Type: -1, MSG: "房间不存在"})
		return
	}
	FunCall(roomHandle.Room, "GetRoomRecord", []reflect.Value{reflect.ValueOf(a)})
}

//检测玩家是否在本服务器上的房间内
func CheckUserInRoom(openid string) bool {
	user, _ := model.GetUserModel().GetUserByOpenId(openid)
	if user == nil || user.RoomId <= 0 {
		return false
	}
	return true
}

//检测当前加入的房间是否在这台服务器上 如果都不在的话，那房间就结束了，随便哪台服务器都可以用
func CheckRoomInServer(roomId int) bool {
	//_, ok := room.RoomList[roomId]
	_, ok := room.GetLockRoomHandle(roomId)
	if ok {
		return true
	}
	return false
}
