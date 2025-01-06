#!/bin/bash

set -eu

version=0.32.2
cp ../control-tower-ops/manifest.yml opsassets/assets/
cp -R ../control-tower-ops/ops opsassets/assets/
cp ../control-tower-ops/createenv-dependencies-and-cli-versions-aws.json opsassets/assets/
cp ../control-tower-ops/createenv-dependencies-and-cli-versions-gcp.json opsassets/assets/
GO111MODULE=on env GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags "
  -X github.com/EngineerBetter/control-tower/fly.ControlTowerVersion=$version
  -X main.ControlTowerVersion=$version
" -o control-tower-linux-amd64

chmod +x control-tower-linux-amd64

echo "$PWD/control-tower"
