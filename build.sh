#!/bin/sh

GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -buildid= -X main.version=$(git describe --always --dirty) -X 'main.build=$(date)'" -trimpath -o bin/clipScript_win64.exe
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -buildid= -X main.version=$(git describe --always --dirty) -X 'main.build=$(date)'" -trimpath -o bin/clipScript_linux64
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -buildid= -X main.version=$(git describe --always --dirty) -X 'main.build=$(date)'" -trimpath -o bin/clipScript_linuxArm64
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -buildid= -X main.version=$(git describe --always --dirty) -X 'main.build=$(date)'" -trimpath -o bin/clipScript_osx64
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -buildid= -X main.version=$(git describe --always --dirty) -X 'main.build=$(date)'" -trimpath -o bin/clipScript_osxArm64
