// Game Logic Server
package game

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	//	"PZ_GameServer/common/util"
	"PZ_GameServer/model"
	"PZ_GameServer/net/network"
	"PZ_GameServer/protocol/pb"
	"PZ_GameServer/server/game/room"
	"PZ_GameServer/server/game/room/ningbo"
	"PZ_GameServer/server/game/room/pinshi"
	"PZ_GameServer/server/game/room/srddz"
	"PZ_GameServer/server/game/room/xiangshan"
	"PZ_GameServer/server/game/room/xizhou"
	"PZ_GameServer/server/game/roombase"
	"PZ_GameServer/server/router"
	"PZ_GameServer/server/user"

	"github.com/golang/protobuf/proto"
)

var (
	GServer     = &GameServer{} // 游戏服务器实例
	room_count  = 500000        // 房间号  (TODO:  房间号逻辑需要重新修改,  房间号过多会冲突)
	offlineTime = 180           // 离线倒计时
)

type GameServer struct {
	ID            int                   // 服务器ID
	Type          int                   // 服务器 类型
	StartTime     int64                 // 服务器开始时间 time.Now().Unix()
	MaxUser       int                   // 服务器人数上限
	MaxRoom       int                   // 服务器房间上线
	UserList      map[string]*user.User // 用户列表
	broadcast     chan []byte           // 广播通道
	cmd           chan int              // 命令通道
	CheckUserList map[int]*user.User    // 用户列表，与userList不同
	mux           sync.RWMutex          // 针对map的互斥锁
}

// Init
func init() {
	GServer = NewGServer()
	go WatchUserList()
}

//初始化配置
func InitConfig(debug bool) {
	roombase.Debug = debug
}

//初始化GameServer
func NewGServer() *GameServer {

	room.InitGames() // 初始化全部游戏

	InitBaseRouter() // 初始化基础路由

	return &GameServer{
		MaxRoom:       2000,
		broadcast:     make(chan []byte),
		UserList:      make(map[string]*user.User),
		CheckUserList: make(map[int]*user.User),
	}
}

//
func StartWsServer(address string, maxConnNum int) {
	var ws = new(network.WSServer)
	ws.Addr = address
	ws.MaxConnNum = maxConnNum
	ws.MaxMsgLen = 10240
	ws.HTTPTimeout = 5 * time.Second
	ws.PendingWriteNum = 1000
	ws.NewAgent = GetNewUser
	ws.Start()

}

// 代理
func GetNewUser(conn *network.WSConn) network.Agent {
	u := user.GetUser(conn)
	return u
}

// 绑定
func Bind(msgID mjgame.MsgID, gameType int32, msg proto.Message, evt_fun interface{}) {
	router.Bind(int32(msgID), gameType, msg, evt_fun)
}

// 调用内部函数
func FunCall(m interface{}, n string, arg []reflect.Value) {
	t := reflect.ValueOf(m)
	f := t.MethodByName(n)
	if f.IsValid() {
		f.Call(arg)
	} else {
		fmt.Println("错误的反射方法 ", n)
	}
}

// 得到新的room id
//func GetNewRoomID() int {
//	temp := strconv.Itoa(util.RandInt(1, 999999))
//	mlen := 6 - len(temp)

//	for i := 0; i < mlen; i++ {
//		temp += "0"
//	}
//	room_count, _ = strconv.Atoi(temp)
//	return room_count

//	//	room_count += util.RandInt(3, 100)
//	//	if room_count > 999999 {
//	//		room_count = 100000
//	//	}
//	//	return room_count
//}

//创建房间
func (gs *GameServer) CreateRoom(croom *mjgame.Create_Room, user *user.User, rid int) int {
	//	if gs.MaxRoom > len(room.RoomList) {
	//		//		roomId := GetNewRoomID()
	//		var roomId int
	//		for {
	//			roomId = GetNewRoomID()
	//			_, hasRoom := room.RoomList[roomId]
	//			if !hasRoom {
	//				break
	//			}
	//		}
	//		if rid > 0 {
	//			roomId = rid
	//		}
	//		roomHandle := room.CreateRoom(roomId, croom, user)
	//		if roomHandle != nil {
	//			return roomId
	//		}
	//	}
	//	return -1

	roomHandle, roomId := room.CreateLockRoom(croom, user, rid)
	if roomHandle != nil {
		return roomId
	}

	return -1
}

//@andy获取用户列表锁
func (gs *GameServer) GetLockUser(sid string) (*user.User, bool) {
	gs.mux.Lock()
	defer gs.mux.Unlock()

	user, ok := gs.UserList[sid]
	return user, ok
}

//@andy获取用户列表锁
func (gs *GameServer) GetLockCheckUser(id int) (*user.User, bool) {
	gs.mux.Lock()
	defer gs.mux.Unlock()

	user, ok := gs.CheckUserList[id]
	return user, ok
}

