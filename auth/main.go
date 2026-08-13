package main

import (
	_ "embed" // go:embed 所需
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/crypto/bcrypt"
)

//go:embed static/login.html
var loginTemplate string // 登录页模板，启动时由编译器嵌入

var (
	listen         = flag.String("listen", "127.0.0.1:8080", "监听地址（由 nginx 反代过来）")
	backend        = flag.String("backend", "http://127.0.0.1:13080", "上游地址（frps 的本地端口）")
	hashPass       = flag.String("hash", "", "生成 bcrypt 密码哈希后退出，例: goauth-proxy -hash 'mypass'")
	insecureCookie = flag.Bool("insecure-cookie", false, "本地 http 调试时关闭 cookie 的 Secure 标志")
)

// 认证入口固定路径：/login、/logout。
// dsh web 服务端只注册了 /plugins 和 /api 两个前缀路由，其余全走 SPA
// catch-all（实测 POST /login 返回 405），固定路径不会冲突；
// 且路径不再随重启变化，旧标签页永远不会失效
func main() {
	flag.Parse()

	// 生成密码哈希模式
	if *hashPass != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(*hashPass), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(h))
		return
	}

	user := os.Getenv("AUTH_USER")
	hash := os.Getenv("AUTH_HASH")
	if user == "" || hash == "" {
		log.Fatal("需要环境变量 AUTH_USER 和 AUTH_HASH（先用 -hash 生成）")
	}

	backendURL, err := url.Parse(*backend)
	if err != nil {
		log.Fatal(err)
	}

	authUser, authHash = user, hash

	http.HandleFunc(loginPath, handleLogin)
	http.HandleFunc(logoutPath, handleLogout)

	// 其余所有路径：会话有效才反代到后端应用
	http.Handle(homePath, requireAuth(proxyHandler(backendURL)))

	log.Printf("listening on %s -> %s", *listen, *backend)
	log.Printf("auth portal: http://%s%s", *listen, loginPath)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
