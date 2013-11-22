export GOPATH := $(shell pwd)

.PHONY: all

all: webmpcd index.html

index.html: html/index.html.erb html/webmpc.js html/webmpc.css
	erb html/index.html.erb > $@

html/webmpc.js: html/webmpc.coffee.erb
	erb -r base64 $^ | coffee -s -c | uglifyjs -m -c > $@

html/webmpc.css: html/webmpc.scss
	scss $^ | cleancss -o $@

webmpcd: deps
	go build -v $@

deps:
	go get -d -v webmpc/...

clean:
	rm -f webmpcd index.html html/{webmpc.js,webmpc.css}

fmt:
	gofmt -l -w -tabs=false -tabwidth=2 src/webmpc{,d}
