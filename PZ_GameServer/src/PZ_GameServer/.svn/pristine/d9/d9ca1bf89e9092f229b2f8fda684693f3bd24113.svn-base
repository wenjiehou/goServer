package common

//文件作用：针对不同结构之间转换，向客户端定义结构输出

import (
	"strconv"

	"PZ_GameServer/protocol/pb"
	rb "PZ_GameServer/server/game/roombase"
)

func BuildSeatBaseToAckUserInfo(seat *rb.SeatBase) *mjgame.ACK_User_Info {
	var user mjgame.ACK_User_Info

	var ip string
	if seat.User.Conn != nil {
		ip = seat.User.Conn.RemoteAddr().String()
	}

	user.Uid = strconv.Itoa(seat.User.ID)
	user.Index = int32(seat.Index)
	user.Ip = ip
	user.Name = seat.User.NickName
	user.Icon = seat.User.Icon
	user.Robot = int32(seat.User.IsRobot)
	user.Coin = int32(seat.User.Coin)
	user.GPS_LAT = seat.User.GPS_LAT
	user.GPS_LNG = seat.User.GPS_LNG
	user.Diamond = int32(seat.User.Diamond)
	user.State = int32(seat.State)
	user.Score = seat.Accumulation.Score

	return &user
}

//
func BuildSeatBaseToVotes(votes []int, starterIndex int, seats []*rb.SeatBase) []*mjgame.DisbandItem {
	var disbandItems = make([]*mjgame.DisbandItem, 0)

	for i := 0; i < len(seats); i++ {
		if seats[i].User == nil {
			continue
		}

		disbandItem := &mjgame.DisbandItem{
			UserId:   strconv.Itoa(seats[i].User.ID),
			NickName: seats[i].User.NickName,
			Icon:     seats[i].User.Icon,
		}
		disbandItem.Result = int64(votes[i])

		if starterIndex == i {
			disbandItem.IsStarter = true
		}

		disbandItems = append(disbandItems, disbandItem)
	}

	return disbandItems
}
