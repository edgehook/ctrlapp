BUILD_ROOT=$(PWD)
BUILD_TARGET=ctrlapp

.PHONY:	${BUILD_TARGET} ${BUILD_TARGET}_arm64 ${BUILD_TARGET}_arm ${BUILD_TARGET}_win.exe
all: ${BUILD_TARGET}

${BUILD_TARGET}:
	@gofmt -l -w ${BUILD_ROOT}/
	@export GO111MODULE=on && \
	export GOPROXY=https://goproxy.io && \
	go build -ldflags "-w" -o $@ ctrlapp.go
	@chmod 777 $@

${BUILD_TARGET}_arm64:
	@gofmt -l -w ${BUILD_ROOT}/
	@export GO111MODULE=on && \
	export GOPROXY=https://goproxy.io && \
	GOARCH=arm64 GOOS="linux" CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc go build -ldflags "-w" -o $@ ctrlapp.go
	@chmod 777 $@ 

${BUILD_TARGET}_arm:
	@gofmt -l -w ${BUILD_ROOT}/
	@export GO111MODULE=on && \
	export GOPROXY=https://goproxy.io && \
	GOARCH=arm GOOS="linux" GOARM=7 CGO_ENABLED=1 CC=arm-linux-gnueabi-gcc go build -ldflags "-w -extldflags -static" -o $@ ctrlapp.go
	@chmod 777 $@ 

${BUILD_TARGET}_win.exe:
	@gofmt -l -w ${BUILD_ROOT}/
	@export GO111MODULE=on && \
	export GOPROXY=https://goproxy.io && \
	GOARCH=amd64 GOOS="windows" CGO_ENABLED=1 go build -o $@ ctrlapp.go
	@chmod 777 $@ 

install:
	@mkdir -p out
	@chmod 777 ${BUILD_TARGET}
	@cp -a conf ${BUILD_TARGET} out/
	sync;sync
	@echo "[Done]"

.PHONY: clean  install
clean:
	@rm -rf ${BUILD_TARGET} ${BUILD_TARGET}_arm64  ${BUILD_TARGET}_arm ${BUILD_TARGET}_win.exe *.db *.tar.gz
	@echo "[clean Done]"
