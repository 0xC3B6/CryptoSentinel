#!/bin/bash

# ================= 配置 =================
PROJECT_DIR="/opt/CryptoSentinel"
BACKUP_DIR="/opt/backups/cryptosentinel"
DATE=$(date +%Y%m%d_%H%M%S)
# =======================================

set -e

echo "🚀 开始部署 CryptoSentinel..."

# 1. 创建备份目录
if [ ! -d "$BACKUP_DIR" ]; then
    mkdir -p "$BACKUP_DIR"
fi

# 2. 备份 docker-compose 和 .env 配置
echo "📦 正在备份配置文件..."
tar -czf "$BACKUP_DIR/config_$DATE.tar.gz" \
    -C "$PROJECT_DIR" docker-compose.yml .env 2>/dev/null || true
echo "✅ 备份完成！"

# 3. 拉取最新镜像
echo "📥 正在从 GHCR 拉取最新镜像..."
cd "$PROJECT_DIR"
docker-compose pull cryptosentinel

# 4. 重启容器
echo "🔄 正在平滑重启服务..."
docker-compose up -d

# 5. 清理旧镜像
echo "🧹 正在清理悬空的旧镜像..."
docker image prune -f

echo "✅ CryptoSentinel 部署完成！"
