#!/bin/bash

set -e

go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct

echo "🛠️ 构建 git-analyst..."
go mod tidy
go build -o git-analyst


