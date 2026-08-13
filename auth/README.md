# goauth-proxy

Go 写的"认证 + 反向代理"，给 frp 穿透出来的 dsh web（或任何内网 Web 服务）加一道登录页。

架构：`nginx(443/HTTPS) -> goauth-proxy(认证) -> frps 本地端口 -> frp 隧道 -> 内网服务`

## 防冲突设计：随机路径前缀

认证资源不占用固定路径，而是挂在**每次启动随机生成**的前缀下（如 `/3f9a2c7b1e5d/login`）：

- `/<hash>/login`    登录页（Vue3 + Vite 构建，Miuix 风格）
- `/<hash>/logout`   登出
- `/<hash>/assets/*` 前端资源（Vite 相对路径构建，全部落在此前缀下）
- 其余所有路径 → 会话校验通过才反代到后端应用

后端应用（如 dsh web 的 SPA）与认证页从构造上隔离：应用几乎不可能碰巧存在同名路径，且前缀每次重启都会更换。未登录访问任意路径会 302 到当前前缀的登录页。

## 构建与运行

```bash
# 1. 构建前端（Vue3 + Vite -> frontend/dist，嵌入 Go 二进制）
cd frontend && npm install && npm run build && cd ..

# 2. 构建 Go 二进制
go build -o goauth-proxy .

# 3. 生成密码哈希
./goauth-proxy -hash '你的密码'

# 4. 启动（把上一步输出的哈希填进 AUTH_HASH）
AUTH_USER=xin AUTH_HASH='$2a$10$...' ./goauth-proxy

# 本地 http 调试时（没有 nginx 的 HTTPS）加 -insecure-cookie
AUTH_USER=xin AUTH_HASH='$2a$10$...' ./goauth-proxy -insecure-cookie
```

启动日志会打印本次的登录入口，如 `auth portal: http://127.0.0.1:8080/3f9a2c7b1e5d/login`。

- `-listen` 默认 `127.0.0.1:8080`（只绑本机，nginx 转过来）
- `-backend` 默认 `http://127.0.0.1:13080`（frps 的本地端口，按你的 frp 配置改）

## nginx 配置（VPS 上）

```nginx
server {
    listen 443 ssl;
    server_name dsh.example.com;

    ssl_certificate     /etc/letsencrypt/live/dsh.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/dsh.example.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;   # goauth-proxy
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        # WebSocket（dsh web 终端流需要）
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 3600s;
    }
}
```

## frp 侧注意

**frps 的 remote port（13080）不要对公网开放**，否则别人绕过登录页直接连 frp 端口。
在 VPS 防火墙/云安全组里只放行 80/443，封掉 13080。

## 安全设计

- 密码：bcrypt 存储与校验（`-hash` 生成）
- 会话：`crypto/rand` 32 字节 token + 内存存储，12h 过期，cookie 带 `HttpOnly`/`SameSite=Strict`/`Secure`
- CSRF：登录表单一次性 token（服务端注入到 Vue 页面）
- 暴力破解：同一 IP 连续失败 5 次锁 5 分钟（取 `X-Forwarded-For` 第一跳）
- 认证入口随机化：每次启动生成新前缀，重启即换，旧入口立即失效
- 重启进程即清空会话，需要重新登录

## 文件结构

```
main.go               入口、随机前缀生成、前端资源挂载
auth.go               登录/登出/session/CSRF/限速
proxy.go              反向代理（标准库，含 WebSocket）
frontend/             Vue3 + Vite 登录页（Miuix 风格）
  src/App.vue         登录卡片、表单、错误提示
  dist/               构建产物（go:embed 嵌入二进制）
```

## 前端开发（改样式时用）

```bash
cd frontend
npm run dev        # Vite dev server，改样式实时预览
npm run build      # 构建后需要重新 go build 才会生效
```
