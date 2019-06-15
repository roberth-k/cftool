package manifest

import (
	"io/ioutil"
	"strings"
	"text/template"
)

const SupportedVersion = "1.1"

type Global struct {
	Constants map[string]string
	Tags      map[string]string
	Default   *Defaults
}

type Tenant struct {
	Label     string
	Default   *Defaults
	Constants map[string]string
	Tags      map[string]string
}

type Stack struct {
	Label   string
	Default *Defaults
	Targets []*Target
	Tags    map[string]string
}

type Target struct {
	Tenant   string
	Override *Defaults
}

type Defaults struct {
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
	Protected *bool
}

func (d Defaults) MergeFrom(other *Defaults) Defaults {
	if other == nil {
		return d
	}

	add := func(to *string, from *string) {
		if *from != "" {
			*to = *from
		}
	}

	add(&d.AccountId, &other.AccountId)
	add(&d.Region, &other.Region)
	add(&d.Template, &other.Template)
	add(&d.StackName, &other.StackName)

	for _, p := range other.Parameters {
		d.Parameters = append(d.Parameters, p)
	}

	if other.Protected != nil {
		d.Protected = other.Protected
	}

	return d
}

type Parameter struct {
	// File is the path of a parameter file relative to Config.
	File  string
	Key   string
	Value string
}

type Manifest struct {
	Version string
	Global  Global
	Tenants []*Tenant
	Stacks  []*Stack
}

func applyTemplate(text string, data interface{}) (string, error) {
	parsed, err := template.
		New("Template").
		Option("missingkey=error").
		Parse(text)

	if err != nil {
		return "", err
	}

	w := strings.Builder{}
	err = parsed.Execute(&w, data)
	if err != nil {
		return "", err
	}

	return w.String(), nil
}

func extendMap(a, b map[string]string) {
	for k, v := range b {
		a[k] = v
	}
}

func (m *Manifest) Deployment(
	tenant *Tenant,
	stack *Stack,
	target *Target,
) (result *Deployment, err error) {
	def := Defaults{}.
		MergeFrom(m.Global.Default).
		MergeFrom(tenant.Default).
		MergeFrom(stack.Default).
		MergeFrom(target.Override)

	// set up the initial values
	d := Deployment{
		TenantLabel: tenant.Label,
		StackLabel:  stack.Label,
	}

	if def.Protected != nil {
		d.Protected = *def.Protected
	}

	// externally we say it's the Deployment structure providing the data,
	// but we build up this map instead to control the variables that
	// are available. this is to enforce the order of templating operations.
	tpl := make(map[string]interface{})
	constants := make(map[string]string)
	tags := make(map[string]string)

	tpl["TenantLabel"] = d.TenantLabel
	tpl["StackLabel"] = d.StackLabel

	extendMap(constants, m.Global.Constants)
	extendMap(constants, tenant.Constants)
	tpl["Constants"] = constants
	d.Constants = constants

	extendMap(tags, m.Global.Tags)
	extendMap(tags, tenant.Tags)
	for k, v := range tags {
		tags[k], err = applyTemplate(v, tpl)
		if err != nil {
			return
		}
	}
	tpl["Tags"] = tags
	d.Tags = tags

	d.AccountId, err = applyTemplate(def.AccountId, tpl)
	if err != nil {
		return
	}
	tpl["AccountId"] = d.AccountId

	d.Region, err = applyTemplate(def.Region, tpl)
	if err != nil {
		return
	}
	tpl["Region"] = d.Region

	d.StackName = def.StackName
	d.StackName, err = applyTemplate(def.StackName, tpl)
	if err != nil {
		return
	}
	tpl["StackName"] = d.StackName

	templatePath, err := applyTemplate(def.Template, tpl)
	if err != nil {
		return
	}
	d.TemplateBody, err = ioutil.ReadFile(templatePath)
	if err != nil {
		return nil, err
	}

	d.Parameters = make(map[string]string)
	for _, p := range def.Parameters {
		switch {
		case p.File != "":
			path, err := applyTemplate(p.File, tpl)
			if err != nil {
				return nil, err
			}

			kvp, err := ReadParametersFromFile(path)
			if err != nil {
				return nil, err
			}
			extendMap(d.Parameters, kvp)
		default:
			d.Parameters[p.Key], err = applyTemplate(p.Value, tpl)
			if err != nil {
				return
			}
		}
	}

	return &d, nil
}

func (m *Manifest) FindDeployment(tenantLabel string, stackLabel string) (*Deployment, bool, error) {
	var tenant *Tenant
	for _, t := range m.Tenants {
		if t.Label == tenantLabel {
			tenant = t
			break
		}
	}
	if tenant == nil {
		return nil, false, nil
	}

	var stack *Stack
	var target *Target
	for _, s := range m.Stacks {
		if s.Label == stackLabel {
			stack = s
			for _, t := range stack.Targets {
				if t.Tenant == tenantLabel {
					target = t
					break
				}
			}
			break
		}
	}
	if stack == nil || target == nil {
		return nil, false, nil
	}

	d, err := m.Deployment(tenant, stack, target)
	return d, true, err
}
