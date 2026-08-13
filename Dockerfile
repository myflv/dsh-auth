# ========== 阶段 1：安装 dsh（node-pty 原生模块需要编译工具链） ==========
FROM node:26-bookworm-slim AS build

ARG NPM_REGISTRY=https://registry.npmjs.org
# dsh 版本：CI 每次检查 npm 新版本后以 --build-arg 覆盖
ARG DSH_VERSION=0.1.0-rc.6
RUN apt-get update \
    && apt-get install -y --no-install-recommends python3 make g++ \
    && rm -rf /var/lib/apt/lists/*

# dsh（版本由 DSH_VERSION 控制）
RUN npm install -g --no-audit --no-fund --registry=${NPM_REGISTRY} @deepseek-ai/dsh@${DSH_VERSION}

# ========== 阶段 2：编译 goauth-proxy（登录页为单文件静态 HTML，直接嵌入） ==========
FROM golang:1.26 AS auth-build
# 国内网络 proxy.golang.org 不可达，用 goproxy.cn 镜像
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /src
COPY auth/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/goauth-proxy .

# ========== 阶段 3：运行时（精简镜像，无需编译工具） ==========
FROM node:26-bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# dsh 及全局依赖（含编译好的原生模块）
COPY --from=build /usr/local/lib/node_modules /usr/local/lib/node_modules
# npm 的 bin 是符号链接，Docker COPY 会解引用成普通文件，导致 ESM 依赖解析失败；
# 改为显式创建符号链接，Node 按真实路径解析依赖
RUN ln -s /usr/local/lib/node_modules/@deepseek-ai/dsh/lib/bin.js /usr/local/bin/dsh
# 认证代理（静态二进制）
COPY --from=auth-build /out/goauth-proxy /usr/local/bin/goauth-proxy

COPY --chmod=+x entrypoint.sh /entrypoint.sh

VOLUME /data
EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
