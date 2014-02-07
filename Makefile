export GOPATH := $(shell pwd)

.PHONY: all

all: webmpcd html/css/webmpc.css

html/css/webmpc.css: html/css/webmpc.scss
	scss -I html/css -t compressed $^ $@

webmpcd: deps
	go build -v $@

deps:
	go get -d -v webmpc/...

clean:
	rm -f webmpcd html/css/webmpc.css

fmt:
	gofmt -l -w -tabs=false -tabwidth=2 src/webmpc{,d}
