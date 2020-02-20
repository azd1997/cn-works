package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

func StartProxy(addr, port string, anony, debug bool) {
	proxy := NewProxy(addr, port, anony, debug)
	log.Printf("Proxy is runing on %s:%s \n", proxy.Addr, proxy.Port)
	log.Fatalln(http.ListenAndServe(proxy.Addr + ":" + proxy.Port, proxy))
}

// HTTP/HTTPS代理服务器
type Proxy struct {
	Addr string			// 监听地址
	Port string			// 监听端口
	IsAnonymous bool	// 是否高匿名模式
	Debug bool			// 调试模式
}

func NewProxy(addr, port string, anony, debug bool) *Proxy {
	return &Proxy{
		Addr:        addr,
		Port:        port,
		IsAnonymous: anony,
		Debug:       debug,
	}
}

func (proxy *Proxy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// 调试模式下，打印接收到的请求
	if proxy.Debug {
		log.Printf("Received request %s %s %s\n", r.Method, r.Host, r.RemoteAddr)
	}

	// 如果方法不是CONNECT就当HTTP处理，是的话，进行HTTPS直连
	if r.Method != http.MethodConnect {
		proxy.HTTP(rw, r)
	} else {
		// 直通模式不作任何处理
		proxy.HTTPS(rw, r)
	}
}

func (proxy *Proxy) HTTP(w http.ResponseWriter, r *http.Request) {
	//
	transport := http.DefaultTransport

	// 1. 复制原请求，并加以修改
	outReq := new(http.Request)
	*outReq = *r	// 复制内容

	if !proxy.IsAnonymous {		// 不匿名的话，把来路的节点地址给加在"X-Forwarded-For"字段
		if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			// 查看请求头中"X-Forwarded-For"字段是否有值
			// 这个字段会存转发路径上所有节点地址，而且每一次都是当前节点去填写上一个节点的地址
			if prior, ok := outReq.Header["X-Forwarded-For"]; ok {
				// 拼接路径上所有节点地址
				clientIP = strings.Join(prior, ",") + "," + clientIP
			}
			// 重新填入
			outReq.Header.Set("X-Forwarded-For", clientIP)
		}
	}

	// 2. 将修改后的请求转发出去，等待服务器回应
	resp, err := transport.RoundTrip(outReq)
	if err != nil {
		// 出错，那么说明这台代理服务器无法连通到，那么在w的状态头中写上BadGateway
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(err.Error()))
		return
	}
	defer resp.Body.Close()

	// 3. 回写http头。 将真正服务器的状态头给写到w的状态头去(就是写回给客户端了)
	for key, value := range resp.Header {
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}

	// 回写状态码
	w.WriteHeader(resp.StatusCode)
	// 回写响应内容
	_, _ = io.Copy(w, resp.Body)
}

func (proxy *Proxy) HTTPS(w http.ResponseWriter, r *http.Request) {
	// 取出host
	host := r.URL.Host
	hij, ok := w.(http.Hijacker)
	if !ok {
		log.Printf("HTTP Server does not support hijacking\n")
	}

	client, _, err := hij.Hijack()
	if err != nil {
		log.Fatalln("Hijack: ", err)
	}

	// 连接远程
	server, err := net.Dial("tcp", host)
	if err != nil {
		log.Fatalln("Dial host: ", err)
	}
	client.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))

	// 直通双向复制
	go io.Copy(server, client)
	go io.Copy(client, server)
}
