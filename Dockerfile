# syntax=docker/dockerfile:1

# ========== 阶段 1：安装 dsh + 构建认证页前端 ==========
FROM node:26-bookworm-slim AS build

ARG NPM_REGISTRY=https://registry.npmjs.org
# dsh 依赖的 node-pty 是原生模块，安装时需要编译工具链
RUN apt-get update \
    && apt-get install -y --no-install-recommends python3 make g++ \
    && rm -rf /var/lib/apt/lists/*

# dsh（锁版本，保证可复现）
RUN npm install -g --no-audit --no-fund --registry=${NPM_REGISTRY} @deepseek-ai/dsh@0.1.0-rc.6

# 认证页前端（Vue3 + Vite）
WORKDIR /src/frontend
COPY auth/frontend/package.json auth/frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund --registry=${NPM_REGISTRY}
COPY auth/frontend/ ./
RUN npm run build

# ========== 阶段 2：编译 goauth-proxy（前端产物嵌入二进制） ==========
FROM golang:1.26 AS auth-build
WORKDIR /src
COPY auth/ ./
COPY --from=build /src/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/goauth-proxy .

# ========== 阶段 3：运行时（精简镜像，无需编译工具） ==========
FROM node:26-bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends curl \
    && rm -rf /var/lib/apt/lists/*

# dsh 及全局依赖（含编译好的原生模块）
COPY --from=build /usr/local/lib/node_modules /usr/local/lib/node_modules
# npm 的 bin 是符号链接，Docker COPY 会解引用成普通文件，导致 ESM 依赖解析失败；
# 改为显式创建符号链接，Node 按真实路径解析依赖
RUN ln -s /usr/local/lib/node_modules/@deepseek-ai/dsh/lib/bin.js /usr/local/bin/dsh
# 认证代理（静态二进制）
COPY --from=auth-build /out/goauth-proxy /usr/local/bin/goauth-proxy

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENV DATA_DIR=/data \
    DSH_HOST=127.0.0.1 \
    DSH_PORT=3080 \
    LISTEN_ADDR=0.0.0.0:8080

VOLUME /data
EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
