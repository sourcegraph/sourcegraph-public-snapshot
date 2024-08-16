package config

import (
	"text/template"

	"github.com/Masterminds/sprig"
	"gopkg.in/yaml.v3"
)

var TemplateFuncMap template.FuncMap

func init() {
	TemplateFuncMap = sprig.FuncMap()
	TemplateFuncMap["toYaml"] = func(obj any) (string, error) {
		serialized, err := yaml.Marshal(obj)
		return string(serialized), err
	}
}
