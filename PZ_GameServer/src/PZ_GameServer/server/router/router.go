package router

import (
	"reflect"

	"github.com/golang/protobuf/proto"
)

type CallInfo struct {
	Id      int32                   // Message ID
	Type    int32                   // Info, Game Type (0=通用, >0=游戏类型)
	Msg     map[int32]proto.Message // Protobuf Message
	Evt_fun map[int32]reflect.Value // Function
	Fun     interface{}             //
}

var RouterMap = make(map[int32]CallInfo)

//绑定
func Bind(msgID int32, gameType int32, msg proto.Message, evt_fun interface{}) {

	if c, ok := RouterMap[msgID]; ok {
		c.Msg[gameType] = msg
		c.Evt_fun[gameType] = reflect.ValueOf(evt_fun)
		c.Fun = evt_fun
	} else {
		ci := CallInfo{
			Id:      msgID,
			Type:    gameType,
			Msg:     make(map[int32]proto.Message),
			Evt_fun: make(map[int32]reflect.Value),
		}
		ci.Msg[gameType] = msg
		ci.Evt_fun[gameType] = reflect.ValueOf(evt_fun)
		ci.Fun = evt_fun
		RouterMap[int32(msgID)] = ci
	}

}

//序列化Protobuf
func GetCallInfo(msgID int32, gameType int32, pt *[]byte) (proto.Message, reflect.Value, interface{}, error) {
	var err error
	var v interface{} = int64(0)

	if c, ok := RouterMap[msgID]; ok {
		if cm, gt_ok := c.Msg[gameType]; gt_ok {
			err = proto.Unmarshal(*pt, cm)
			return cm, c.Evt_fun[gameType], c.Fun, err
		}
	}
	return nil, reflect.ValueOf(v), v, err
}
