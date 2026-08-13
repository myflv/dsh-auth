package main

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/crypto/bcrypt"
)

//go:embed static/login.html
var loginHTML embed.FS

// 登录页模板（启动时读取一次，每次请求替换占位符）
var loginTemplate []byte

var (
	listen         = flag.String("listen", "127.0.0.1:8080", "监听地址（由 nginx 反代过来）")
	backend        = flag.String("backend", "http://127.0.0.1:13080", "上游地址（frps 的本地端口）")
	hashPass       = flag.String("hash", "", "生成 bcrypt 密码哈希后退出，例: goauth-proxy -hash 'mypass'")
	insecureCookie = flag.Bool("insecure-cookie", false, "本地 http 调试时关闭 cookie 的 Secure 标志")
)

// 认证路径前缀：每次启动随机生成（如 /3f9a2c.../），所有认证资源挂在其下，
// 与后端应用的路由从构造上隔离——应用几乎不可能碰巧存在同名路径，且重启即换
var authPrefix string

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

	loginTemplate, err = loginHTML.ReadFile("static/login.html")
	if err != nil {
		log.Fatal(err)
	}

	initAuth(user, hash)
	authPrefix = newAuthPrefix()

	// 认证资源全部挂在随机前缀下：/<hash>/login、/<hash>/logout
	http.HandleFunc(authPrefix+"login", handleLogin)
	http.HandleFunc(authPrefix+"logout", handleLogout)

	// 其余所有路径：会话有效才反代到后端应用
	http.Handle("/", requireAuth(proxyHandler(backendURL)))

	log.Printf("listening on %s -> %s", *listen, *backend)
	log.Printf("auth portal: http://%s%slogin", *listen, authPrefix)
	log.Fatal(http.ListenAndServe(*listen, nil))
}

// 每次启动生成一个随机路径前缀（48 bit），避免与后端应用的路由冲突
func newAuthPrefix() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}
	return "/" + hex.EncodeToString(b) + "/"
}
