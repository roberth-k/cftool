package manifest

import (
	"github.com/pkg/errors"
	"strings"
	"text/template"
)

const Version1_0 = "1.0"

type Manifest struct {
	Version string `json:"Version"`
	Global  Global
	Tenants []*Tenant
	Stacks  []*Stack
}

func (m *Manifest) AllDeployments() ([]*Decision, error) {
	return m.Process(ProcessInput{})
}

type Global struct {
	Constants map[string]string
	Tags      map[string]string
	Default   *Deployment
}

type Tenant struct {
	Name      string
	Constants map[string]string
	Label     string
	Default   *Deployment
	Tags      map[string]string
}

func (t *Tenant) ApplyTemplate(tpl *Template) (err error) {
	for k, v := range t.Constants {
		tpl.Constants[k] = v
	}

	err = tpl.ApplyTo(&t.Name)

	for k, v := range t.Tags {
		if err == nil {
			err = tpl.ApplyTo(&v)
			t.Tags[k] = v
		}
	}

	if err == nil && t.Default != nil {
		err = t.Default.ApplyTemplate(tpl)
	}

	return err
}

type Stack struct {
	// Label is the longer, human-readable summary of a stack.
	Name    string
	Label   string
	Default *Deployment
	Targets []*Target
	Tags    map[string]string
}

type Target struct {
	Tenant   string
	Override *Deployment
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
	Protected *bool
}

func (d *Deployment) ApplyTemplate(tpl *Template) (err error) {
	err = tpl.ApplyTo(&d.AccountId)

	if err == nil {
		err = tpl.ApplyTo(&d.Region)
	}

	if err == nil {
		err = tpl.ApplyTo(&d.Template)
	}

	if err == nil {
		err = tpl.ApplyTo(&d.StackName)
	}

	// Nested structures come last for templating.

	for _, p := range d.Parameters {
		if err == nil {
			err = p.ApplyTemplate(tpl)
		}
	}

	return err
}

func (d *Deployment) MergeFrom(other *Deployment) {
	if other == nil {
		return
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
}

type Parameter struct {
	// File is the path of a parameter file relative to Config.
	File  string
	Key   string
	Value string
}

func (p *Parameter) ApplyTemplate(tpl *Template) (err error) {
	err = tpl.ApplyTo(&p.File)

	if err == nil {
		err = tpl.ApplyTo(&p.Value)
	}

	return err
}

type Template struct {
	Constants map[string]string
	Tags      map[string]string
	Tenant    *Tenant
	Stack     *Deployment
}

func NewTemplate() *Template {
	return &Template{
		Constants: make(map[string]string),
		Tags:      make(map[string]string),
	}
}

func (tpl *Template) ApplyTo(text *string) error {
	parsed, err := template.New("Template").Parse(*text)
	if err != nil {
		return errors.Wrapf(err, "parse template text \"%s\"", *text)
	}

	w := strings.Builder{}
	err = parsed.Execute(&w, tpl)
	if err != nil {
		return errors.Wrapf(err, "execute template \"%s\"", *text)
	}

	*text = w.String()
	return nil
}
