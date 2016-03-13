package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
)

func init() {
	tmpl.FuncMap["component"] = func(ctx context.Context, stores interface{}, component string, props ...interface{}) (template.HTML, error) {
		if len(props)%2 == 1 {
			return "", errors.New("component requires an even number of prop key-value items (key1 val1 key2 val2 ...)")
		}
		propMap := make(map[string]interface{}, len(props)/2)
		for i := 0; i < len(props); i++ {
			if i%2 == 1 {
				name, ok := props[i-1].(string)
				if !ok {
					return "", errors.New("component requires props to be specified in key-value pairs where the keys are strings (key1 val1 key2 val2 ...)")
				}
				propMap[name] = props[i]
			}
		}

		propsJSON, err := json.Marshal(propMap)
		if err != nil {
			return "", err
		}

		return template.HTML(fmt.Sprintf(`<div data-react="%s" data-props="%s"></div>`,
			html.EscapeString(component),
			html.EscapeString(string(propsJSON)),
		)), nil
	}
}
