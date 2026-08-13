#!/bin/sh
set -e

# ===== 必填环境变量 =====
: "${AUTH_USER:?请设置 AUTH_USER 环境变量（登录用户名）}"
: "${AUTH_PASSWORD:?请设置 AUTH_PASSWORD 环境变量（登录密码）}"

# ===== 可选环境变量 =====
DATA_DIR="${DATA_DIR:-/data}"
DSH_HOST="${DSH_HOST:-127.0.0.1}"
DSH_PORT="${DSH_PORT:-3080}"
LISTEN_ADDR="${LISTEN_ADDR:-0.0.0.0:8080}"
# dsh web 的 /api browser-trust fence：非 loopback 访问（如公网域名）时必须声明，
# 空格分隔多个，如 TRUSTED_HOSTS="dsh.example.com dsh2.example.com"
TRUSTED_HOSTS="${TRUSTED_HOSTS:-}"

# dsh 工作目录：HOME 和 cwd 都指向挂载卷，dsh 的所有数据（.dsh/）落在里面
mkdir -p "$DATA_DIR"
export HOME="$DATA_DIR"
cd "$DATA_DIR"
echo "[dsh-auth] 数据目录: $DATA_DIR"

# 密码 -> bcrypt 哈希
AUTH_HASH="$(goauth-proxy -hash "$AUTH_PASSWORD")"
echo "[dsh-auth] 认证用户: $AUTH_USER"

# 启动 dsh web
if [ -n "$TRUSTED_HOSTS" ]; then
    echo "[dsh-auth] dsh web ($DSH_HOST:$DSH_PORT)，trusted hosts: $TRUSTED_HOSTS"
    # shellcheck disable=SC2086 # 有意按空格拆分为多个 --trusted-host 参数
    dsh web --host "$DSH_HOST" --port "$DSH_PORT" --trusted-host $TRUSTED_HOSTS &
else
    echo "[dsh-auth] 启动 dsh web ($DSH_HOST:$DSH_PORT) ..."
    dsh web --host "$DSH_HOST" --port "$DSH_PORT" &
fi
DSH_PID=$!

# 等待 dsh web 就绪
for i in $(seq 1 60); do
    if curl -sf "http://$DSH_HOST:$DSH_PORT/" >/dev/null 2>&1; then
        echo "[dsh-auth] dsh web 就绪"
        break
    fi
    if ! kill -0 "$DSH_PID" 2>/dev/null; then
        echo "[dsh-auth] 错误: dsh web 启动失败" >&2
        exit 1
    fi
    sleep 1
done

# 启动认证代理（前台）
echo "[dsh-auth] 启动认证代理，监听 $LISTEN_ADDR"
AUTH_USER="$AUTH_USER" AUTH_HASH="$AUTH_HASH" \
    goauth-proxy -listen "$LISTEN_ADDR" -backend "http://$DSH_HOST:$DSH_PORT" &
AUTH_PID=$!

# docker stop 时优雅退出
trap 'echo "[dsh-auth] 收到退出信号，停止服务"; kill "$DSH_PID" "$AUTH_PID" 2>/dev/null; wait' TERM INT
wait "$AUTH_PID"
