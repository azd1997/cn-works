package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const (
	NETWORK = "tcp"
	IP = "localhost"
	PORT = 8000
)

var (
	ErrMethodNotGET = errors.New("request method not GET")
)


func main() {

	log.Println("Web服务器启动...")

	// 1. 绑定端口，开启监听
	listenAddr := IP + ":" + strconv.Itoa(PORT)
	listener, err := net.Listen(NETWORK, listenAddr)
	fatalErr(err)
	defer listener.Close()
	log.Printf("监听成功： %s\n", listenAddr)

	// 2. 死循环，等待连接
	for {
		// 2.1 接收client连接请求
		conn, err := listener.Accept()
		fatalErr(err)
		log.Printf("接受连接： %s\n", conn.RemoteAddr().String())

		// 2.2 在连接conn中处理逻辑
		handleConn(conn)
	}
}

// handleConn的任务是：
// 1. 解释所收到的请求(HTTP请求)，确定请求的文件
// 2. 从文件系统获取该文件
// 3. 构造请求的回复(HTTP响应报文)，如文件不存在，返回404报文
func handleConn(conn net.Conn) {

	// 1. 解释HTTP请求，我们默认为GET请求
	fileurl, err := resolveGETHTTP(conn)
	fatalErr(err)

	// 2. 寻找文件。当这一步出错时，不应该终止程序
	file, err := seekFile(fileurl)

	// 3. 回复HTTP报文
	if err != nil {
		response404(conn)
		log.Printf("文件不存在或打开出错: %s，返回404\n", err)
	} else {
		response200(conn, file)
	}

	// 关闭文件
	file.Close()
}

// 解析GET HTTP
// 	// HTTP请求的报文格式:
//	// 第一行(请求行)： Method URL Version
//	// 示例：  GET /somedir/page.html HTTP/1.1
//	// 接下来(首部行)： 首部字段名： 值
//	// 示例：  ContentType: text/json
//	// 首部行可以有多条，结束之后用一个空行标记
//	// 空行之后是真正内容，称 “实体体”
// 这里我们只支持题目要求的 GET file 这一种简单请求
// 这种请求，我们只需要处理 请求行
// 首部行不作设置
// 检查方法是否为GET， 然后获取请求的文件名(路径)
func resolveGETHTTP(conn net.Conn) (string, error) {

	// 读取请求第一行，也就是请求行
	r := bufio.NewReader(conn)
	header, _, err := r.ReadLine()
	if err != nil {return "", err}

	log.Printf("来自 %s 的请求行： %s\n", conn.RemoteAddr(), string(header))

	// 由于请求行以空格区分，所以直接将slice用空格断开
	// 第0个元素就是方法，第1个就是请求的文件路径
	reqSplited := bytes.Split(header, []byte{' '})

	// 检查方法
	method := string(reqSplited[0])
	if method != "GET" {return "", ErrMethodNotGET}

	// 获取文件url
	fileurl := string(reqSplited[1])

	// 返回
	return fileurl, nil
}


// 根据文件名寻找本地文件
// 寻找到则返回文件指针
// 没有则返回错误
// 例如请求头中url "/tmp/123.txt"
// 这里要将第一个"/"替换成服务器程序所在目录下
func seekFile(fileurl string) (*os.File, error) {

	// 获取程序当前绝对路径
	curPath, err := filepath.Abs(os.Args[0])
	if err != nil {return nil, err}
	curDir := filepath.Dir(curPath)		// 除非是根目录/，否则返回的Dir末尾不带/
	//fmt.Println(curDir)

	// 将fileurl转成filename
	filename := curDir + fileurl
	fmt.Println(filename)

	// 打开文件，返回文件描述符
	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	//fmt.Println(err)
	return file, nil
}


// HTTP 回复报文
// 第一行（状态行）： 版本 状态码 短语
// 示例： HTTP/1.1 200 OK
// 若干行（首部行）： 首部字段名: 值
// 与请求报文一致
// 空行
// 实体体

// 回复时首部行不做设置
// 因此状态行后需跟两个换行符 (\n)

// 回复HTTP报文
func response200(conn net.Conn, file *os.File) {
	// 1. 构建状态行
	header := "HTTP/1.1 " + strconv.Itoa(http.StatusOK) + " " + http.StatusText(http.StatusOK) + "\n\n"

	// 2. 将状态行写入
	_, err := conn.Write([]byte(header))
	fatalErr(err)

	// 3. 将文件内容写入
	buf := make([]byte, 4096)
	for {
		// 读文件内容到buf缓冲区
		n, err := file.Read(buf)
		if err == io.EOF {		// 读完
			log.Printf("向 %s 发送文件完毕\n", conn.RemoteAddr())
			return
		}
		fatalErr(err)

		// 写到conn
		_, err = conn.Write(buf[:n])
		fatalErr(err)
	}

}

// 回复HTTP报文
func response404(conn net.Conn) {
	// 1. 构建状态行
	header := "HTTP/1.1 " + strconv.Itoa(http.StatusNotFound) + http.StatusText(http.StatusNotFound) + "\n\n"

	// 2. 将状态行写入
	_, err := conn.Write([]byte(header))
	fatalErr(err)
}

// err不为空时退出线程，就是退出程序
func fatalErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}