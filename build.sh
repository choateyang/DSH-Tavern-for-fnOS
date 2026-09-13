#!/bin/bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
cd "$ROOT_DIR"

echo "[INFO] 正在安装并构建 Vue 3 前端静态资产..."
(cd frontend && npm ci && npm run build)

test -d tavern-source
echo "[INFO] 正在打包内置 DSH Tavern 源码..."
tar -czf fnpack/app/dsh-tavern.tar.gz -C tavern-source .

rm -rf fnpack-x86 fnpack-arm
cp -a fnpack fnpack-x86
cp -a fnpack fnpack-arm

rm -f fnpack-x86/app/bin/skill/trim-cli-linux-arm64
rm -f fnpack-arm/app/bin/skill/trim-cli-linux-x64

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o fnpack-x86/app/bin/dsh.tavern .
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o fnpack-arm/app/bin/dsh.tavern .
sed -i 's/^platform[[:space:]]*=[[:space:]]*x86[[:space:]]*$/platform              = arm/' fnpack-arm/manifest

chmod +x fnpack-x86/fnpack-1.2.3-linux-amd64 fnpack-arm/fnpack-1.2.3-linux-amd64
(cd fnpack-x86 && ./fnpack-1.2.3-linux-amd64 build)
(cd fnpack-arm && ./fnpack-1.2.3-linux-amd64 build)
cp fnpack-x86/dsh.tavern.fpk dsh.tavern-x86.fpk
cp fnpack-arm/dsh.tavern.fpk dsh.tavern-arm.fpk
echo "[INFO] 已生成 dsh.tavern-x86.fpk 与 dsh.tavern-arm.fpk。"
