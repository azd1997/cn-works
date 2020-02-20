package main

import (
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// http提供了专门的反向代理结构： httputil.ReverseProxy

// 直接取个别名
type HTTPReverseProxy = httputil.ReverseProxy

// 填写的是代理的若干服务器节点的地址
func NewHTTPReverseProxy(targets []*url.URL) *HTTPReverseProxy {
	n := len(targets)
	director := func(r *http.Request) {
		target := targets[rand.Intn(n)]		// 随机选择一个服务器节点
		// 修改request的请求行
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host
		r.URL.Path = target.Path
	}
	return &HTTPReverseProxy{Director:director}
}
