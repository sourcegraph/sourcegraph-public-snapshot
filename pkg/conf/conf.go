package conf

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

// Get returns a copy of the configuration. The returned value should NEVER be modified.
func Get() schema.SiteConfiguration {
	return cfg
}

// cfg is initialized to configuration defaults.
var cfg = schema.SiteConfiguration{
	MaxReposToSearch: 30,
}

func init() {
	// Read env vars to config
	if err := initConfig(); err != nil {
		log.Fatalf("failed to read configuration from environment: %s", err)
	}
}

// initConfig initializes configuration by reading from environment variables. It attempts to read values first from an
// environment variable with the same name as the field's JSON tag and, if that doesn't exist, falls back to reading
// from the environment variable given by the legacy env map. If the field type is a string, the
// value of the environment variable is stored directly. If the field type is an array, struct, bool, or other non-string
// type, the environment variable is unmarshalled into that type.
func initConfig() error {
	configType := reflect.TypeOf(cfg)
	configVal := reflect.ValueOf(&cfg)

	for i := 0; i < configType.NumField(); i++ {
		typeField := configType.Field(i)

		var envVal string
		// Read from environment variable with the same name as the JSON tag
		jsonName := typeField.Tag.Get("json")
		jsonName = strings.TrimSuffix(jsonName, ",omitempty")
		if jsonName == "" && typeField.Name != "PublicRepoRedirects" {
			return fmt.Errorf("missing JSON struct tag for config field %s", typeField.Name)
		}
		envVal = os.Getenv(jsonName)

		if envVal == "" {
			// Fall back to reading from legacy environment variable
			if legacyEnvName := legacyEnvToFieldName[typeField.Name]; legacyEnvName != "" {
				envVal = os.Getenv(legacyEnvName)
			}
		}

		// Set config value
		if envVal != "" {
			valField := configVal.Elem().FieldByName(typeField.Name)
			switch valField.Kind() {
			case reflect.String:
				valField.SetString(envVal)
			case reflect.Bool:
				valBool, err := strconv.ParseBool(envVal)
				if err != nil {
					return fmt.Errorf("could not parse value for field %s: %s", typeField.Name, err)
				}
				valField.SetBool(valBool)
			case reflect.Int:
				valInt, err := strconv.Atoi(envVal)
				if err != nil {
					return fmt.Errorf("could not parse value for field %s: %s", typeField.Name, err)
				}
				valField.SetInt(int64(valInt))
			case reflect.Slice:
				fallthrough
			case reflect.Struct:
				if err := json.Unmarshal([]byte(envVal), valField.Addr().Interface()); err != nil {
					return fmt.Errorf("could not parse value for field %s: %s", typeField.Name, err)
				}
			default:
				return fmt.Errorf("unhandled config field type: %s", valField.Kind())
			}
		}

	}
	// Special case for PUBLIC_REPO_REDIRECTS
	if prd := os.Getenv("PUBLIC_REPO_REDIRECTS"); prd != "" {
		if publicRepoRedirects, err := strconv.ParseBool(prd); err == nil {
			cfg.DisablePublicRepoRedirects = !publicRepoRedirects
		}
	}

	return nil
}
