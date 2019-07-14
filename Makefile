progname := cftool
ifeq ($(GOOS),windows)
	progname := $(progname).exe
endif

version := $(shell git describe --tags --always --match 'v*')

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
	go build \
		-o .build/$(GOOS)-$(GOARCH)/$(progname) \
		-ldflags="-s -w -X $(shell go list ./internal/cli).gitVersion=$(version)" \
		.
