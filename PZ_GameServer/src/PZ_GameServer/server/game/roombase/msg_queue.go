package roombase

import (
	"PZ_GameServer/server/user"
	//"time"
	"fmt"
	"reflect"
	"sync"
)

// 消息队列
type MsgQueue struct {
	IsRun    bool             // 是否运行
	MsgQueue []*user.RunParam // 消息队列

	isProcess bool       // 是否是在处理过程中
	runChan   chan int   // 消息通道
	count     int        //
	mxLock    sync.Mutex //
}

// 初始化
func (mq *MsgQueue) Init() {
	mq.MsgQueue = make([]*user.RunParam, 0)
	mq.runChan = make(chan int, 1)
}

// 添加消息
func (mq *MsgQueue) AddMessage(param *user.RunParam) {
	mq.mxLock.Lock()
	mq.MsgQueue = append(mq.MsgQueue, param)
	mq.mxLock.Unlock()
	mq.next()
}

func (mq *MsgQueue) Stop() {
	mq.runChan <- -1
}

func (mq *MsgQueue) next() {
	if mq.isProcess || len(mq.runChan) >= 1 || len(mq.MsgQueue) == 0 {
		return
	}
	mq.isProcess = true
	mq.runChan <- 1

}

// 运行
func (mq *MsgQueue) Run() {

	if mq.IsRun {
		return
	}
	mq.IsRun = true

	go func() {

		for {
			select {
			case status := <-mq.runChan:
				switch status {
				case 1:
					mq.count++
					mq.runEvent()
					//case -1:
					//break
				}
			}
		}
		//mq.IsRun = false
	}()

}

// 得到要运行的事件
func (mq *MsgQueue) runEvent() {
	if len(mq.MsgQueue) == 0 {
		return
	}
	mq.isProcess = true
	evt := mq.MsgQueue[0]
	if evt.FunName != "" && evt.Room != nil {
		t := reflect.ValueOf(evt.Room)
		f := t.MethodByName(evt.FunName)
		if f.IsValid() {
			f.Call(evt.Params)
		} else {
			fmt.Println("MsgQueue错误的反射方法 ", evt.FunName)
		}
	} else {
		evt.Fun.Call(evt.Params)
	}
	mq.clearFirstEvent()
	mq.isProcess = false
	mq.next()
}

// 清除事件
func (mq *MsgQueue) clearFirstEvent() {
	mq.mxLock.Lock()
	mq.MsgQueue = append(mq.MsgQueue[1:])
	mq.mxLock.Unlock()
}

func (mq *MsgQueue) StopAll() {
	mq.MsgQueue = []*user.RunParam{}
}
