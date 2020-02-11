package main

import (
	"log"
	"net"
	"sync"
	"time"
)

// 功能类
// 设计是可以一次ping多个主机，新建时需要设置超时和ping次数以及协议
type PingClient struct {
	timeout time.Duration	// ns
	pingnum int
	network string

	wg *sync.WaitGroup
}

func NewPingTool(timeout time.Duration, pingnum int, network string) *PingClient {
	return &PingClient{
		timeout: timeout,
		pingnum: pingnum,
		network: network,
		wg:&sync.WaitGroup{},
	}
}

// 单go程版本
func (client *PingClient) PingServer(servers ...string) [][]int64 {
	n := len(servers)
	res := make([][]int64, n)
	for i:=0; i<n; i++ {
		res[i] = make([]int64, client.pingnum)
	}
	for i:=0; i<n; i++ {
		res[i] = client.pingServer(servers[i])
	}
	return res
}

// 并发版本
func (client *PingClient) PingServerConcurrent(servers ...string) [][]int64 {
	n := len(servers)
	res := make([][]int64, n)
	for i:=0; i<n; i++ {
		res[i] = make([]int64, client.pingnum)
	}

	client.wg.Add(n)	// 等待n个结束
	for i:=0; i<n; i++ {
		go func(i int) {	// 这里需要传参，否则会形成闭包引用，下面的i会用的是循环结束最终的i也就是n，会导致index of panic
			res[i] = client.pingServer(servers[i])
			client.wg.Done()
		}(i)
	}

	client.wg.Wait()	// 等待这些事情结束

	return res
}

func (client *PingClient) pingServer(server string) []int64 {
	rtts := make([]int64, client.pingnum)

	// 连接Pong Server
	conn, err := net.DialTimeout(client.network, server, 1*time.Second)	// 这里的超时不是本题要求
	if err != nil {
		log.Fatalf("尝试向 %s 发起 %s 连接失败： %s\n", server, client.network, err)
	}

	// 连续ping10次
	for i:=0; i<client.pingnum; i++ {
		t1 := time.Now()	// 初始计时
		// 发送ping
		_, err = conn.Write([]byte(PING))
		if err != nil {
			log.Printf("第 %d 次 Ping %s 出错： %s\n", i, server, err)
			continue
		}
		// 等待pong
		_ = conn.SetReadDeadline(t1.Add(client.timeout))	// 设置UDP读超时
		buff := make([]byte, 1024)
		n, err := conn.Read(buff)
		if err != nil {
			log.Printf("第 %d 次 Ping %s 出错： %s\n", i, server, err)
			continue
		}
		// 计算RTT，记录，打印
		t2 := time.Now()
		rtts[i] = t2.Sub(t1).Nanoseconds()
		log.Printf("第 %d 次 Ping %s 成功， 接收到回复： %s, 往返时延: %s\n", i, server, string(buff[:n]), t2.Sub(t1).Round(time.Millisecond).String())
	}
	return rtts
}

