
KAOHI_DAEMON_BIN = kaohi
KAOHI_CONSOLE_BIN = kaohi_console
KAOHI_DAEMON_GO_FILES = kaohi.go logger.go util.go config.go common.go cmd.go watcher.go config_mel.go
CURDIR = $(shell pwd)
GOPATH = $(CURDIR)/.gopath
GOARCH = amd64

all: dependencies darwin

dependencies:
	GOPATH=${GOPATH} go get github.com/riboseinc/go-nereon
	GOPATH=${GOPATH} go get github.com/riboseinc/go-nereon/genconfig

darwin:
	${GOPATH}/bin/genconfig -generate config.mel
	GOPATH=${GOPATH} GOOS=darwin GOARCH=${GOARCH} go build -o bin/${KAOHI_DAEMON_BIN}-darwin-${GOARCH} ${KAOHI_DAEMON_GO_FILES}

test: dependencies
#	GOPATH=${GOPATH} GOOS=darwin GOARCH=${GOARCH} go build  -o tests/test_config test_config.go config.go common.go util.go
#	GOPATH=${GOPATH} GOOS=darwin GOARCH=${GOARCH} go build  -o tests/test_logger test_logger.go logger.go common.go
#	GOPATH=${GOPATH} GOOS=darwin GOARCH=${GOARCH} go build  -o tests/test_cmd test_cmd.go cmd.go common.go logger.go
#	GOPATH=${GOPATH} GOOS=darwin GOARCH=${GOARCH} go build  -o tests/test_watcher test_watcher.go watcher.go common.go logger.go

clean:
	rm -rf bin/* tests/*
	rm -rf ${GOPATH}
	rm -rf config_mel.go
