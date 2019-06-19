progname := cftool
ifeq ($(GOOS),windows)
	progname := $(progname).exe
endif

.PHONY: all
all: build-all

.PHONY: build-all
build-all:
	@GOARCH=amd64 GOOS=windows $(MAKE) build-target
	@GOARCH=amd64 GOOS=linux   $(MAKE) build-target
	@GOARCH=amd64 GOOS=darwin  $(MAKE) build-target

.PHONY: build-target
build-target:
	@mkdir -p .build/$(GOOS)-$(GOARCH)
	go build -o .build/$(GOOS)-$(GOARCH)/$(progname) -ldflags '-s -w' .
