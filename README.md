# dsh-auth

开箱即用的 **dsh web + 认证网关** 一体镜像：`docker compose up` 就能用，只有一个挂载卷。

```
浏览器 → [nginx(HTTPS) 或 frp] → goauth-proxy(登录认证) → dsh web
```

认证入口挂在**每次启动随机生成**的路径前缀下（如 `/3f9a2c.../login`），与 dsh web 的路由从构造上隔离，重启即换。

## 快速开始

```bash
# 1. 复制并配置环境变量（用户名/密码）
cp .env.example .env
vim .env        # 改 AUTH_PASSWORD

# 2. 启动（首次会自动构建镜像，需要几分钟）
docker compose up -d

# 3. 看日志，拿本次的登录入口
docker logs -f dsh-auth | grep "auth portal"
# → auth portal: http://127.0.0.1:8080/3f9a2c7b1e5d/login
```

浏览器打开 `http://宿主机IP:8080/<hash>/login` 即可登录。

## 环境变量

| 变量 | 必填 | 默认 | 说明 |
|---|---|---|---|
| `AUTH_USER` | ✅ | `admin` | 登录用户名 |
| `AUTH_PASSWORD` | ✅ | 无 | 登录密码（不设启动报错） |
| `DSH_PORT` | | `3080` | dsh web 端口（容器内） |
| `LISTEN_ADDR` | | `0.0.0.0:8080` | 认证代理监听地址（容器内） |
| `DATA_DIR` | | `/data` | dsh 工作目录（容器内路径） |
| `TRUSTED_HOSTS` | ⚠️ 公网时 | 空 | dsh web 的 `/api` 只信任 loopback 和这里声明的来源；走域名访问**必须**填（如 `dsh.example.com`，空格分隔多个） |
| `NPM_REGISTRY` | | npmjs | 构建时 npm 镜像源（国内可设 `https://registry.npmmirror.com`） |

## 挂载卷（唯一的持久化点）

`./data` → `/data`：dsh web 的 HOME 和 cwd 都指向这里，`.dsh/` 配置、会话、日志全在里面。删掉它重来，一切归零。

## 端口与网络

默认 bridge 网络，**只发布到宿主机 localhost**（`127.0.0.1:8080:8080`）：

- 同一台机器上的 frp client 直接连 `127.0.0.1:8080` 即可
- 局域网/公网都走 frp，不要在云上直接暴露 8080
- 想局域网直接访问：ports 改成 `"8080:8080"`

> 不建议 host 网络模式：dsh web 的 3080 会直接占宿主机端口，与你本机正在跑的 dsh web 冲突。真要切：compose 里加 `network_mode: host` 并删除 ports。

## 与 frp + nginx 对接（公网 HTTPS）

```bash
# 宿主机 frpc.toml
[[proxies]]
name = "dsh"
type = "tcp"
localIP = "127.0.0.1"
localPort = 8080
remotePort = 13080        # VPS 上的端口，防火墙只放行给本机
```

```nginx
# VPS nginx（HTTPS 终止在 VPS）
location / {
    proxy_pass http://127.0.0.1:13080;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;   # 认证代理据此开启 cookie Secure
    proxy_set_header Upgrade $http_upgrade;        # dsh web 终端流需要
    proxy_set_header Connection "upgrade";
    proxy_read_timeout 3600s;
}
```

cookie 的 `Secure` 标志自动适配：直连 http 时不加，经 nginx HTTPS 时自动加。

> ⚠️ **公网域名访问必须在 `.env` 里设置 `TRUSTED_HOSTS=你的域名`**：dsh web 的 `/api` 有来源校验（防 DNS rebinding），只信任 loopback 和这里声明的域名，不设的话界面所有 API 请求会被 403 拒绝。设完 `docker compose up -d` 生效。

## 重新构建

```bash
docker compose build        # 改了代码后
docker compose up -d --build
```

## 安全设计

- bcrypt 密码（`AUTH_PASSWORD` 在容器内即时哈希，不落盘）
- 会话：32 字节随机 token + 内存存储，12h 过期，cookie `HttpOnly`/`SameSite=Strict`/`Secure`
- CSRF：登录表单一次性 token
- 登录限速：同 IP 连续失败 5 次锁 5 分钟
- 认证入口随机化：每次启动换前缀，旧入口立即失效

## 目录结构

```
Dockerfile          多阶段：dsh 安装(编译工具链) → 前端构建 → Go 编译 → 精简运行镜像
docker-compose.yml  一键编排
entrypoint.sh       环境变量 → 哈希 → 双进程管理 + 优雅退出
auth/               goauth-proxy 源码（Go + Vue3 认证页）
```
