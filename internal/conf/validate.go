package conf

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"

	"github.com/sourcegraph/sourcegraph/internal/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

// ignoreLegacyKubernetesFields is the set of field names for which validation errors should be
// ignored. The validation errors occur only because deploy-sourcegraph config merged site config
// and Kubernetes cluster-specific config. This is deprecated. Until we have transitioned fully, we
// suppress validation errors on these fields.
var ignoreLegacyKubernetesFields = map[string]struct{}{
	"alertmanagerConfig":    {},
	"alertmanagerURL":       {},
	"authProxyIP":           {},
	"authProxyPassword":     {},
	"deploymentOverrides":   {},
	"gitoliteIP":            {},
	"gitserverCount":        {},
	"gitserverDiskSize":     {},
	"gitserverSSH":          {},
	"httpNodePort":          {},
	"httpsNodePort":         {},
	"indexedSearchDiskSize": {},
	"langGo":                {},
	"langJava":              {},
	"langJavaScript":        {},
	"langPHP":               {},
	"langPython":            {},
	"langSwift":             {},
	"langTypeScript":        {},
	"namespace":             {},
	"nodeSSDPath":           {},
	"phabricatorIP":         {},
	"prometheus":            {},
	"pyPIIP":                {},
	"rbac":                  {},
	"storageClass":          {},
	"useAlertManager":       {},
}

const RedactedSecret = "REDACTED"

// problemKind represents the kind of a configuration problem.
type problemKind string

const (
	problemSite            problemKind = "SiteConfig"
	problemExternalService problemKind = "ExternalService"
)

// Problem contains kind and description of a specific configuration problem.
type Problem struct {
	kind        problemKind
	description string
}

// NewSiteProblem creates a new site config problem with given message.
func NewSiteProblem(msg string) *Problem {
	return &Problem{
		kind:        problemSite,
		description: msg,
	}
}

// NewExternalServiceProblem creates a new external service config problem with given message.
func NewExternalServiceProblem(msg string) *Problem {
	return &Problem{
		kind:        problemExternalService,
		description: msg,
	}
}

// IsSite returns true if the problem is about site config.
func (p Problem) IsSite() bool {
	return p.kind == problemSite
}

// IsExternalService returns true if the problem is about external service config.
func (p Problem) IsExternalService() bool {
	return p.kind == problemExternalService
}

func (p Problem) String() string {
	return p.description
}

func (p *Problem) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"kind":        string(p.kind),
		"description": p.description,
	})
}

func (p *Problem) UnmarshalJSON(b []byte) error {
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	p.kind = problemKind(m["kind"])
	p.description = m["description"]
	return nil
}

// Problems is a list of problems.
type Problems []*Problem

// newProblems converts a list of messages with their kind into problems.
func newProblems(kind problemKind, messages ...string) Problems {
	problems := make([]*Problem, len(messages))
	for i := range messages {
		problems[i] = &Problem{
			kind:        kind,
			description: messages[i],
		}
	}
	return problems
}

// NewSiteProblems converts a list of messages into site config problems.
func NewSiteProblems(messages ...string) Problems {
	return newProblems(problemSite, messages...)
}

// NewExternalServiceProblems converts a list of messages into external service config problems.
func NewExternalServiceProblems(messages ...string) Problems {
	return newProblems(problemExternalService, messages...)
}

// Messages returns the list of problems in strings.
func (ps Problems) Messages() []string {
	if len(ps) == 0 {
		return nil
	}

	msgs := make([]string, len(ps))
	for i := range ps {
		msgs[i] = ps[i].String()
	}
	return msgs
}

// Site returns all site config problems in the list.
func (ps Problems) Site() (problems Problems) {
	for i := range ps {
		if ps[i].IsSite() {
			problems = append(problems, ps[i])
		}
	}
	return problems
}

// ExternalService returns all external service config problems in the list.
func (ps Problems) ExternalService() (problems Problems) {
	for i := range ps {
		if ps[i].IsExternalService() {
			problems = append(problems, ps[i])
		}
	}
	return problems
}

// Validate validates the configuration against the JSON Schema and other
// custom validation checks.
func Validate(input conftypes.RawUnified) (problems Problems, err error) {
	siteProblems, err := doValidate(input.Site, schema.SiteSchemaJSON)
	if err != nil {
		return nil, err
	}
	problems = append(problems, NewSiteProblems(siteProblems...)...)

	customProblems, err := validateCustomRaw(conftypes.RawUnified{
		Site: string(jsonc.Normalize(input.Site)),
	})
	if err != nil {
		return nil, err
	}
	problems = append(problems, customProblems...)
	return problems, nil
}

// ValidateSite is like Validate, except it only validates the site configuration.
func ValidateSite(input string) (messages []string, err error) {
	raw := Raw()
	unredacted, err := UnredactSecrets(input, raw)
	if err != nil {
		return nil, err
	}
	raw.Site = unredacted

	problems, err := Validate(raw)
	if err != nil {
		return nil, err
	}
	return problems.Messages(), nil
}

