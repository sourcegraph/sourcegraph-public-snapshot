package parse

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

type SiteConfiguration struct {
	Basic *schema.BasicSiteConfiguration `json:"basic"`
	Core  *schema.CoreSiteConfiguration  `json:"core"`
}

func ParseBasic(data string) (*schema.BasicSiteConfiguration, error) {
	var basic schema.BasicSiteConfiguration

	err := tolerantUnmarshal(data, &basic)
	if err != nil {
		return nil, err
	}

	// For convenience, make sure this is not nil.
	if basic.ExperimentalFeatures == nil {
		basic.ExperimentalFeatures = &schema.ExperimentalFeatures{}
	}

	return &basic, nil
}

func DeprecatedParseBasicConfigFromEnvironment(data string) (*schema.BasicSiteConfiguration, error) {
	var tmpConfig, err = ParseBasic(data)
	if err != nil {
		return nil, err
	}

	// Env var config takes highest precedence but is deprecated.
	if v, envVarNames, err := basicConfigFromEnv(); err != nil {
		return nil, err
	} else if len(envVarNames) > 0 {
		if err := json.Unmarshal(v, tmpConfig); err != nil {
			return nil, err
		}
	}
	return tmpConfig, nil
}

func ParseCore(data string) (*schema.CoreSiteConfiguration, error) {
	var core schema.CoreSiteConfiguration

	err := tolerantUnmarshal(data, &core)
	if err != nil {
		return nil, err
	}

	return &core, nil
}

func DeprecatedParseCoreConfigFromEnvironment(data string) (*schema.CoreSiteConfiguration, error) {
	var tmpConfig, err = ParseCore(data)
	if err != nil {
		return nil, err
	}

	// Env var config takes highest precedence but is deprecated.
	if v, envVarNames, err := coreConfigFromEnv(); err != nil {
		return nil, err
	} else if len(envVarNames) > 0 {
		if err := json.Unmarshal(v, tmpConfig); err != nil {
			return nil, err
		}
	}
	return tmpConfig, nil
}

func tolerantUnmarshal(data string, v interface{}) error {
	if data == "" {
		return nil
	}

	massagedData, err := jsonc.Parse(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(massagedData, v)
}

// requireRestart describes the list of config properties that require
// restarting the Sourcegraph Server in order for the change to take effect.
//
// Experimental features are special in that they are denoted individually
// via e.g. "experimentalFeatures::myFeatureFlag".
var requireRestart = []string{
	"siteID",
	"executeGradleOriginalRootPaths",
	"lightstepAccessToken",
	"lightstepProject",
	"auth.accessTokens",
	"privateArtifactRepoURL",
	"auth.userOrgMap",
	"auth.sessionExpiry",
	"noGoGetDomains",
	"auth.disableAccessTokens",
	"auth.providers",
	"appURL",
	"tls.letsencrypt",
	"git.cloneURLToRepositoryName",
	"searchScopes",
	"extensions",
	"disableBrowserExtension",
	"tlsCert",
	"update.channel",
	"useJaeger",
	"privateArtifactRepoPassword",
	"disablePublicRepoRedirects",
	"privateArtifactRepoUsername",
	"blacklistGoGet",
	"privateArtifactRepoID",
	"tlsKey",
}

// merge a map, overwriting keys
func mergeMap(destMap, srcMap reflect.Value) {
	mapType := destMap.Type()
	if mapType.Kind() != reflect.Map {
		fmt.Printf("error: not a map: %T\n", destMap)
		return
	}
	valueType := mapType.Elem()
	zero := reflect.Zero(valueType)
	keys := srcMap.MapKeys()
	for _, key := range keys {
		srcValue := srcMap.MapIndex(key)
		destValue := destMap.MapIndex(key)
		switch srcValue.Kind() {
		case reflect.Struct:
			if destValue.IsNil() {
				destMap.SetMapIndex(key, srcValue)
			} else {
				mergeStruct(destValue.Interface(), srcValue.Interface())
			}
		case reflect.Slice:
			destMap.SetMapIndex(key, reflect.AppendSlice(destValue, srcValue))
		case reflect.Map:
			mergeMap(destValue, srcValue)
		default:
			if srcValue.Interface() != zero.Interface() {
				destMap.SetMapIndex(key, srcValue)
			}
		}
		destMap.SetMapIndex(key, srcMap.MapIndex(key))
	}
}

// merge a struct. recurse on structs, append arrays,
// overwrite everything else.
func mergeStruct(destInterface, srcInterface interface{}) {
	destType := reflect.TypeOf(destInterface)
	dest := reflect.ValueOf(destInterface)
	if destType.Kind() == reflect.Ptr {
		dest = dest.Elem()
		destType = dest.Type()
	}
	srcType := reflect.TypeOf(srcInterface)
	src := reflect.ValueOf(srcInterface)
	if srcType.Kind() == reflect.Ptr {
		src = src.Elem()
		srcType = src.Type()
	}
	if destType != srcType {
		fmt.Printf("fatal: destType '%T' and srcType '%T' are not equal.\n", dest, src)
		return
	}
	for i := 0; i < destType.NumField(); i++ {
		destField := dest.Field(i)
		srcField := src.Field(i)
		zero := reflect.Zero(destField.Type())
		switch destField.Kind() {
		case reflect.Struct:
			mergeStruct(destField, srcField)
		case reflect.Slice:
			destField.Set(reflect.AppendSlice(destField, srcField))
		case reflect.Map:
			mergeMap(destField, srcField)
		case reflect.Ptr:
			switch destField.Elem().Kind() {
			case reflect.Struct:
				srcValid := srcField.Elem().IsValid()
				destValid := destField.Elem().IsValid()
				if srcValid {
					if destValid {
						mergeStruct(destField.Interface(), srcField.Interface())
					} else {
						destField.Set(srcField)
					}
				}
			case reflect.Slice:
				destField.Elem().Set(reflect.AppendSlice(destField.Elem(), srcField.Elem()))
			case reflect.Map:
				mergeMap(destField.Elem(), srcField.Elem())
			}
		default:
			if srcField.Interface() != zero.Interface() {
				destField.Set(srcField)
			}
		}
	}
}

// recursively merge components of site config
func AppendConfig(dest, src *SiteConfiguration) *SiteConfiguration {
	if dest == nil {
		return src
	}
	if src == nil {
		return dest
	}
	mergeStruct(dest, src)
	return dest
}

// NeedRestartToApply determines if a restart is needed to apply the changes
// between the two configurations.
func NeedRestartToApplyBasic(before, after *schema.BasicSiteConfiguration) bool {
	diff := diff(before, after)

	// Check every option that changed to determine whether or not a server
	// restart is required.
	for option := range diff {
		for _, requireRestartOption := range requireRestart {
			if option == requireRestartOption {
				return true
			}
		}
	}
	return false
}

// NeedRestartToApply determines if a restart is needed to apply the changes
// between the two configurations.
func NeedRestartToApplyCore(before, after *schema.CoreSiteConfiguration) bool {
	diff := diff(before, after)

	// Check every option that changed to determine whether or not a server
	// restart is required.
	for option := range diff {
		for _, requireRestartOption := range requireRestart {
			if option == requireRestartOption {
				return true
			}
		}
	}
	return false
}
