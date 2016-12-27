.PHONY: build docker

PACKAGES = $(shell go list ./... | grep -v /vendor/)
TAG = $(shell git describe --match 'v[0-9]*' --dirty --always)
VERSION_MAJOR = $(shell awk '/VersionMajor = / { print $$3; exit }' ./pkg/version/version.go)
VERSION_MINOR = $(shell awk '/VersionMinor = / { print $$3; exit }' ./pkg/version/version.go)
VERSION_PATCH = $(shell awk '/VersionPatch = / { print $$3; exit }' ./pkg/version/version.go)

VERSION = $(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_PATCH)-$(TAG)

GO_LDFLAGS=-ldflags "-X github.com/zhangpeihao/zim/pkg/version.VersionDev=$(TAG)"

all: build_static

test:
	go test -cover $(PACKAGES)

# build the release files
build: build_static build_cross build_tar

build_static:
	go build ${GO_LDFLAGS} github.com/zhangpeihao/zim
	mkdir -p release
	cp $(GOPATH)/bin/zim release/

build_cross: build_cross_linux build_cross_linux_386 build_cross_windows build_cross_windows_386 build_cross_darwin

build_cross_linux:
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -o release/linux/amd64/zim       ${GO_LDFLAGS} github.com/zhangpeihao/zim

build_cross_linux_386:
	GOOS=linux   GOARCH=386   CGO_ENABLED=0 go build -o release/linux/386/zim         ${GO_LDFLAGS} github.com/zhangpeihao/zim

build_cross_windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o release/windows/amd64/zim.exe ${GO_LDFLAGS} github.com/zhangpeihao/zim

build_cross_windows_386:
	GOOS=windows GOARCH=386   CGO_ENABLED=0 go build -o release/windows/386/zim.exe   ${GO_LDFLAGS} github.com/zhangpeihao/zim

build_cross_darwin:
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -o release/darwin/amd64/zim      ${GO_LDFLAGS} github.com/zhangpeihao/zim

build_tar: build_cross
	tar -cvzf release/linux/amd64/zim.tar.gz   -C release/linux/amd64   zim
	tar -cvzf release/linux/386/zim.tar.gz     -C release/linux/386     zim
	tar -cvzf release/windows/amd64/zim.tar.gz -C release/windows/amd64 zim.exe
	tar -cvzf release/windows/386/zim.tar.gz   -C release/windows/386   zim.exe
	tar -cvzf release/darwin/amd64/zim.tar.gz  -C release/darwin/amd64  zim

# build docker image
docker: docker_build docker_save

docker_build: build_cross_linux
	docker build -t zim:latest .
	docker tag zim:latest zim:$(VERSION)

docker_save: release/docker
	docker save -o release/docker/zim-$(VERSION).docker zim:$(VERSION)

release/docker:
	mkdir -p release/docker