#!/usr/bin/env bash
# GoDevKit 打包脚本
# 用法:
#   ./build.sh cli              打包 TUI 版（当前平台）
#   ./build.sh cli-all          打包 TUI 版（跨平台）
#   ./build.sh gui              打包 GUI 版（fyne package，当前平台）
#   ./build.sh all              cli-all + gui
set -euo pipefail

cd "$(dirname "$0")"

VERSION="$(git describe --tags --always 2>/dev/null || echo dev)"
OUT=dist
LDFLAGS="-s -w"

targets=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

build_cli_one() {
    local goos="$1" goarch="$2" name="gokit"
    [[ "$goos" == "windows" ]] && name="gokit.exe"
    echo "==> 编译 TUI: $goos/$goarch"
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
        go build -trimpath -ldflags "$LDFLAGS" -o "$OUT/${goos}-${goarch}/$name" ./cmd/gokit
}

build_cli() {
    echo "==> 编译 TUI: $(go env GOOS)/$(go env GOARCH)"
    CGO_ENABLED=0 go build -trimpath -ldflags "$LDFLAGS" -o "$OUT/gokit" ./cmd/gokit
    echo "==> 完成: $OUT/gokit"
}

build_cli_all() {
    for t in "${targets[@]}"; do
        build_cli_one "${t%%/*}" "${t##*/}"
    done
    echo "==> 压缩包"
    for t in "${targets[@]}"; do
        dir="${t//\//-}"
        (cd "$OUT" && tar -czf "$dir.tar.gz" "$dir" && rm -rf "$dir")
    done
    ls -lh "$OUT"/*.tar.gz
}

build_gui() {
    echo "==> 打包 GUI (fyne package)"
    local os="$(go env GOOS)"
    local ext="tar.xz"
    [[ "$os" == "darwin" ]] && ext="app.tar.xz"
    fyne package -os "$os" -icon res/logo.png -name GoDevKit \
        -appVersion "$VERSION" -source main.go
    echo "==> 完成: GoDevKit.$ext"
}

mkdir -p "$OUT"

case "${1:-cli}" in
    cli)      build_cli ;;
    cli-all)  build_cli_all ;;
    gui)      build_gui ;;
    all)      build_cli_all; build_gui ;;
    *)        echo "用法: $0 {cli|cli-all|gui|all}"; exit 1 ;;
esac
