.PHONY: all build build_all build_cross docker test fmt lint vet build stop_stress_test

BUILD := ./build

TAG = $(shell git describe --match 'v[0-9]*' --always)
VERSION_MAJOR = $(shell awk '/VersionMajor = / { print $$3; exit }' ./pkg/version/version.go)
VERSION_MINOR = $(shell awk '/VersionMinor = / { print $$3; exit }' ./pkg/version/version.go)
VERSION_PATCH = $(shell awk '/VersionPatch = / { print $$3; exit }' ./pkg/version/version.go)
VERSION_BASE = $(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_PATCH)
VERSION = $(VERSION_BASE)-$(TAG)

PACKAGES = $(shell go list ./... | grep -v -e vendor -e tmp)
ALL_SRC := $(shell find . -name "*.go" | grep -v -e tmp -e vendor \
	-e ".*/\..*" \
	-e ".*/_.*" \
	-e ".*/mocks.*")
TEST_DIRS := $(sort $(dir $(filter %_test.go,$(ALL_SRC))))

DOCKER_RELEASE_FILE = release/docker/zim-$(VERSION_BASE).docker
ZIM_RELEASE_FILES = release/linux/amd64/zim.tar.gz release/linux/amd64/zim release/linux/386/zim.tar.gz release/linux/386/zim release/windows/amd64/zim.tar.gz release/windows/amd64/zim.exe release/windows/386/zim.tar.gz release/windows/386/zim.exe release/darwin/amd64/zim.tar.gz release/darwin/amd64/zim
RELEASE_FILES = $(DOCKER_RELEASE_FILE) $(ZIM_RELEASE_FILES)

GO_LDFLAGS=-ldflags "-X github.com/zhangpeihao/zim/pkg/version.VersionDev=$(TAG)"

all: build

# build the release files
build_all: build build_cross build_tar

build:
	go build ${GO_LDFLAGS} github.com/zhangpeihao/zim

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

rebuild: clean build

clean:
	rm -f ./zim

# build docker image
docker: docker_build docker_save

docker_build: build_cross_linux
	docker build -t zhangpeihao/zim:latest .
	docker tag zhangpeihao/zim:latest zhangpeihao/zim:$(VERSION_BASE)

docker_save: release/docker
	docker save -o $(DOCKER_RELEASE_FILE) zhangpeihao/zim:$(VERSION_BASE)

release/docker:
	mkdir -p release/docker

docker_push: docker
	docker push zhangpeihao/zim:latest
	docker push zhangpeihao/zim:$(VERSION_BASE)

# build test

test/service-stub/service-stub:
	cd test/service-stub && go build

buid_test_stub: test/service-stub/service-stub

# run
run_stub: build
	@nohup ./zim stub > ./test/stub.log 2>&1 &

run_gateway: build
	@nohup ./zim gateway --config=./test/gateway.yaml > ./test/gateway.log 2>&1 &

run_gateway_touch: run_gateway
	tail -f ./test/gateway.log

run_stress_client: build
	@nohup ./zim stress > ./test/stress.log 2>&1 &

stress_test: build
	@nohup ./zim stub > ./test/stub.log 2>&1 &
	@nohup ./zim gateway --config=./test/gateway.yaml > ./test/gateway.log 2>&1 &
	@nohup ./zim stress > ./test/stress.log 2>&1 &

stop_stress_test:
	@killall -2 zim && sleep 4

# fmt
fmt:
	gofmt -w -s cmd pkg test

# lint
lint:
	@for pkg in $(PACKAGES); do golint $$pkg; done

# vet
vet:
	@for pkg in $(PACKAGES); do go vet $$pkg; done

# test
test:
	@echo Testing packages:
	@mkdir -p $(BUILD)
	@echo "mode: atomic" > $(BUILD)/cover.out
	@for dir in $(TEST_DIRS); do \
		mkdir -p $(BUILD)/"$$dir"; \
		go test "$$dir" -race -v -timeout 5m -coverprofile=$(BUILD)/"$$dir"/coverage.out || exit 1; \
		cat $(BUILD)/"$$dir"/coverage.out | grep -v "mode: atomic" >> $(BUILD)/cover.out; \
	done

# cover
cover: test
	go tool cover -html=$(BUILD)/cover.out

cover_ci: build test
	goveralls -coverprofile=$(BUILD)/cover.out -service=travis-ci || echo -e "\x1b[31mCoveralls failed\x1b[m"

# code check
check: fmt lint vet test

# call graph
call_graph:
	go-callvis -sub pkg -limit github.com/zhangpeihao/zim github.com/zhangpeihao/zim | dot  -Tpng -o ./doc/call-graph.png

# release
release: build_tar docker_push call_graph
	git add $(RELEASE_FILES) ./doc/call-graph.png
	git commit -m "commit all release files by make"
