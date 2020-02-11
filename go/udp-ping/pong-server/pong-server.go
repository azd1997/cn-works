package main

import (
	"errors"
	"log"
	"math/rand"
	"net"
	"time"
)

type PongServer struct {
	Addrs []*net.UDPAddr
}

func NewPongServer(addrs ...*net.UDPAddr) *PongServer {
	if len(addrs)==0 {log.Fatalln("请输入至少一个*UDPAddr!")}
	return &PongServer{Addrs:addrs}
}

func (server *PongServer) ListenAndPong() {
	for i, addr := range server.Addrs {
		addr := addr	// 拷贝一次
		go server.listenAndPong(i, addr)
	}
}

func (server *PongServer) listenAndPong(no int, addr *net.UDPAddr) {
	// udp没有连接，开启监听之后就是写（谁连接了就可以收到）或者读
	// listen的类型是*UDPConn
	listener, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("第 %d 个监听地址 %s 开启监听失败: %s\n", no, addr, err)
	}
	log.Printf("第 %d 个监听地址 %s 开启监听成功\n", no, addr)

	// 随机延迟，模拟网络的不定性(这只是针对内网测试，内网延迟较低且几乎不可能超时，外网的话直接回就好了)
	// 正常延迟在0~499ms // 异常延迟在1s以上
	delays := make([]int, 800)
	for i:=0; i<500; i++ {delays[i] = i}
	for i:=500; i<800; i++ {delays[i] = 1000 + i}

	buff := make([]byte, 1024)
	rand.Seed(time.Now().Unix())	//设置随机数种子
	for {
		// 接入连接后，处理ping，返回pong
		n, clientAddr, err := listener.ReadFromUDP(buff)
		if err != nil {
			log.Printf("第 %d 个监听地址 %s 读 %s 失败: %s\n", no, addr, clientAddr, err)
			continue
		}
		if string(buff[:n])!="ping" {
			log.Printf("第 %d 个监听地址 %s 读 %s 失败: %s\n", no, addr, clientAddr, errors.New("not ping"))
			continue
		}

		// 随机等待
		time.Sleep(time.Duration(delays[rand.Intn(800)]) * time.Millisecond)
		// 回复pong消息
		_, _ = listener.WriteToUDP([]byte("pong"), clientAddr)
		log.Printf("第 %d 个监听地址 %s 已向 %s 回复pong\n", no, addr, clientAddr)
	}
}