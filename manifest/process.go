package manifest

import (
	"github.com/pkg/errors"
	"io/ioutil"
)

type ProcessInput struct {
	Stack  string
	Tenant string
}

type Deployment struct {
	AccountId    string
	Region       string
	TemplateBody string
	Parameters   map[string]string
	StackName    string
	Protected    bool
	TenantLabel  string
	StackLabel   string
	Tags         map[string]string
	Constants    map[string]string
}

type NameLabel struct {
	Name  string
	Label string
}

func (m *Manifest) Process(input ProcessInput) ([]*Deployment, error) {
	out := []*Deployment{}

	tpl := NewTemplate()

	for k, v := range m.Global.Constants {
		tpl.Constants[k] = v
	}

	for k, v := range m.Global.Tags {
		tpl.Tags[k] = v
	}

	for _, tenant := range m.Tenants {
		if input.Tenant != "" && tenant.Label != input.Tenant {
			continue
		}

		tpl.Tenant = tenant

		err := tenant.ApplyTemplate(tpl)
		if err != nil {
			return nil, errors.Wrapf(err, "apply template to tenant %s", tenant.Label)
		}

		for k, v := range tenant.Tags {
			tpl.Tags[k] = v
		}

		for _, stack := range m.Stacks {
			if input.Stack != "" && stack.Label != input.Stack {
				continue
			}

			for k, v := range stack.Tags {
				tpl.Tags[k] = v
			}

			for _, target := range stack.Targets {
				if target.Tenant != tenant.Label {
					continue
				}

				if target == nil {
					return nil, errors.Errorf(
						"no deployment of stack %s for tenant %s",
						stack.Label,
						tenant.Label)
				}

				deployment := Defaults{}
				deployment.MergeFrom(m.Global.Default)
				deployment.MergeFrom(tenant.Default)
				deployment.MergeFrom(stack.Default)
				deployment.MergeFrom(target.Override)

				tpl.Stack = &deployment

				err = deployment.ApplyTemplate(tpl)
				if err != nil {
					return nil, errors.Wrapf(err, "apply template to deployment")
				}

				templateBody, err := ioutil.ReadFile(deployment.Template)
				if err != nil {
					return nil, err
				}

				parameters := map[string]string{}
				for _, param := range deployment.Parameters {
					if param.File != "" {
						kvp, err := ReadParametersFromFile(param.File)
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

				decision := Deployment{
					AccountId:    deployment.AccountId,
					Region:       deployment.Region,
					TemplateBody: string(templateBody),
					Parameters:   parameters,
					StackName:    deployment.StackName,
					Protected:    false,
					TenantLabel:  tenant.Label,
					StackLabel:   stack.Label,
					Constants:    tpl.Constants,
					Tags:         tpl.Tags,
				}

				out = append(out, &decision)
			}
		}
	}

	return out, nil
}
