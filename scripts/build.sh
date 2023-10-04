#!/bin/sh

set -e

GOOS=linux GOARCH=amd64 go build -o "infoniqa" ./cmd/infoniqa
GOOS=windows GOARCH=amd64 go build -o "infoniqa.exe" ./cmd/infoniqa

echo "Build finished"