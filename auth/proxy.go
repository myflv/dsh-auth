package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// proxyHandler 把请求转发到 frps 的本地端口（127.0.0.1:13080）
// 标准库 ReverseProxy 自带 WebSocket 转发支持，dsh web 的终端流没问题
func proxyHandler(backend *url.URL) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(backend)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error: %v", err)
		http.Error(w, "上游不可用（frp 隧道可能没连上）", http.StatusBadGateway)
	}
	return proxy
}
