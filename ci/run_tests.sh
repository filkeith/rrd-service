#!/usr/bin/env bash
set -e

TARGET_OS=${TARGET_OS:=$GOOS}
TARGET_ARCH=${TARGET_ARCH:=$GOARCH}

mkdir -p ./coverage
GOOS=${TARGET_OS} GOARCH=${TARGET_ARCH} go test --timeout 120s --cover -coverprofile=./coverage/coverage.out ./...
go tool cover -func ./coverage/coverage.out | tail -n 1