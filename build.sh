#!/bin/bash

# Minlink Docker 镜像构建脚本
# 使用方式: ./build.sh [版本号]
# 示例: ./build.sh 1.0.0

# 默认版本号
VERSION=${1:-1.0.1}

# 镜像名称
IMAGE_NAME="minlink"

# 构建镜像
echo "Building Docker image: ${IMAGE_NAME}:${VERSION}"
docker build -t ${IMAGE_NAME}:${VERSION} -t ${IMAGE_NAME}:latest .

echo ""
echo "Build completed!"
echo "Image: ${IMAGE_NAME}:${VERSION}"
echo "Image: ${IMAGE_NAME}:latest"
echo ""
echo "To run with docker-compose:"
echo "  docker-compose up -d"