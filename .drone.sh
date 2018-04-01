#!/bin/sh

set -ex

# compile the main binary
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.build=${DRONE_BUILD_NUMBER}" -a -tags netgo -o release/linux/amd64/drone-stack megpoid.xyz/go/drone-stack/cmd/drone-stack
