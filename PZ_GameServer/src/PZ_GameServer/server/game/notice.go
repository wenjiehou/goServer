package game

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"PZ_GameServer/model"
	"PZ_GameServer/protocol/pb"
	"PZ_GameServer/server/user"
)

type Queue struct {
	maxId uint
	list  *list.List
	mutex sync.RWMutex
}

var queue = Queue{
	list: list.New(),
}

func Add(notice *model.Notice) {

	// 去掉, 会造成cpu bug
	//	queue.mutex.Lock()
	//	defer queue.mutex.Unlock()

	//	queue.maxId = notice.ID
	//	queue.list.PushBack(notice)
}

func Remove() {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	var next *list.Element
	for v := queue.list.Front(); v != nil; {
		notice := v.Value.(*model.Notice)
		if notice.EndTime.Before(time.Now()) {
			next = v.Next()
			queue.list.Remove(v)
			v = next
		}
	}
}

func GetNoticesById(id uint) []*mjgame.NoticeInfo {
	var notices = make([]*mjgame.NoticeInfo, 0)

	//	for v := queue.list.Front(); v != nil; v = v.Next() {
	//		notice := v.Value.(*model.Notice)
	//		if notice.ID > id {
	//			o := &mjgame.NoticeInfo{
	//				Id:      int32(notice.ID),
	//				Content: notice.Content,
	//			}
	//			notices = append(notices, o)
	//		}
	//	}

	return notices
}

var StopGameNotice *model.Notice

func GetStopGameNotice() {
	notice, err := model.GetNoticeModel().GetStopGameNotice()
	StopGameNotice = notice
	if err != nil || notice == nil {
		return
	}

	SendAllStopGameNotice()
	return
}

func SendStopGameNotice(user *user.User) {
	if StopGameNotice == nil {
		fmt.Println("SendStopGameNotice failed err:StopGameNotice == nil")
		return
	}
	notice := &mjgame.GameNotice{
		Id:      int32(StopGameNotice.ID),
		Content: StopGameNotice.Content,
	}
	user.SendMessage(mjgame.MsgID_MSG_NOTIFY_GAMENOTICE, notice)

	user.StopGameNoticeLog[StopGameNotice.ID] = true
}

func SendAllStopGameNotice() {
	//@andy
	//GameServer.mux.Lock()
	//defer GameServer.mux.Unlock()

	noticeId := StopGameNotice.ID
	for _, user := range GServer.CheckUserList {
		if v, ok := user.StopGameNoticeLog[noticeId]; ok && v {
			fmt.Println("SendAllStopGameNotice failed err:StopGameNoticeLog")
			continue
		}
		SendStopGameNotice(user)
	}
}
