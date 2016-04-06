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

		var componentHTML []byte
		if ctx != nil && shouldPrerenderReact(ctx) {
			// Only wait this long for components to render. Many
			// components will NOT render within this time, but that's
			// OK: they will simply be rendered on the browser. It's
			// good to keep this low to reduce the consequences of
			// unexpected jsserver failures (if it totally failed and
			// the timeout was 5s, every page load would take at least
			// 5s), at least until we get more familiar with the ops
			// behavior of it.
			const wait = 1 * time.Second

			ctx, cancel := context.WithTimeout(ctx, wait)
			defer cancel()

			start := time.Now()
			var err error
			componentHTML, err = renderReactComponent(ctx, component, propMap, stores)
			if err != nil {
				propsSummary := propsJSON
				if len(propsSummary) > 200 {
					propsSummary = propsSummary[:200]
				}
				if err == context.DeadlineExceeded {
					log15.Warn("Rendering React component on the server timed out.", "elapsed", time.Since(start), "component", component, "props", string(propsSummary))
				} else {
					log15.Error("Error rendering React component on the server (falling back to client-side rendering)", "err", err, "component", component, "props", string(propsSummary), "hasPreloadedStoreData", stores != nil)
				}
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
