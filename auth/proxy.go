package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// proxyHandler 把请求转发到 dsh web（容器内 127.0.0.1:3080）
// 标准库 ReverseProxy 自带 WebSocket 转发支持，dsh web 的终端流没问题
func proxyHandler(backend *url.URL) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(backend)
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		// dsh web 的 /api browser-trust fence 只信任 loopback 主机（或
		// --trusted-host 声明，安装版 rc.6 该参数实测未生效）。我们的容器
		// 里 3080 不对外暴露，认证由本代理负责，fence 无实际防御价值——
		// 直接把 Host/Origin 改写为回环地址，fence 恒通过
		req.Host = backend.Host
		if origin := req.Header.Get("Origin"); origin != "" {
			req.Header.Set("Origin", "http://"+backend.Host)
		}
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error: %v", err)
		http.Error(w, "上游不可用（dsh web 可能没起来）", http.StatusBadGateway)
	}
	return proxy
}
