GOSRC   := $(shell find src -name '*.go')
GOBIN   := webmpcd
HTMLSRC := $(shell find html -type f)
HTMLERB := html/index.html.erb
HTMLBIN := index.html

.PHONY: all

all: $(GOBIN) $(HTMLBIN)

$(HTMLBIN): $(HTMLSRC)
	erb -r sass -r coffee_script -r uglifier $(HTMLERB) > $@

$(GOBIN): $(GOSRC)
	gd -o $@

clean:
	rm -f $(GOBIN) $(HTMLBIN)
	gd clean

fmt:
	gd fmt -w2
