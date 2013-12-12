export GOPATH := $(shell pwd)

.PHONY: all

all: webmpcd html/webmpc.css

html/webmpc.css: html/webmpc.scss
	scss -t compressed $^ $@

webmpcd: deps
	go build -v $@

deps:
	go get -d -v webmpc/...

clean:
	rm -f webmpcd html/webmpc.css

fmt:
	gofmt -l -w -tabs=false -tabwidth=2 src/webmpc{,d}
