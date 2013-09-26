HTMLSRC := $(shell find html -type f)
HTMLERB := html/index.html.erb
HTMLBIN := index.html
GOBIN   := webmpcd

export GOPATH := $(shell pwd)

.PHONY: all

all: $(GOBIN) $(HTMLBIN)

$(HTMLBIN): $(HTMLSRC)
	erb -r sass -r coffee_script -r uglifier $(HTMLERB) > $@

$(GOBIN): deps
	go build -v $@

deps:
	go get -d -v webmpc/...

clean:
	rm -f $(GOBIN) $(HTMLBIN)

fmt:
	gofmt -l -w -tabs=false -tabwidth=2 src/webmpc{,d}
