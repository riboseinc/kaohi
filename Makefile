
KAOHI_DAEMON_BIN = kaohi
KAOHI_CONSOLE_BIN = kaohi_console
KAOHI_DAEMON_GO_FILES = kaohi.go daemon.go logger.go util.go config.go common.go
CURDIR = $(shell pwd)
GOPATH = $(CURDIR)/.gopath
GOARCH = amd64

all: dependencies darwin

dependencies:
	GOPATH=${GOPATH} go get github.com/tidwall/match
	GOPATH=${GOPATH} go get github.com/bitly/go-simplejson

darwin:
	GOPATH=${GOPATH} GOOS=darwin GOARCH=${GOARCH} go build  -o bin/${KAOHI_DAEMON_BIN}-darwin-${GOARCH} ${KAOHI_DAEMON_GO_FILES}

test: dependencies
	GOPATH=${GOPATH} GOOS=darwin GOARCH=${GOARCH} go build  -o tests/test_config test_config.go config.go common.go
	GOPATH=${GOPATH} GOOS=darwin GOARCH=${GOARCH} go build  -o tests/test_logger test_logger.go logger.go common.go

clean:
	rm -rf bin/* tests/*
	rm -rf ${GOPATH}
