#!/bin/sh
set -e

# ===== 必填环境变量 =====
: "${AUTH_USER:?请设置 AUTH_USER 环境变量（登录用户名）}"
: "${AUTH_PASSWORD:?请设置 AUTH_PASSWORD 环境变量（登录密码）}"

# ===== 可选环境变量 =====
# dsh web 的 /api browser-trust fence：非 loopback 访问（如公网域名）时必须声明，
# 空格分隔多个，如 TRUSTED_HOSTS="dsh.example.com"
TRUSTED_HOSTS="${TRUSTED_HOSTS:-}"

# 内部接线（固定值，与 compose 的 ports/volumes 配套，改动需同步两处）
DATA_DIR=/data
DSH_HOST=127.0.0.1
DSH_PORT=3080
LISTEN_ADDR=0.0.0.0:8080

# dsh 工作目录：HOME 和 cwd 都指向挂载卷，dsh 的所有数据（.dsh/）落在里面
mkdir -p "$DATA_DIR"
export HOME="$DATA_DIR"
cd "$DATA_DIR"
echo "[dsh-auth] 数据目录: $DATA_DIR"

# 密码 -> bcrypt 哈希
AUTH_HASH="$(goauth-proxy -hash "$AUTH_PASSWORD")"
echo "[dsh-auth] 认证用户: $AUTH_USER"

# 启动 dsh web
echo "[dsh-auth] 启动 dsh web ($DSH_HOST:$DSH_PORT) ..."
dsh web --host "$DSH_HOST" --port "$DSH_PORT" &
DSH_PID=$!

# 启动认证代理（前台）
echo "[dsh-auth] 启动认证代理，监听 $LISTEN_ADDR"
AUTH_USER="$AUTH_USER" AUTH_HASH="$AUTH_HASH" \
    goauth-proxy -listen "$LISTEN_ADDR" -backend "http://$DSH_HOST:$DSH_PORT" &
AUTH_PID=$!

# docker stop 时优雅退出
trap 'echo "[dsh-auth] 收到退出信号，停止服务"; kill "$DSH_PID" "$AUTH_PID" 2>/dev/null; wait' TERM INT
wait "$AUTH_PID"
