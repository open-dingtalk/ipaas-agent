#!/bin/bash

# 创建bin目录，如果它不存在
mkdir -p bin

# APP Version
APP_VERSION=0.0.1

# APP Name
APP_NAME=agent

# 编译Windows的可执行文件
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$APP_VERSION" -o bin/${APP_NAME}_windows_amd64_${APP_VERSION}.exe

# 编译Linux的可执行文件
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$APP_VERSION" -o bin/${APP_NAME}_linux_amd64_${APP_VERSION}

# 编译mac的可执行文件 (Intel)
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$APP_VERSION" -o bin/${APP_NAME}_darwin_amd64_${APP_VERSION}

# 编译mac的可执行文件 (ARM)
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$APP_VERSION" -o bin/${APP_NAME}_darwin_arm64_${APP_VERSION}