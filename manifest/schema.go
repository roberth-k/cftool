package manifest

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
	"strings"
)

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

var manifestSchema = []byte(`
$schema: "http://json-schema.org/draft-07/schema#"
type: object
required:
  - Version
properties:
  Version:
    type: string
    enum: ["1.0"]
  Global:
    type: object
    properties:
      Constants:
        $ref: "#/definitions/TagSet"
  Tenants:
    type: array
    items:
      type: object
      required:
        - Name
      properties:
        Name:
          type: string
        Label:
          type: string
        Default:
          $ref: "#/definitions/Stack"
        Tags:
          $ref: "#/definitions/TagSet"
  Stacks:
    type: array
    items:
      type: object
      required:
        - Name
      properties:
        Name:
          type: string
        Label:
          type: string
        Default:
          $ref: "#/definitions/Stack"
        Targets:
          type: array
          items:
            $ref: "#/definitions/Target"

definitions:
  TagSet:
    type: object
    additionalProperties:
      type: string

  Parameter:
    $oneOf:
      - type: object
        required:
          - File
        properties:
          File:
            type: string
      - type: object
        required:
          - Key
          - Value
        properties:
          Key:
            type: string
          Value:
            type: string

  Stack:
    type: object
    properties:
      AccountId:
        type: string
      Parameters:
        type: array
        items:
          $ref: "#/definitions/Parameter"
      Protected:
        type: boolean
      Region:
        type: string
      StackName:
        type: string
      Template:
        type: string

  Target:
    type: object
    required:
      - Tenant
    properties:
      Tenant:
        type: string
      Override:
        $ref: "#/definitions/Stack"
`)

var parametersSchema = []byte(`
$schema: "http://json-schema.org/draft-07/schema#"
type: array
items:
  type: object
  required:
    - ParameterKey
    - ParameterValue
  properties:
    ParameterKey:
      type: string
    ParameterValue:
      type: string
`)
