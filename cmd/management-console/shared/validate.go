package shared

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

// validateConfig returns the first error reported from the list of validators.
func validateConfig(contents string, validators ...func(schema.CriticalConfiguration) error) error {
	c := schema.CriticalConfiguration{}
	if err := jsonc.Unmarshal(contents, &c); err != nil {
		return fmt.Errorf("unmarshal JSON: %v", err)
	}

	for _, v := range validators {
		err := v(c)
		if err != nil {
			return err
		}
	}
	return nil
}

// validateExternalURL checks if "externalURL" in critical config is a non-empty
// valid URL with correct scheme. It returns found errors as a list of strings.
func validateExternalURL(c schema.CriticalConfiguration) error {
	if c.ExternalURL == "" {
		return errors.New(`"externalURL": value cannot be empty`)
	}

	if !strings.HasPrefix(c.ExternalURL, "http://") &&
		!strings.HasPrefix(c.ExternalURL, "https://") {
		return errors.New(`"externalURL": must start with http:// or https://"`)
	}

	if _, err := url.Parse(c.ExternalURL); err != nil {
		return fmt.Errorf(`"externalURL": %v`, err)
	}
	return nil
}
