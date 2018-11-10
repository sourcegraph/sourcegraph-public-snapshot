package parse

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/schema"
)

// legacyEnvToFieldName maps from the legacy env var name to the
// SiteConfiguration struct field that the env var sets. Only top-level fields
// are traversed, so we needn't worry about nested names or collisions.
var legacyEnvToFieldName = map[string]string{
	"ExternalURL": "SRC_APP_URL",
	// AuthUserOrgMap has no env var
	"CorsOrigin": "CORS_ORIGIN",
	// DisablePublicRepoRedirects is handled specially (inverted PUBLIC_REPO_REDIRECTS)
	"ExecuteGradleOriginalRootPaths": "EXECUTE_GRADLE_ORIGINAL_ROOT_PATHS",
	"Github":                         "GITHUB_CONFIG",
	"GithubClientID":                 "GITHUB_CLIENT_ID",
	"GithubClientSecret":             "GITHUB_CLIENT_SECRET",
	"GitMaxConcurrentClones":         "GIT_MAX_CONCURRENT_CLONES",
	"GitoliteHosts":                  "GITOLITE_HOSTS",
	"HtmlBodyBottom":                 "HTML_BODY_BOTTOM",
	"HtmlBodyTop":                    "HTML_BODY_TOP",
	"HtmlHeadBottom":                 "HTML_HEAD_BOTTOM",
	"HtmlHeadTop":                    "HTML_HEAD_TOP",
	"LicenseKey":                     "LICENSE_KEY",
	"LightstepAccessToken":           "LIGHTSTEP_ACCESS_TOKEN",
	"LightstepProject":               "LIGHTSTEP_PROJECT",
	"MaxReposToSearch":               "MAX_REPOS_TO_SEARCH",
	"NoGoGetDomains":                 "NO_GO_GET_DOMAINS",
	"OidcClientID":                   "OIDC_CLIENT_ID",
	"OidcClientSecret":               "OIDC_CLIENT_SECRET",
	"OidcEmailDomain":                "OIDC_EMAIL_DOMAIN",
	"OidcProvider":                   "OIDC_OP",
	"Phabricator":                    "PHABRICATOR_CONFIG",
	"PrivateArtifactRepoID":          "PRIVATE_ARTIFACT_REPO_ID",
	"PrivateArtifactRepoPassword":    "PRIVATE_ARTIFACT_REPO_PASSWORD",
	"PrivateArtifactRepoURL":         "PRIVATE_ARTIFACT_REPO_URL",
	"PrivateArtifactRepoUsername":    "PRIVATE_ARTIFACT_REPO_USERNAME",
	"RepoListUpdateInterval":         "REPO_LIST_UPDATE_INTERVAL",
	"ReposList":                      "REPOS_LIST",
	"SamlIDProviderMetadataURL":      "SAML_ID_PROVIDER_METADATA_URL",
	"SamlSPCert":                     "SAML_CERT",
	"SamlSPKey":                      "SAML_KEY",
	"SearchScopes":                   "SEARCH_SCOPES",
	// Settings has no env var
	"TlsCert": "TLS_CERT",
	"TlsKey":  "TLS_KEY",
}

// configFromLegacyEnvVars constructs site config JSON from env vars. This is merged into the
// SOURCEGRAPH_CONFIG site config JSON.
func configFromEnv() (configJSON []byte, envVarNames []string, err error) {
	var cfg schema.SiteConfiguration

	configType := reflect.TypeOf(cfg)
	configVal := reflect.ValueOf(&cfg)

	for i := 0; i < configType.NumField(); i++ {
		typeField := configType.Field(i)

		var envVal string
		// Read from environment variable with the same name as the JSON tag
		jsonName := typeField.Tag.Get("json")
		jsonName = strings.TrimSuffix(jsonName, ",omitempty")
		envName := strings.Replace(jsonName, ".", "__", -1)
		if envName == "" && typeField.Name != "PublicRepoRedirects" {
			return nil, nil, fmt.Errorf("missing JSON struct tag for config field %s", typeField.Name)
		}

		envVal = os.Getenv(envName)
		if envVal != "" {
			envVarNames = append(envVarNames, envName)
		}

		if envVal == "" {
			// Fall back to reading from legacy environment variable
			if legacyEnvName := legacyEnvToFieldName[typeField.Name]; legacyEnvName != "" {
				envVal = os.Getenv(legacyEnvName)
				if envVal != "" {
					envVarNames = append(envVarNames, legacyEnvName)
				}
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
					return nil, nil, fmt.Errorf("could not parse value for field %s: %s", typeField.Name, err)
				}
				valField.SetBool(valBool)
			case reflect.Int:
				valInt, err := strconv.Atoi(envVal)
				if err != nil {
					return nil, nil, fmt.Errorf("could not parse value for field %s: %s", typeField.Name, err)
				}
				valField.SetInt(int64(valInt))
			case reflect.Ptr, reflect.Slice, reflect.Struct:
				if err := json.Unmarshal([]byte(envVal), valField.Addr().Interface()); err != nil {
					return nil, nil, fmt.Errorf("could not parse value for field %s: %s", typeField.Name, err)
				}
			default:
				return nil, nil, fmt.Errorf("unhandled config field type: %s (field: %s)", valField.Kind(), typeField.Name)
			}
		}

	}
	// Special case for PUBLIC_REPO_REDIRECTS
	if prd := os.Getenv("PUBLIC_REPO_REDIRECTS"); prd != "" {
		if publicRepoRedirects, err := strconv.ParseBool(prd); err == nil {
			cfg.DisablePublicRepoRedirects = !publicRepoRedirects
		}
	}

	configJSON, err = json.Marshal(cfg)
	return configJSON, envVarNames, err
}
