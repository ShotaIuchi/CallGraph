#!/bin/sh

go mod download

GOTARGET="darwin/amd64 linux/amd64 windows/amd64"

for target in $GOTARGET; do
  GOOS=${target%/*}
  GOARCH=${target#*/}
  DIR_NAME=${GOOS}-${GOARCH}

  if [ "$GOOS" = "windows" ]; then
    OUTPUT_NAME="callgraph.exe"
  else
    OUTPUT_NAME="callgraph"
  fi

  mkdir -p ./${DIR_NAME}
  GOOS=$GOOS GOARCH=$GOARCH go build -o ./${DIR_NAME}/${OUTPUT_NAME} ./
done
