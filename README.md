# dsh-auth

开箱即用的 **dsh web + 认证网关** 一体镜像：`docker compose up` 就能用，只有一个挂载卷。

```
浏览器 → [nginx(HTTPS) 或 frp] → goauth-proxy(登录认证) → dsh web
```

认证入口是**固定路径** `/login`（以及 `/logout`）：dsh web 服务端只注册了
`/plugins` 和 `/api` 两个前缀路由，其余路径全走 SPA catch-all（实测 POST `/login`
返回 405），固定路径不会与它冲突；且路径不随重启变化，旧标签页永远不会失效。

## 快速开始

```bash
# 1. 创建环境变量文件（用户名/密码）
cat > .env <<'EOF'
AUTH_USER=admin
AUTH_PASSWORD=你的密码
EOF

# 2. 启动（自动从 GHCR 拉取镜像，无需本地构建）
docker compose up -d

# 3. 看日志确认就绪
docker logs -f dsh-auth | grep "auth portal"
# → auth portal: http://127.0.0.1:8080/login
```

浏览器打开 `http://宿主机IP:8080/login` 即可登录。

> **HTTPS 直连（自签）**：容器自带自签 HTTPS（8443 端口），证书只含
> `localhost`/`127.0.0.1`，浏览器访问 `https://127.0.0.1:8443/login`（首次
> 手动信任证书）即可获得安全上下文。证书随每次容器启动重新生成，指纹不固定，
> 信任提示可能反复出现；需要固定证书请走 nginx 正式证书。走域名/公网时请用 nginx 反代（正式证书，
> 见下文 nginx 配置），自签仅供本机/调试。

## 环境变量

| 变量 | 必填 | 默认 | 说明 |
|---|---|---|---|
| `AUTH_USER` | | `admin` | 登录用户名 |
| `AUTH_PASSWORD` | ✅ | 无 | 登录密码（不设启动报错） |

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

## 镜像更新（dsh 发新版自动出新镜像）

镜像由 GitHub Actions 构建并推送到 `ghcr.io/myflv/dsh-auth`，**每天自动检查 npm 上
@deepseek-ai/dsh 的新版本**，有新版就重新构建（`latest` + 版本号双 tag）。
支持 **amd64 + arm64** 双架构（NAS 等 ARM 设备自动拉取对应架构）：

- `docker compose up -d` 每次都会拉取最新镜像（`pull_policy: always`）
- 想立刻检查/强制重建：GitHub 仓库 Actions 页面手动触发 `build-push` workflow
- ⚠️ 若你是 NAS 上手工维护的 compose（没有 `pull_policy: always`），镜像不会
  自动更新，需手动 `docker compose pull && docker compose up -d`，或补上该配置

> ⚠️ **国内网络注意**：`ghcr.io` 在部分地区访问不稳定，拉取失败时请挂代理重试，
> 或把 compose 里的 `image:` 换成其他可访问的 registry。

## 安全设计

- bcrypt 密码（`AUTH_PASSWORD` 在容器内即时哈希，不落盘）
- 会话：32 字节随机 token + 内存存储，12h 过期，cookie `HttpOnly`/`SameSite=Strict`/`Secure`
- CSRF：登录表单一次性 token
- 登录限速：同 IP 连续失败 5 次锁 5 分钟

## 目录结构

```
Dockerfile          多阶段：dsh 安装(编译工具链) → Go 编译 → 精简运行镜像（登录页已嵌入二进制）
docker-compose.yml  一键编排
entrypoint.sh       环境变量 → 哈希 → 双进程管理 + 优雅退出
auth/               goauth-proxy 源码（Go + 内嵌静态登录页）
```
