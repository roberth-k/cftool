package manifest

type Deployment struct {
	TenantLabel  string
	StackLabel   string
	Protected    bool
	Constants    map[string]string
	Tags         map[string]string
	AccountId    string
	Region       string
	StackName    string
	TemplateBody []byte
	Parameters   map[string]string
}
