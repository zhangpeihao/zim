.PHONY: build

PACKAGES = $(shell go list ./... | grep -v /vendor/)
VERSION=$(shell git describe --match 'v[0-9]*' --dirty --always)
GO_LDFLAGS=-ldflags "-X github.com/zhangpeihao/zim/pkg/version.VersionDev=$(VERSION)"

all: build_static

test:
	go test -cover $(PACKAGES)

# build the release files
build: build_static build_cross build_tar

build_static:
	go build ${GO_LDFLAGS} github.com/zhangpeihao/zim
	mkdir -p release
	cp $(GOPATH)/bin/zim release/

build_cross:
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -o release/linux/arm64/zim   ${GO_LDFLAGS} github.com/zhangpeihao/zim
	GOOS=linux   GOARCH=arm   CGO_ENABLED=0 go build -o release/linux/arm/zim     ${GO_LDFLAGS} github.com/zhangpeihao/zim
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o release/windows/amd64/zim ${GO_LDFLAGS} github.com/zhangpeihao/zim
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -o release/darwin/amd64/zim  ${GO_LDFLAGS} github.com/zhangpeihao/zim

build_tar:
	tar -cvzf release/linux/arm64/zim.tar.gz   -C release/linux/arm64   zim
	tar -cvzf release/linux/arm/zim.tar.gz     -C release/linux/arm     zim
	tar -cvzf release/windows/amd64/zim.tar.gz -C release/windows/amd64 zim
	tar -cvzf release/darwin/amd64/zim.tar.gz  -C release/darwin/amd64  zim
