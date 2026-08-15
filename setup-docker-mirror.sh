#!/bin/bash
# ============================================================
#  Docker 镜像加速配置修复脚本
#  解决 hub-mirror.c.163.com 等失效镜像源导致的构建失败
# ============================================================

set -e

DAEMON_JSON="/etc/docker/daemon.json"
BACKUP_FILE="/etc/docker/daemon.json.bak.$(date +%Y%m%d%H%M%S)"

echo "============================================"
echo "  Docker 镜像加速配置修复工具"
echo "============================================"
echo ""

# 检查权限
if [ "$EUID" -ne 0 ]; then
    echo "请使用 sudo 运行此脚本"
    exit 1
fi

# 备份原有配置
if [ -f "$DAEMON_JSON" ]; then
    cp "$DAEMON_JSON" "$BACKUP_FILE"
    echo "已备份原配置到: $BACKUP_FILE"
else
    echo "未找到现有 daemon.json, 将创建新配置"
    mkdir -p /etc/docker
fi

# 2026年8月实测可用的镜像加速源
cat > "$DAEMON_JSON" <<'EOF'
{
  "registry-mirrors": [
    "https://docker.1ms.run",
    "https://dockerproxy.net",
    "https://docker.m.daocloud.io",
    "https://docker.xuanyuan.me"
  ]
}
EOF

echo "已写入新的镜像加速配置:"
cat "$DAEMON_JSON"
echo ""

# 重启 Docker 服务
echo "正在重启 Docker 服务..."
systemctl daemon-reload
systemctl restart docker

echo ""
echo "Docker 镜像加速配置已更新并重启服务完成。"
echo ""
echo "可用的镜像加速源 (2026年8月实测):"
echo "  1. https://docker.1ms.run         (毫秒镜像)"
echo "  2. https://dockerproxy.net        (DockerProxy)"
echo "  3. https://docker.m.daocloud.io   (DaoCloud)"
echo "  4. https://docker.xuanyuan.me     (轩辕镜像)"
echo ""
echo "验证命令: docker info | grep -A5 'Registry Mirrors'"
echo ""
echo "现在可以重新构建: docker compose up -d --build"