//@andy设置用户游戏类型
func (gs *GameServer) UpdateGameType(sid string, gameType reflect.Type) {
	gs.mux.Lock()
	defer gs.mux.Unlock()

	if user, ok := gs.UserList[sid]; ok {
		//GServer.UserList[sid].GameType = gameType
		user.GameType = gameType
	}
}

//// 断开用户
//func (gs *GameServer) DisconnectUser(sid string) {

//	gs.UserList[sid] = nil
//}

//// 销毁房间
//func (gs *GameServer) DestoryRoom(rid int) {
//}

//// 获得房间
//func (gs *GameServer) GetRoom(rid int, pwd string) *Room {
//	return gs.RoomList[rid]
//}

func WatchUserList() {

	for {
		select {
		case u := <-user.ChanUser:
			rid := u.RoomId

			if u.RoomId > 0 {
				//if v, ok := room.RoomList[u.RoomId]; ok {
				if v, ok := room.GetLockRoomHandle(u.RoomId); ok {

					rtype := reflect.TypeOf(v.Room)
					if rtype == xiangshan.IFCXiangShanType {
						v.Room.(*xiangshan.RoomXiangshan).ExitUser(&u)
					} else if rtype == xizhou.IFCXiZhouType {
						v.Room.(*xizhou.RoomXiZhou).ExitUser(&u)
					} else if rtype == ningbo.IFCNingBoType {
						//v.Room.(*ningbo.RoomNingBo).ExitUser(&u)
					} else if rtype == srddz.IFCSrddzType {
						//v.Room.(*srddz.RoomSrddz).ExitUser(&u)
					} else if rtype == pinshi.IFCPinshiType {

					}

				}
			}
			model.GetUserModel().UpdateRoomID(u.User)
			BroadcastUserState(&u, rid)
			GServer.mux.Lock()
			delete(GServer.CheckUserList, u.ID)
			GServer.mux.Unlock()
		}
	}
}

//广播用户记录
func BroadcastUserState(user *user.User, roomid int) {

	if user == nil || roomid == 0 {
		//fmt.Println(" BroadcastUserState 为空", user)
		return
	}

	//if v, ok := room.RoomList[roomid]; ok {
	if v, ok := room.GetLockRoomHandle(roomid); ok {

		rtype := reflect.TypeOf(v.Room)

		userState := &mjgame.NotifyUserState{
			Id:          strconv.Itoa(user.ID),
			Status:      int32(user.State),
			OfflineTime: int32(offlineTime),
		}

		//更新记录当前掉线时间
		if rtype == xizhou.IFCXiZhouType {
			roomXiZhou := v.Room.(*xizhou.RoomXiZhou) // TODO 这里崩溃
			index := roomXiZhou.GetSeatIndexById(user.ID)
			if index >= 0 {
				roomXiZhou.Seats[index].OfflineTime = time.Now()
			}
			v.Room.(*xizhou.RoomXiZhou).BCMessage(mjgame.MsgID_MSG_NOTIFY_USER_STATE, userState)

		} else if rtype == xiangshan.IFCXiangShanType {

			roomXiangShan := v.Room.(*xiangshan.RoomXiangshan)
			index := roomXiangShan.GetSeatIndexById(user.ID)
			if index >= 0 {
				roomXiangShan.Seats[index].OfflineTime = time.Now()
			}
			v.Room.(*xiangshan.RoomXiangshan).BCMessage(mjgame.MsgID_MSG_NOTIFY_USER_STATE, userState)
		} else if rtype == ningbo.IFCNingBoType {

			roomNingbo := v.Room.(*ningbo.RoomNingBo)
			index := roomNingbo.GetSeatIndexById(user.ID)
			if index >= 0 {
				roomNingbo.Seats[index].OfflineTime = time.Now()
			}
			v.Room.(*ningbo.RoomNingBo).BCMessage(mjgame.MsgID_MSG_NOTIFY_USER_STATE, userState)
		} else if rtype == srddz.IFCSrddzType {
			roomSrddz := v.Room.(*srddz.RoomSrddz)
			index := roomSrddz.GetSeatIndexById(user.ID)
			if index >= 0 {
				roomSrddz.Seats[index].OfflineTime = time.Now()
			}
			v.Room.(*srddz.RoomSrddz).BCMessage(mjgame.MsgID_MSG_NOTIFY_USER_STATE, userState)
		} else if rtype == pinshi.IFCPinshiType {
			roomPinshi := v.Room.(*pinshi.RoomPinshi)
			index := roomPinshi.GetSeatIndexById(user.ID)
			if index >= 0 {
				roomPinshi.Seats[index].OfflineTime = time.Now()
			}
			v.Room.(*pinshi.RoomPinshi).BCMessage(mjgame.MsgID_MSG_NOTIFY_USER_STATE, userState)
		}

		GServer.UpdateGameType(user.Sid, user.GameType)
		//		if _, ok := GServer.UserList[user.Sid]; ok {
		//			GServer.UserList[user.Sid].GameType = user.GameType
		//		}
	}
}
