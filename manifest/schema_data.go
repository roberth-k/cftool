// Code generated by schema_codegen.go. DO NOT EDIT.

package manifest
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
var manifestSchema = []byte(`
$schema: "http://json-schema.org/draft-07/schema#"
type: object
additionalProperties: false
required:
  - Version
properties:
  Version:
    type: string
    enum: ["1.1"]
  Global:
    type: object
    properties:
      Constants:
        $ref: "#/definitions/TagSet"
  Tenants:
    type: array
    items:
      type: object
      additionalProperties: false
      required:
        - Label
      properties:
        Label:
          type: string
        Constants:
          $ref: "#/definitions/TagSet"
        Default:
          $ref: "#/definitions/Stack"
        Tags:
          $ref: "#/definitions/TagSet"
  Stacks:
    type: array
    items:
      type: object
      additionalProperties: false
      required:
        - Label
      properties:
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
        additionalProperties: false
        required:
          - File
        properties:
          File:
            type: string
      - type: object
        additionalProperties: false
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
    additionalProperties: false
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
    additonalProperties: false
    required:
      - Tenant
    properties:
      Tenant:
        type: string
      Override:
        $ref: "#/definitions/Stack"
`)

