package conf

import (
	"fmt"
	"log"
	"os"

	"github.com/xeipuuv/gojsonschema"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

// Validate validates the site configuration against its JSON schema.
//
// TODO(sqs): it only validates the SOURCEGRAPH_CONFIG value, not the merged
// config from all env vars. This env var is only used in cmd/server, but it
// is passed onto frontend, so frontend can print useful validation messages
// about it.
func Validate() {
	input := os.Getenv("SOURCEGRAPH_CONFIG")
	if input == "" {
		return
	}
	normalizedInput := normalizeJSON(input)

	res, err := validate([]byte(schema.SiteSchemaJSON), normalizedInput)
	if err != nil {
		log.Printf("Warning: Unable to validate Sourcegraph site configuration: %s", err)
		return
	}
	validationErrors := make([]string, len(res.Errors()))
	for i, e := range res.Errors() {
		validationErrors[i] = e.String()
	}

	customValidationErrors, err := ValidateCustom(normalizedInput)
	if err != nil {
		log.Printf("Warning: Unable to validate Sourcegraph site configuration: %s", err)
		return
	}
	validationErrors = append(validationErrors, customValidationErrors...)

	if len(validationErrors) > 0 {
		fmt.Fprintln(os.Stderr, "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Fprintln(os.Stderr, "⚠️ Warning: Invalid Sourcegraph site configuration:")
		for _, verr := range validationErrors {
			fmt.Fprintf(os.Stderr, " - %s\n", verr)
		}
		fmt.Fprintln(os.Stderr, "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
}

func validate(schema, input []byte) (*gojsonschema.Result, error) {
	return gojsonschema.Validate(
		jsonLoader{gojsonschema.NewBytesLoader(schema)},
		gojsonschema.NewBytesLoader(input),
	)
}

type jsonLoader struct {
	gojsonschema.JSONLoader
}

func (l jsonLoader) LoaderFactory() gojsonschema.JSONLoaderFactory {
	return &jsonLoaderFactory{}
}

type jsonLoaderFactory struct {
	schemas map[string]string
}

func (f jsonLoaderFactory) New(source string) gojsonschema.JSONLoader {
	switch source {
	case "https://sourcegraph.com/v1/settings.schema.json":
		return gojsonschema.NewStringLoader(schema.SettingsSchemaJSON)
	case "https://sourcegraph.com/v1/site.schema.json":
		return gojsonschema.NewStringLoader(schema.SiteSchemaJSON)
	}
	return nil
}
