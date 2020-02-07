package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// 针对server的简单客户端
// 逻辑：
// 1. 尝试与server建立TCP连接
// 2. 构建HTTP请求头，发送过去
// 3. 创建文件指针，等待服务器回传数据
// 4. 解析回传数据，解析header，再决定是否保存文件或退出

const (
	NETWORK = "tcp"
	SERVER_IP = "localhost"
	SERVER_PORT = 8000
	BUFF_SIZE = 8

	REQ_FILE = "/tmp/123.txt"
)

func main() {

	log.Println("客户端启动...")

	// 连接服务器
	serverAddr := SERVER_IP + ":" + strconv.Itoa(SERVER_PORT)
	conn, err := dialServer(serverAddr)
	if err != nil {
		log.Fatalf("连接服务器 %s 失败: %s，程序退出\n", serverAddr, err)
	}
	defer conn.Close()
	log.Printf("连接服务器 %s 成功\n", serverAddr)


	// 请求文件
	var file *os.File
	file, err = requestFile(conn, REQ_FILE)
	if err != nil {
		log.Printf("向服务器 %s 请求文件 %s 失败： %s\n", serverAddr, REQ_FILE, err)
	}
	file.Close()
}

// 一次连接不成尝试第二次
func dialServer(server string) (net.Conn, error) {
	conn, err := net.Dial(NETWORK, server)
	if err != nil {
		log.Printf("第一次连接服务器 %s 失败，尝试重新连接...\n", server)
		conn, err = net.Dial(NETWORK, server)
		if err != nil {
			log.Printf("第二次连接服务器 %s 失败，程序即将退出...\n", server)
			return nil, err
		}
		return conn, nil
	}
	return conn, nil
}


// 请求文件
func requestFile(conn net.Conn, fileurl string) (*os.File, error) {

	// 构造请求行
	method := http.MethodGet
	url := fileurl
	version := "HTTP/1.1"
	header := method + " " + url + " " + version + "\n\n"
	_, err := conn.Write([]byte(header))
	if err != nil {return nil, err}
	log.Println("向服务器发送请求: ", header)

	// 发送成功后等待接收
	// 获取程序当前绝对路径
	curPath, err := filepath.Abs(os.Args[0])
	if err != nil {return nil, err}
	curDir := filepath.Dir(curPath)		// 除非是根目录/，否则返回的Dir末尾不带/
	filename := curDir + fileurl
	fmt.Println(filename)
	file, err := os.Create(filename)		// 存储在和服务器一样的路径
	if err != nil {return nil, err}

	// 检查响应头
	r := bufio.NewReader(conn)
	respHeader, _, err := r.ReadLine()
	if err != nil {return nil, err}
	respH := bytes.Split(respHeader, []byte{' '})
	fmt.Println(string(respHeader))
	if string(respH[2]) != http.StatusText(http.StatusOK) {
		return nil, errors.New("404")
	}

	// 接收相应数据
	buf := make([]byte, BUFF_SIZE)
	r.ReadLine()	// 我们知道还有一个空行，所以直接读掉
	for {
		//fmt.Println("111")
		n, err := r.Read(buf)
		//fmt.Println(n, err)
		if n<BUFF_SIZE {	// 说明读完了。 这里很奇怪，n==0和err==io.EOF都行不通
			_, err = file.Write(buf[:n])
			if err!=nil {return nil, err}
			log.Println("文件接收结束")
			return file, nil
		}
		if err!=nil {return nil, err}
		//fmt.Println("buf=", string(buf))

		_, err = file.Write(buf[:n])
		if err!=nil {return nil, err}
	}
}