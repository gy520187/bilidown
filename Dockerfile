# ============================================================
#  支持国内镜像加速的 Dockerfile
#  使用 REGISTRY_MIRROR 构建参数指定镜像加速地址前缀
#  默认为空 (直接使用 Docker Hub)
#  示例: docker build --build-arg REGISTRY_MIRROR=docker.1ms.run/
# ============================================================

ARG REGISTRY_MIRROR=""

# ============================================================
#  Stage 1: 构建前端 (Vite + VanJS + Bootstrap)
# ============================================================
FROM ${REGISTRY_MIRROR}node:20-slim AS frontend-builder

WORKDIR /app

# 启用 pnpm
RUN corepack enable

# 先复制依赖清单以利用 Docker 层缓存
COPY client/package.json client/pnpm-lock.yaml ./client/
RUN cd client && pnpm install --frozen-lockfile

# 复制前端源码并构建（输出到 ../server/static）
COPY client/ ./client/
RUN cd client && pnpm build

# ============================================================
#  Stage 2: 构建后端 (Go + pure-Go SQLite, 无需 CGO)
# ============================================================
FROM ${REGISTRY_MIRROR}golang:1.23-alpine AS backend-builder

WORKDIR /app

# 复制后端源码
COPY server/ ./server/

# 将前端构建产物复制到 server/static 目录
COPY --from=frontend-builder /app/server/static/ ./server/static/

# 使用 dockerheadless 构建标签跳过 systray (GUI 依赖)
# CGO_ENABLED=0: modernc.org/sqlite 是纯 Go 实现, 不需要 CGO
# 同时 systray 被 build tag 排除, 也不需要 C 库
# GOPROXY 使用国内加速源
RUN cd server && \
    CGO_ENABLED=0 GOPROXY=https://goproxy.cn,direct go build -tags dockerheadless -o bilidown .

# ============================================================
#  Stage 3: 运行时镜像 (最小化, 仅含二进制 + FFmpeg + 静态文件)
# ============================================================
FROM ${REGISTRY_MIRROR}debian:bookworm-slim

# 替换为国内 Debian 镜像源, 加速 apt 下载 (FFmpeg 依赖很多)
RUN sed -i 's|deb.debian.org|mirrors.aliyun.com|g' /etc/apt/sources.list.d/debian.sources

# 安装 FFmpeg (视频合并必需) 和证书
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ffmpeg \
        ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 复制编译好的二进制
COPY --from=backend-builder /app/server/bilidown ./

# 复制前端静态文件
COPY --from=frontend-builder /app/server/static/ ./static/

# 创建数据目录和下载目录, 并将 data.db 软链接到 data/ 目录以支持持久化
RUN mkdir -p data download && ln -sf data/data.db data.db

# 暴露 HTTP 服务端口
EXPOSE 8098

# 数据持久化挂载点
VOLUME ["/app/data", "/app/download"]

# 启动命令
CMD ["./bilidown"]
