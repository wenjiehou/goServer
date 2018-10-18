package network

import (
	"fmt"
	"net"
)

var (
	conn      *TCPConn
	msgParser *MsgParser
)

// 连接数据库
func DBConnect(ipAddr string) {

	fmt.Print("\rDB Connect " + ipAddr)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ipAddr)
	con, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Print("DBServer连接失败, 请检查IP地址,端口号. \r\n")
		return
	}
	msgParser = NewMsgParser()
	conn = newTCPConn(con, 1024, msgParser)
	fmt.Print(" DBServer Connect Successed. \r\n")
}

//检查用户是否存在
//func DBCheckUser(uid string) bool {
//	if conn != nil {
//		conn.WriteMsg([]byte(msg))
//	} else {
//		fmt.Println("DBServer 无连接.")
//	}
//	return false
//}

// 发送消息
func SendMessage(msg string) {
	if conn != nil {
		conn.WriteMsg([]byte(msg))
	} else {
		fmt.Println("DBServer 无连接.")
	}
}

// 发送消息Byte
func SendMessageByte(msg []byte) {
	if conn != nil {
		conn.WriteMsg(msg)
	} else {
		fmt.Println("DBServer 无连接.")
	}
}
