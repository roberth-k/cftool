.PHONY: build
build:
	@mkdir -p .build
	@go build -o .build/cfn-tool cli/*.go
