package manifest

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
	"strings"
)

//go:generate go run -tags codegen ./schema_codegen.go

func validateSchema(schema []byte, data []byte) error {
	schemaJson, err := yaml.YAMLToJSON(schema)
	if err != nil {
		return errors.Wrap(err, "schema yaml to json conversion error")
	}

	dataJson, err := yaml.YAMLToJSON(data)
	if err != nil {
		return errors.Wrap(err, "data yaml to json conversion error")
	}

	schemaLoader := gojsonschema.NewBytesLoader(schemaJson)
	documentLoader := gojsonschema.NewBytesLoader(dataJson)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		validationErrors := make([]string, len(result.Errors()))

		for _, resultError := range result.Errors() {
			validationErrors = append(validationErrors, resultError.String())
		}

		return errors.New(strings.Join(validationErrors, "; "))
	}

	return err
}
