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

	err = validateSchema(manifestSchema, data)
	if err != nil {
		return nil, errors.Wrap(err, "manifest schema validation failure")
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

type Deployment struct {
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
		if input.Tenant != "" && tenant.Name != input.Tenant {
			continue
		}

		tpl.Tenant = tenant

		err := tenant.ApplyTemplate(tpl)
		if err != nil {
			return nil, errors.Wrapf(err, "apply template to tenant %s", tenant.Name)
		}

		for k, v := range tenant.Tags {
			tpl.Tags[k] = v
		}

		for _, stack := range m.Stacks {
			if input.Stack != "" && stack.Name != input.Stack {
				continue
			}

			for k, v := range stack.Tags {
				tpl.Tags[k] = v
			}

			for _, target := range stack.Targets {
				if target.Tenant != tenant.Name {
					continue
				}

				if target == nil {
					return nil, errors.Errorf(
						"no deployment of stack %s for tenant %s",
						stack.Name,
						tenant.Name)
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

				decision := Deployment{
					AccountId:    deployment.AccountId,
					Region:       deployment.Region,
					TemplateBody: string(templateBody),
					Parameters:   parameters,
					StackName:    deployment.StackName,
					Protected:    false,
					Tenant:       NameLabel{tenant.Name, tenant.Label},
					Stack:        NameLabel{stack.Name, stack.Label},
				}

				out = append(out, &decision)
			}
		}
	}

	return out, nil
}
