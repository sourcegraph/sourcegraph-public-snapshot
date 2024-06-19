package types

type DefaultModels struct {
	Chat           ModelRef `json:"chat"`
	FastChat       ModelRef `json:"fastChat"`
	CodeCompletion ModelRef `json:"codeCompletion"`
}

type ModelMap map[ModelRef][]Model

const CurrentModelSchemaVersion = "1.0"

type ModelConfiguration struct {
	SchemaVersion string `json:"schemaVersion"`
	Revision      string `json:"revision"`

	Providers []Provider `json:"providers"`
	Models    []Model    `json:"models"`

	DefaultModels DefaultModels `json:"defaultModels"`
}