func UnredactSecrets(input string, raw conftypes.RawUnified) (string, error) {
	newCfg, err := ParseConfig(conftypes.RawUnified{
		Site: input,
	})
	if err != nil {
		return input, err
	}
	oldCfg, err := ParseConfig(raw)
	if err != nil {
		return input, err
	}
	oldSecrets := getSecretsMap(oldCfg)
	for _, ap := range newCfg.AuthProviders {
		if ap.Openidconnect != nil && ap.Openidconnect.ClientSecret == RedactedSecret {
			ap.Openidconnect.ClientSecret = oldSecrets[ap.Openidconnect.ClientID]
		}
		if ap.Github != nil && ap.Github.ClientSecret == RedactedSecret {
			ap.Github.ClientSecret = oldSecrets[ap.Github.ClientID]
		}
		if ap.Gitlab != nil && ap.Gitlab.ClientSecret == RedactedSecret {
			ap.Gitlab.ClientSecret = oldSecrets[ap.Gitlab.ClientID]
		}
	}
	unredactedSite, err := jsonc.Edit(input, newCfg.AuthProviders, "auth.providers")
	if err != nil {
		return input, err
	}
	if newCfg.ExecutorsAccessToken == RedactedSecret {
		unredactedSite, err = jsonc.Edit(unredactedSite, oldCfg.ExecutorsAccessToken, "executors.accessToken")
	}
	return unredactedSite, err
}

func getSecretsMap(cfg *Unified) map[string]string {
	secretsMap := make(map[string]string, len(cfg.AuthProviders))
	for _, ap := range cfg.AuthProviders {
		if ap.Openidconnect != nil {
			secretsMap[ap.Openidconnect.ClientID] = ap.Openidconnect.ClientSecret
		}
		if ap.Github != nil {
			secretsMap[ap.Github.ClientID] = ap.Github.ClientSecret
		}
		if ap.Gitlab != nil {
			secretsMap[ap.Gitlab.ClientID] = ap.Gitlab.ClientSecret
		}
	}
	return secretsMap
}

func RedactSecrets(raw conftypes.RawUnified) (conftypes.RawUnified, error) {
	cfg, err := ParseConfig(raw)
	if err != nil {
		return raw, err
	}
	for _, ap := range cfg.AuthProviders {
		if ap.Openidconnect != nil {
			ap.Openidconnect.ClientSecret = RedactedSecret
		}
		if ap.Github != nil {
			ap.Github.ClientSecret = RedactedSecret
		}
		if ap.Gitlab != nil {
			ap.Gitlab.ClientSecret = RedactedSecret
		}
	}
	newSite, err := jsonc.Edit(raw.Site, cfg.AuthProviders, "auth.providers")
	if err != nil {
		return raw, err
	}
	if cfg.ExecutorsAccessToken != "" {
		newSite, err = jsonc.Edit(newSite, RedactedSecret, "executors.accessToken")
	}
	return conftypes.RawUnified{
		Site: newSite,
	}, err
}

func ValidateSetting(input string) (problems []string, err error) {
	return doValidate(input, schema.SettingsSchemaJSON)
}

func doValidate(inputStr, schema string) (messages []string, err error) {
	input := jsonc.Normalize(inputStr)

	res, err := validate([]byte(schema), input)
	if err != nil {
		return nil, err
	}
	messages = make([]string, 0, len(res.Errors()))
	for _, e := range res.Errors() {
		if _, ok := ignoreLegacyKubernetesFields[e.Field()]; ok {
			continue
		}

		var keyPath string
		if c := e.Context(); c != nil {
			keyPath = strings.TrimPrefix(e.Context().String("."), "(root).")
		} else {
			keyPath = e.Field()
		}

		messages = append(messages, fmt.Sprintf("%s: %s", keyPath, e.Description()))
	}
	return messages, nil
}

func validate(schema, input []byte) (*gojsonschema.Result, error) {
	s, err := gojsonschema.NewSchema(jsonLoader{gojsonschema.NewBytesLoader(schema)})
	if err != nil {
		return nil, err
	}
	return s.Validate(gojsonschema.NewBytesLoader(input))
}

type jsonLoader struct {
	gojsonschema.JSONLoader
}

func (l jsonLoader) LoaderFactory() gojsonschema.JSONLoaderFactory {
	return &jsonLoaderFactory{}
}

type jsonLoaderFactory struct{}

func (f jsonLoaderFactory) New(source string) gojsonschema.JSONLoader {
	switch source {
	case "settings.schema.json":
		return gojsonschema.NewStringLoader(schema.SettingsSchemaJSON)
	case "site.schema.json":
		return gojsonschema.NewStringLoader(schema.SiteSchemaJSON)
	}
	return nil
}

// MustValidateDefaults should be called after all custom validators have been
// registered. It will panic if any of the default deployment configurations
// are invalid.
func MustValidateDefaults() {
	mustValidate("DevAndTesting", confdefaults.DevAndTesting)
	mustValidate("DockerContainer", confdefaults.DockerContainer)
	mustValidate("KubernetesOrDockerComposeOrPureDocker", confdefaults.KubernetesOrDockerComposeOrPureDocker)
}

// mustValidate panics if the configuration does not pass validation.
func mustValidate(name string, cfg conftypes.RawUnified) conftypes.RawUnified {
	problems, err := Validate(cfg)
	if err != nil {
		panic(fmt.Sprintf("Error with %q: %s", name, err))
	}
	if len(problems) > 0 {
		panic(fmt.Sprintf("conf: problems with default configuration for %q:\n  %s", name, strings.Join(problems.Messages(), "\n  ")))
	}
	return cfg
}
