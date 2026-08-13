package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	cookieName  = "dsh_session"    // 登录会话 cookie
	sessionTTL  = 12 * time.Hour   // 会话有效期
	csrfTTL     = 10 * time.Minute // 登录表单 token 有效期
	maxFailures = 5                // 连续失败次数
	lockTime    = 5 * time.Minute  // 锁定时长
)

var (
	authUser string
	authHash string
)

func initAuth(user, hash string) {
	authUser = user
	authHash = hash
}

// ---------- session（内存存储，重启即失效，个人工具够用） ----------

type session struct {
	token   string
	expires time.Time
}

var (
	sessMu   sync.Mutex
	sessions = map[string]session{}
)

func newSession() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}
	token := hex.EncodeToString(b)

	sessMu.Lock()
	defer sessMu.Unlock()
	now := time.Now()
	for k, s := range sessions { // 顺手清一波过期 session
		if now.After(s.expires) {
			delete(sessions, k)
		}
	}
	sessions[token] = session{token: token, expires: now.Add(sessionTTL)}
	return token
}

func validSession(token string) bool {
	if token == "" {
		return false
	}
	sessMu.Lock()
	defer sessMu.Unlock()
	s, ok := sessions[token]
	if !ok {
		return false
	}
	if time.Now().After(s.expires) {
		delete(sessions, token)
		return false
	}
	return true
}

func deleteSession(token string) {
	sessMu.Lock()
	defer sessMu.Unlock()
	delete(sessions, token)
}

// ---------- CSRF（登录表单一次性 token） ----------

var (
	csrfMu     sync.Mutex
	csrfTokens = map[string]time.Time{}
)

func newCSRF() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}
	token := hex.EncodeToString(b)

	csrfMu.Lock()
	defer csrfMu.Unlock()
	now := time.Now()
	for k, t := range csrfTokens {
		if now.After(t) {
			delete(csrfTokens, k)
		}
	}
	csrfTokens[token] = now.Add(csrfTTL)
	return token
}

// 校验并立即消费（一次性）
func checkCSRF(token string) bool {
	csrfMu.Lock()
	defer csrfMu.Unlock()
	t, ok := csrfTokens[token]
	if !ok {
		return false
	}
	delete(csrfTokens, token)
	return time.Now().Before(t)
}

// ---------- 登录失败限速 ----------

type attempt struct {
	count int
	until time.Time
}

var failures sync.Map // ip -> *attempt

// nginx 反代时从 X-Forwarded-For 取第一个，否则用 RemoteAddr
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// limited 只读检查：锁定期内返回 true。注意不能在这里删除计数，
// 否则每次请求都从 0 开始，永远触发不了锁定
func limited(ip string) bool {
	if v, ok := failures.Load(ip); ok {
		a := v.(*attempt)
		if time.Now().Before(a.until) {
			return true
		}
	}
	return false
}

func recordFailure(ip string) {
	v, _ := failures.LoadOrStore(ip, &attempt{})
	a := v.(*attempt)
	if !a.until.IsZero() && time.Now().After(a.until) {
		// 上次锁定已过期，重新计数
		a.until = time.Time{}
		a.count = 0
	}
	a.count++
	if a.count >= maxFailures {
		a.until = time.Now().Add(lockTime)
		a.count = 0
	}
}

// ---------- handlers ----------

func handleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		token := newCSRF()
		// 登录页必须禁缓存：浏览器缓存旧 HTML 会带着已被消费的一次性 token，
		// 导致刷新后依旧 CSRF 失败
		w.Header().Set("Cache-Control", "no-store")
		// 把构建好的 Vue 登录页里的 {{CSRF}} 替换为一次性 token
		page := strings.ReplaceAll(string(frontendIndex), "{{CSRF}}", token)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(page))

	case http.MethodPost:
		ip := clientIP(r)
		if limited(ip) {
			http.Error(w, "尝试次数过多，请 5 分钟后再试", http.StatusTooManyRequests)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "参数错误", http.StatusBadRequest)
			return
		}
		if !checkCSRF(r.FormValue("csrf")) {
			http.Error(w, "CSRF 校验失败，请刷新页面重试", http.StatusBadRequest)
			return
		}
		if r.FormValue("username") != authUser ||
			bcrypt.CompareHashAndPassword([]byte(authHash), []byte(r.FormValue("password"))) != nil {
			recordFailure(ip)
			http.Error(w, "用户名或密码错误", http.StatusUnauthorized)
			return
		}

		failures.Delete(ip)
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    newSession(),
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			// 直连 http（如容器裸端口）时自动关闭 Secure，否则浏览器不存 cookie；
			// 经过 nginx HTTPS（X-Forwarded-Proto: https）时保持 Secure
			Secure: !*insecureCookie && (r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")),
			MaxAge: int(sessionTTL.Seconds()),
		})
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(cookieName); err == nil {
		deleteSession(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, authPrefix+"login", http.StatusFound)
}

// 认证中间件：有有效会话才放行，否则跳登录页
func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie(cookieName); err == nil && validSession(c.Value) {
			next.ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, authPrefix+"login", http.StatusFound)
	})
}
