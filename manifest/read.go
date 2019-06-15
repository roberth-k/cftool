package manifest

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
)

func readWithValidation(r io.Reader, schema []byte, out interface{}) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	err = validateSchema(schema, data)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, out)
	if err != nil {
		return err
	}

	return nil
}

func Read(r io.Reader) (*Manifest, error) {
	var m Manifest
	err := readWithValidation(r, manifestSchema, &m)
	if err != nil {
		return nil, err
	}

	if m.Version != SupportedVersion {
		return nil, errors.Errorf("expected version %s", SupportedVersion)
	}

	return &m, nil
}

func ReadFromFile(path string) (*Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return Read(f)
}

func ReadParameters(r io.Reader) (map[string]string, error) {
	var params []cloudformation.Parameter
	err := readWithValidation(r, parametersSchema, &params)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, param := range params {
		result[*param.ParameterKey] = *param.ParameterValue
	}

	return result, nil
}

func ReadParametersFromFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return ReadParameters(f)
}
