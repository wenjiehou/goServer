package network

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	logf "PZ_GameServer/log"

	//	"strconv"

	"github.com/gorilla/websocket"
)

type WSServer struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	MaxMsgLen       uint32
	HTTPTimeout     time.Duration
	NewAgent        func(*WSConn) Agent
	ln              net.Listener
	Handler         *WSHandler
}

type WSHandler struct {
	maxConnNum      int
	pendingWriteNum int
	maxMsgLen       uint32
	newAgent        func(*WSConn) Agent
	upgrader        websocket.Upgrader
	conns           WebsocketConnSet
	mutexConns      sync.Mutex
	wg              sync.WaitGroup
}

type clientConn struct {
	w  http.ResponseWriter
	r  *http.Request
	ch chan bool
}

var ClientConnList []*clientConn = make([]*clientConn, 0)

var HeimingIpList []string = make([]string, 0)

func ChenkInHeimingIp(ip string) bool {
	for _, v := range HeimingIpList {
		if v == ip {
			return true
		}
	}
	return false
}

func (handler *WSHandler) TimeTicker() { //连接缓存触发用，防止太多的人一起连接导致服务器挂了
	for {
		time.Sleep(10000000) //2毫秒
		if len(ClientConnList) > 0 {
			fmt.Println("Connection Total1", len(handler.conns))
			if ClientConnList[0] == nil || ClientConnList[0].r == nil || ClientConnList[0].w == nil || ChenkInHeimingIp(strings.Split(ClientConnList[0].r.RemoteAddr, ":")[0]) == true {
				if ClientConnList[0] != nil {
					ClientConnList[0].ch <- true
				}
				ClientConnList = append(ClientConnList[1:])
				fmt.Println("连接处理失败")
				continue
			}

			r := ClientConnList[0].r
			w := ClientConnList[0].w
			ch := ClientConnList[0].ch

			ClientConnList = append(ClientConnList[1:])

			//			fmt.Println("shengyu::" + strconv.Itoa(len(ClientConnList)))

			go func(cha chan bool) {
				defer func() {
					if e := recover(); e != nil {
						fmt.Println("Runtime error caught: %v", e)
					}
					cha <- true
				}()

				if r.Method != "GET" {
					http.Error(w, "Method not allowed", 405)
					return
				}
				conn, err := handler.upgrader.Upgrade(w, r, nil)
				if err != nil {
					logf.Debug("upgrade error: %v", err)
					return
				}
				conn.SetReadLimit(int64(handler.maxMsgLen))

				handler.wg.Add(1)
				defer handler.wg.Done()

				handler.mutexConns.Lock()
				if handler.conns == nil {
					handler.mutexConns.Unlock()
					conn.Close()
					return
				}

				if len(handler.conns) >= handler.maxConnNum {
					handler.mutexConns.Unlock()
					conn.Close()
					logf.Debug("TimeTicker too many connections")
					return
				}
				handler.conns[conn] = struct{}{}
				handler.mutexConns.Unlock()
				//	fmt.Println("len(handler.conns)::" + strconv.Itoa(len(handler.conns)))
				wsConn := newWSConn(conn, handler.pendingWriteNum, handler.maxMsgLen)
				agent := handler.newAgent(wsConn)
				//	fmt.Println("111111....1111111")
				agent.Run() //
				//	fmt.Println("222222....222222")

				// cleanup
				wsConn.Close()

				handler.mutexConns.Lock()
				delete(handler.conns, conn)
				handler.mutexConns.Unlock()

				agent.OnClose()
			}(ch)
		}
	}
}

