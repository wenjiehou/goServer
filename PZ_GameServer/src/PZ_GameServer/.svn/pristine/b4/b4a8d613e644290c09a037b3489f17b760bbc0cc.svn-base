package testCard

import (
	"PZ_GameServer/common/util"
	rm "PZ_GameServer/server/game/room"
	"reflect"
)

type setter struct {
	RoomHandle *rm.RoomHandle
}

func (s *setter) SetNextCard(cid int, isBack int) {
	print("this is setnextcard testcard")
	util.FunCall(s.RoomHandle.Room, "SetNextCard", []reflect.Value{reflect.ValueOf(cid), reflect.ValueOf(isBack)})
}

func NewSetter(rid int) (*setter, string) {
	roomHandle, ok := rm.RoomList[rid]
	s := setter{}
	if rid <= 0 || !ok {
		return &s, "房间不存在"
	}
	s.RoomHandle = roomHandle
	return &s, ""
}

func (s *setter) SetInitCards(uid string, cids []string) {
	print("this is SetInitCards testcard")
	util.FunCall(s.RoomHandle.Room, "SetInitCards", []reflect.Value{reflect.ValueOf(uid), reflect.ValueOf(cids)})
}
