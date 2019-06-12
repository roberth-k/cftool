cftool
===

cftool ("CloudFormation Tool") is a boring command-line program for working with CloudFormation. It is mainly a tool for generating and visualising change sets ahead of stack deployments, cutting down on iteration time. It works with any existing CloudFormation templates, and does not apply any preprocessing of its own. 

# Installing

- [Download latest Linux binary](https://github.com/tetratom/cftool/releases/latest/download/linux-amd64.zip)
- [Download latest Windows binary](https://github.com/tetratom/cftool/releases/latest/download/windows-amd64.zip)
- [Download latest macOS binary](https://github.com/tetratom/cftool/releases/latest/download/darwin-amd64.zip)

Alternatively, install it with `go get github.com/tetratom/cftool`.

# Features

## Update Stack

This is essentially equivalent to `aws cloudformation create-change-set` followed by `aws cloudformation execute-change-set`, plus some `describe-stack` operations to monitor the status of a deployment. The program will exit when the stack update is complete. If an error is encountered and the stack rolls back, cftool prints these errors and waits for rollback completion. Stack outputs are written out at the end of a successful update.

Example:

```sh
$ cftool -p live update \
    -t templates/base/network.yml \
    -p stacks/live/eu-west-1/live-base-network.json \
    -n live-base-network
```

The default behaviour is to display a summary of the change set, and to prompt the user for confirmation before executing it. This can be bypassed with `-y/--yes`, although it will still ask if the stack doesn't exist at all.

The optional `-d` parameter will display a diff comparing the current and updated templates if the operation is a stack update.

See below for detailed information about usage. 

## Deploy Stack from Manifest

A manifest file (typically `.cftool.yml` at the root of an IaC repository) is used to declare a matrix of tenants and stacks. This manifest is a complete description of the templates, parameter files, stack names, and other information such as applicable AWS regions. The manifest supports templating to prevent repetition.

Example:

```sh
$ cftool -p live deploy -t live -s network
```

See below for detailed information about usage.

# Usage

```sh
$ cftool [general-options] [subcommand] [subcommand-options]
```

## General

```
-p/--profile PROFILE: override AWS profile.
-r/--region REGION: override default AWS region.
-e/--endpoint ENDPOINT: override CloudFormation endpoint.
-v/--verbose: enable verbose output.
-c/--color on|off: enable or disable colorized output (default: on). 
```

## `cftool update`

```
cftool [general-options] update -t FILE [-p FILE ...] [-P KEY=VALUE ...] [-n NAME] [-d] [-y]

-t/--template FILE: path to CloudFormation template.
-p/--parameter-file FILE: path to CloudFormation parameter value.
-P/--parameter KEY=VALUE: override parameters directly.
-n/--stack-name NAME: override stack name.
-d/--diff: show a diff comparing the stack's template in CloudFormation to the template on disk. 
-y/--yes: do not prompt for confirmation when updating the stack.
```

If `-n NAME` is not provided, it is derived based on the following rules:

1. If there is exactly one `-p FILE`, take the name of the file without its extension.
2. Otherwise take the name of the `-t FILE` without its extension.

The `update` feature is optimised for a one-to-one correspondence between parameter files and stacks.  

## `cftool deploy`

```
cftool [general-options] deploy -t TENANT -s STACK [-f FILE] [-d] [-y]

-t/--tenant TENANT: tenant from the manifest.
-s/--stack STACK: stack from the manifest.
-f/--manifest FILE: path to manifest (default: .cfn-tool.yml in a parent directory).
-d/--diff: show a diff comparing the stack's template in CloudFormation to the template on disk. 
-y/--yes: do not prompt for confirmation when updating the stack.
```

# Development  

## Test

```sh
go test ./...
```

## Build

```sh
go run mage.go build:all   # Build for all targets into .build/$GOOS-$GOARCH.
go run mage.go install     # Build for host and install to $GOPATH/bin.
```

# Manifest files

A manifest file is a structure for recursively building up a deployment from a selection of tenants and stacks.

A deployment consists of the following fields:

```yaml
AccountId: "111111111111" # Supports template.
Parameters:
  - File: "stack-parameters.json" # Supports template.
  - File: "stack-parameters.yml"
  - Key: ExampleKey
    Value: "ExampleValue" # Supports template.
Protected: false
Region: eu-west-1 # Supports template.
StackName: "my-stack" # Supports template.
Template: "my-template.yml" # Supports template.
```

The broad structure of a manifest is:

```yaml
Version: "1.0"

Global:
    Constants:
      ExampleConstant: "ExampleConstantValue"
    Default: {} # Deployment.
      
Tenants:
    - Name: ExampleTenant
      Label: Example Tenant Label # Supports template.
      Default: {} # Deployment
      Tags:
        ExampleTag: "ExampleTagValue" # Supports template.
        
Stacks:
    - Name: ExampleStackName
      Label: Example Stack
      Default: {} # Deployment.
      Targets:
        - Tenant: ExampleTenant
          Override: {} # Deployment
```

The final deployment is built up by merging, in this order, the `Default` fields of Global, the selected tenant, and the selected stack, and finally with the `Override` of the deployment.

### Template syntax

Some fields (see comments above) support template replacement using Go template syntax. Replacements are run from the following structure:

```go
type Template struct {
	Constants  map[string]string
	Tags       map[string]string
	Tenant     *Tenant
	Deployment *Deployment
}
```

`"{{.Tags.Env}}-mystack"` is an example of Go templates in use; more can be found in the [manifest/testdata](manifest/testdata) directory. Note that a templated value will probably have to be surrounded by quotation marks to de-conflict YAML.

Values are available for templating in deployment merge order. The structure being substituted is available when substituting the structure itself (e.g. other deployment fields can reference the `StackName`). Substitution runs on nested fields last.
