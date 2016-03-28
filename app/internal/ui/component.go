package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
)

var funcs = template.FuncMap{
	"component": func(ctx context.Context, stores *StoreData, component string, props ...interface{}) (template.HTML, error) {
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

		var componentHTML string
		if ctx != nil && shouldPrerenderReact(ctx) {
			var err error
			ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
			defer cancel()
			componentHTML, err = renderReactComponent(ctx, component, propMap, stores)
			if err == errRendererCreationTimedOut {
				log15.Debug("Not rendering React component on the server because the JS renderer creation timed out. This is expected to occur right after the process starts (or in dev mode if the bundle JS was recently changed) before the JS renderer is ready.")
			} else if err != nil {
				log15.Error("Error rendering React component on the server (falling back to client-side rendering)", "err", err, "component", component, "props", string(propsJSON), "hasPreloadedStoreData", stores != nil)
			}
		}

		return template.HTML(fmt.Sprintf(`<div data-react="%s" data-props="%s">%s</div>`,
			html.EscapeString(component),
			html.EscapeString(string(propsJSON)),
			componentHTML,
		)), nil
	},
}

func init() {
	for name, f := range funcs {
		tmpl.FuncMap[name] = f
	}
}
