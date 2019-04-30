cfn-tool (CloudFormation Tool)
===

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