func (handler *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ServeHTTP Client")
	cli := &clientConn{
		w:  w,
		r:  r,
		ch: make(chan bool),
	}

	newConIp := strings.Split(cli.r.RemoteAddr, ":")[0]

	fmt.Println("lianjielianjie", newConIp)

	if ChenkInHeimingIp(newConIp) == true {
		return
	}

	ClientConnList = append(ClientConnList, cli)

	ip := ""
	lianNum := 0
	//
	fmt.Println("ClientConnList::", len(ClientConnList))
	if len(ClientConnList) > 100 {
		for _, v := range ClientConnList {
			fmt.Println("lianNum::", lianNum)
			if v != nil && v.r != nil {
				temp := strings.Split(v.r.RemoteAddr, ":")[0]
				if ip != temp {
					ip = temp
					lianNum = 0
				} else if ip == temp {
					lianNum++
				}
				fmt.Println("lianNum::", lianNum)
				if lianNum >= 10 { //连续有10个相同，暂时把这20个去掉，并且把这个ip暂时纪录入黑名单
					HeimingIpList = append(HeimingIpList, temp)
				}
			}
		}
	}

	//	for _, v := range ClientConnList {
	//		if v != nil {
	//			fmt.Println("lianjielianjie", strings.Split(v.r.RemoteAddr, ":")[0])
	//		}
	//	}

	_, ok := <-cli.ch
	if ok {
		close(cli.ch)
		fmt.Println("chulile")
	}

	//下面是原来的
	return
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("Runtime error caught: %v", e)
		}
	}()

	//	panic("ddd")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	conn, err := handler.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logf.Debug("upgrade error: %v", err)
		return
	}
	conn.SetReadLimit(int64(handler.maxMsgLen))

	handler.wg.Add(1)
	defer handler.wg.Done()

	handler.mutexConns.Lock()
	if handler.conns == nil {
		handler.mutexConns.Unlock()
		conn.Close()
		return
	}

	if len(handler.conns) >= handler.maxConnNum {
		handler.mutexConns.Unlock()
		conn.Close()
		logf.Debug("ServeHTTP too many connections")
		return
	}
	handler.conns[conn] = struct{}{}
	handler.mutexConns.Unlock()
	//	fmt.Println("len(handler.conns)::" + strconv.Itoa(len(handler.conns)))
	wsConn := newWSConn(conn, handler.pendingWriteNum, handler.maxMsgLen)
	agent := handler.newAgent(wsConn)
	//	fmt.Println("111111....1111111")
	agent.Run() //
	//	fmt.Println("222222....222222")

	// cleanup
	wsConn.Close()

	handler.mutexConns.Lock()
	delete(handler.conns, conn)
	handler.mutexConns.Unlock()

	agent.OnClose()
}

func (server *WSServer) Start() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		logf.Fatal("%v", err)
	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 20480
		logf.Release("invalid MaxConnNum, reset to %v", server.MaxConnNum)
	}
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 20480
		logf.Release("invalid PendingWriteNum, reset to %v", server.PendingWriteNum)
	}
	if server.MaxMsgLen <= 0 {
		server.MaxMsgLen = 4096
		logf.Release("invalid MaxMsgLen, reset to %v", server.MaxMsgLen)
	}
	if server.HTTPTimeout <= 0 {
		server.HTTPTimeout = 10 * time.Second
		logf.Release("invalid HTTPTimeout, reset to %v", server.HTTPTimeout)
	}

	if server.NewAgent == nil {
		logf.Fatal("NewAgent must not be nil")
	}

	server.ln = ln
	server.Handler = &WSHandler{
		maxConnNum:      server.MaxConnNum,
		pendingWriteNum: server.PendingWriteNum,
		maxMsgLen:       server.MaxMsgLen,
		newAgent:        server.NewAgent,
		conns:           make(WebsocketConnSet),
		upgrader: websocket.Upgrader{
			HandshakeTimeout: server.HTTPTimeout,
			CheckOrigin:      func(_ *http.Request) bool { return true },
		},
	}

	httpServer := &http.Server{
		Addr:           server.Addr,
		Handler:        server.Handler, //代理了一下
		ReadTimeout:    server.HTTPTimeout,
		WriteTimeout:   server.HTTPTimeout,
		MaxHeaderBytes: 2048, //2048
	}
	fmt.Println("kaiqi websocket serve!!")

	go server.Handler.TimeTicker()

	go httpServer.Serve(ln)

}

func (server *WSServer) Close() {
	fmt.Println("nimabi")
	server.ln.Close()

	server.Handler.mutexConns.Lock()
	for conn := range server.Handler.conns {
		conn.Close()
	}
	server.Handler.conns = nil
	server.Handler.mutexConns.Unlock()

	server.Handler.wg.Wait()
}
