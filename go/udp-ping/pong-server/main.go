package main

import (
	"net"
)

// 测试用的 UDP Pong服务端

func main() {

	listenAddr1 := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8000,
	}
	listenAddr2 := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8001,
	}

	pongserver := NewPongServer(listenAddr1, listenAddr2)
	pongserver.ListenAndPong()

	select {}	// 阻止主go程退出
}

