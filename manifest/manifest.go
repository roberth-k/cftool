package manifest

const Version1_0 = "1.0"

type Manifest struct {
	Version string `json:"Version"`
	Global  Global
	Tenants []Tenant
	Stacks  []Stack
}

type Global struct {
	Constants map[string]string
}

type Tenant struct {
	Name    string
	Default Deployment
	Tags    map[string]string
}

type Stack struct {
	// Label is the longer, human-readable summary of a stack.
	Name        string
	Label       string
	Default     Deployment
	Deployments []StackDeployment
}

type StackDeployment struct {
	Tenant   string
	Override Deployment
}

type Deployment struct {
	// AccountId is an AWS account ID to check the profile against.
	AccountId string

	// Region is an AWS region, if different from the profile's default.
	Region string

	// Template is the path of a template file relative to Config.
	Template string

	// Parameter contains paths to parameter files and direct overrides.
	Parameters []*Parameter

	// StackName can include substitutions (as Go templates).
	StackName string

	// Protected deployments ignore the --yes flag.
	Protected bool
}

type Parameter struct {
	// File is the path of a parameter file relative to Config.
	File  string
	Key   string
	Value string
}
