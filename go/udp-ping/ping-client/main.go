package main

import (
	"time"
)

const (
	DEFAULT_NETWORK = "udp"
	SERVER_ADDR1 = ":8000"
	SERVER_ADDR2 = ":8001"
	PING = "ping"
	PONG = "pong"
	READ_TIMEOUT time.Duration = 1*time.Second	// 1s
	PING_NUM = 10
)


func main() {
	client := NewPingTool(READ_TIMEOUT, PING_NUM, DEFAULT_NETWORK)
	//client.PingServer(SERVER_ADDR1, SERVER_ADDR2)
	client.PingServerConcurrent(SERVER_ADDR1, SERVER_ADDR2)
}

