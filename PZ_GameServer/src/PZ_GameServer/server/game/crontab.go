package game

import (
	"PZ_GameServer/model"

	"github.com/robfig/cron"
)

func InitCrontab() {
	c := cron.New()
	spec := "*/30 * * * * ?"
	c.AddFunc(spec, SyncNotice)

	//清理过期公告
	spec = "0 0 */1 * * ?"
	c.AddFunc(spec, ClearNotice)

	//停服公告
	spec = "*/5 * * * * ?"
	c.AddFunc(spec, GetStopGameNotice)

	c.Start()
}

func SyncNotice() {
	notices, err := model.GetNoticeModel().GetNotices(queue.maxId)
	if err != nil {
		return
	}

	for _, notice := range notices {
		Add(notice)
	}
}

func ClearNotice() {
	Remove()
}
