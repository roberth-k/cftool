package manifest

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
)

func Parse(reader io.Reader) (*Manifest, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "read manifest")
	}

	var m Manifest
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshal manifest")
	}

	if m.Version != Version1_0 {
		return nil, errors.Errorf("expected version %s", Version1_0)
	}

	return &m, nil
}

type ProcessInput struct {
	Stack  string
	Tenant string
}

type Decision struct {
	AccountId    string
	Region       string
	TemplateBody string
	Parameters   map[string]string
	StackName    string
	Protected    bool
	Tenant       NameLabel
	Stack        NameLabel
}

type NameLabel struct {
	Name  string
	Label string
}

func (m *Manifest) Process(input ProcessInput) ([]*Decision, error) {
	var stack *Stack
	var tenant *Tenant
	var stackDeployment *Target
	tpl := NewTemplate()

	for k, v := range m.Global.Constants {
		tpl.Constants[k] = v
	}

	for k, v := range m.Global.Tags {
		tpl.Tags[k] = v
	}

	for _, t := range m.Tenants {
		if t.Name == input.Tenant {
			tenant = t
			break
		}
	}

	if tenant == nil {
		return nil, errors.Errorf("tenant '%s' not found", input.Tenant)
	}

	tpl.Tenant = tenant

	for k, v := range tenant.Tags {
		tpl.Tags[k] = v
	}

	err := tenant.ApplyTemplate(tpl)
	if err != nil {
		return nil, errors.Wrapf(err, "apply template to tenant %s", tenant.Name)
	}

	for _, s := range m.Stacks {
		if s.Name == input.Stack {
			stack = s
			break
		}
	}

	if stack == nil {
		return nil, errors.Errorf("stack '%s' not found", input.Stack)
	}

	for k, v := range stack.Tags {
		tpl.Tags[k] = v
	}

	for _, sd := range stack.Targets {
		if sd.Tenant == tenant.Name {
			stackDeployment = sd
			break
		}
	}

	if stackDeployment == nil {
		return nil, errors.Errorf(
			"no deployment of stack %s for tenant %s",
			stack.Name,
			tenant.Name)
	}

	var deployment Deployment
	deployment.MergeFrom(m.Global.Default)
	deployment.MergeFrom(tenant.Default)
	deployment.MergeFrom(stack.Default)
	deployment.MergeFrom(stackDeployment.Override)

	tpl.Deployment = &deployment

	err = deployment.ApplyTemplate(tpl)
	if err != nil {
		return nil, errors.Wrapf(err, "apply template to deployment")
	}

	templateBody, err := ioutil.ReadFile(deployment.Template)
	if err != nil {
		return nil, errors.Wrapf(err, "read %s", deployment.Template)
	}

	parameters := map[string]string{}
	for _, param := range deployment.Parameters {
		if param.File != "" {
			kvp, err := ParseParameterFile(param.File)
			if err != nil {
				return nil, errors.Wrapf(err, "parse parameter file")
			}

			for k, v := range kvp {
				parameters[k] = v
			}
		} else {
			parameters[param.Key] = param.Value
		}
	}

	result := []*Decision{
		{
			AccountId:    deployment.AccountId,
			Region:       deployment.Region,
			TemplateBody: string(templateBody),
			Parameters:   parameters,
			StackName:    deployment.StackName,
			Protected:    false,
			Tenant:       NameLabel{tenant.Name, tenant.Label},
			Stack:        NameLabel{stack.Name, stack.Label},
		},
	}

	return result, nil
}
