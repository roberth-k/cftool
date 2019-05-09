execname := cfn-tool

ifeq ($(GOOS),windows)
	execname := $(execname).exe
endif

.PHONY: all
all: build build-target-all

.PHONY: build
build:
	@mkdir -p .build
	@go build -o .build/cfn-tool $(CURDIR)/cli

.PHONY: build-target-all
build-target-all: \
	build \
	build-target-windows \
	build-target-linux \
	build-target-darwin

.PHONY: build-target
build-target:
	@mkdir -p .build/$(GOOS)/$(GOARCH)
	@go build -o .build/$(GOOS)/$(GOARCH)/$(execname) $(CURDIR)/cli

.PHONY: build-target-windows
build-target-windows:
	@GOOS=windows GOARCH=amd64 $(MAKE) build-target

.PHONY: build-target-linux
build-target-linux:
	@GOOS=linux GOARCH=amd64 $(MAKE) build-target

.PHONY: build-target-darwin
build-target-darwin:
	@GOOS=darwin GOARCH=amd64 $(MAKE) build-target
