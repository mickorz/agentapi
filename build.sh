#!/bin/bash
# AgentAPI 完整构建脚本（包含 Chat UI）
# 使用方法: ./build.sh

set -e

echo "=== Step 1: 安装前端依赖 ==="
cd chat
bun install

echo "=== Step 2: 构建前端 Chat UI ==="
MSYS_NO_PATHCONV=1 NEXT_PUBLIC_BASE_PATH="/magic-base-path-placeholder" bun run build

echo "=== Step 3: 复制前端构建产物 ==="
cd ..
rm -rf lib/httpapi/chat
mkdir -p lib/httpapi/chat
cp -r chat/out/. lib/httpapi/chat/

echo "=== Step 4: 编译 Go 项目 ==="
GOPROXY=https://goproxy.cn,direct CGO_ENABLED=0 go build -o out/agentapi.exe main.go

echo ""
echo "[成功] 构建完成！"
echo "输出文件: out/agentapi.exe"
ls -lh out/agentapi.exe
