package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)


func StartHTTPProxy(addr, port string) {
	proxy := &HTTPProxy{
		Addr: addr,
		Port: port,
	}

	log.Printf("HTTPProxy is runing on %s:%s \n", proxy.Addr, proxy.Port)
	log.Fatalln(http.ListenAndServe(proxy.Addr + ":" + proxy.Port, proxy))
}


// 仅支持HTTP单连接的代理服务器
type HTTPProxy struct {
	Addr string
	Port string
}

// 正常客户端与服务器http连接具体连接依赖底层TCP连接，目标服务器地址也是给在了TCP层，
// 因此http层不知道也不需要知道服务器地址，这就是为什么在http请求报文中请求行
// 通常只是 METHOD URL VERSION， URL处填类似 /index.html 的资源路径，
// 而不带上服务器地址作为前缀(https需要加上服务器地址，这里不提)
//
// 因此对于HTTP代理服务器而言，它工作于http层，并不知道真正的服务器地址
// 这个地址需要客户端在请求行中加上，也就是现在URL要填写 127.0.0.1:8000/index.html
// 当我们使用浏览器设置代理服务器进行网站访问时，浏览器自动把我们把这个真正的服务器地址给加上了
func (proxy *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// w 是当前代理用来返回客户端数据的，r是客户端的请求

	fmt.Printf("Received request: %s %s %s\n", r.Method, r.Host, r.RemoteAddr)

	transport := http.DefaultTransport

	// 1. 复制原请求，并加以修改
	outReq := new(http.Request)
	*outReq = *r	// 复制内容

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

	// 2. 将修改后的请求转发出去，等待服务器回应
	resp, err := transport.RoundTrip(outReq)
	if err != nil {
		// 出错，那么说明这台代理服务器无法连通到，那么在w的状态头中写上BadGateway
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 3. 将真正服务器的状态头给写到w的状态头去(就是写回给客户端了)
	for key, value := range resp.Header {
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}

	// 再把服务器返回的状态码和真正内容写入到w，返回给客户端
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
