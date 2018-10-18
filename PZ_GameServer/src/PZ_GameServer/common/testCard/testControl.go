package testCard

import (
	"PZ_GameServer/protocol/pb"
	"PZ_GameServer/server/user"
	"encoding/json"
	"fmt"
)

func SetNextCard(m *mjgame.MessageJson, a *user.User) {
	params := struct {
		Rid        int
		Cid        int
		IsBack     int
		HandleType int
	}{}

	errstr := json.Unmarshal([]byte(m.GetJSON()), &params)
	if errstr != nil {
		a.SendMessage(mjgame.MsgID_MSG_ACK_MSG, &mjgame.ACK_MSG{Type: -1, MSG: "设置失败:" + errstr.Error()})
	}

	if params.HandleType == 10 {
		setter, errstr := NewSetter(params.Rid)
		if len(errstr) != 0 {
			a.SendMessage(mjgame.MsgID_MSG_ACK_MSG, &mjgame.ACK_MSG{Type: -1, MSG: "设置失败:" + errstr})
			return
		} else {
			setter.SetNextCard(params.Cid, params.IsBack)
		}
	}
	a.SendMessage(mjgame.MsgID_MSG_ACK_MSG, &mjgame.ACK_MSG{Type: 1, MSG: "设置成功"})
}
func SetInitCards(m *mjgame.MessageJson, a *user.User) {
	params := struct {
		Rid  int
		Uid  string
		Cids []string
	}{}
	err := json.Unmarshal([]byte(m.GetJSON()), &params)
	if err != nil {
		a.SendMessage(mjgame.MsgID_MSG_ACK_MSG, &mjgame.ACK_MSG{Type: -1, MSG: "设置失败:" + err.Error()})
		fmt.Println("首牌设置失败,错误为", err.Error())
	}
	setter, errstr := NewSetter(params.Rid)
	if len(errstr) != 0 {
		a.SendMessage(mjgame.MsgID_MSG_ACK_MSG, &mjgame.ACK_MSG{Type: -1, MSG: "设置失败:" + errstr})
		return
	} else {
		setter.SetInitCards(params.Uid, params.Cids)
	}

	a.SendMessage(mjgame.MsgID_MSG_ACK_MSG, &mjgame.ACK_MSG{Type: 1, MSG: "设置成功"})

}
