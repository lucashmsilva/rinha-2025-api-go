#!/bin/sh
set -e
cd "$(dirname "$0")"

if [ "$BUILD_WITH_RACE_DETECTION" = "1" ];
then
  echo "building with race detection..."
  CGO_ENABLED=1
  go build -race -o /usr/local/bin/main ../cmd/api/main.go
else
  go build -o /usr/local/bin/main ../cmd/api/main.go
fi

echo "finished build"
