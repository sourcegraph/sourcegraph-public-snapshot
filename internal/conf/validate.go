package conf

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/xeipuuv/gojsonschema"
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

// problemKind represents the kind of a configuration problem.
type problemKind string

const (
	problemCritical        problemKind = "CriticalConfig"
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

// IsCritical returns true if the problem is about critical config.
func (p Problem) IsCritical() bool {
	return p.kind == problemCritical
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

// Critical returns all critical config problems in the list.
func (ps Problems) Critical() (problems Problems) {
	for i := range ps {
		if ps[i].IsCritical() {
			problems = append(problems, ps[i])
		}
	}
	return problems
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
		Critical: string(jsonc.Normalize(input.Critical)),
		Site:     string(jsonc.Normalize(input.Site)),
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
	raw.Site = input

	problems, err := Validate(raw)
	if err != nil {
		return nil, err
	}
	return problems.Messages(), nil
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
	if len(input) > 0 {
		// HACK: Remove the "settings" field from site config because
		// github.com/xeipuuv/gojsonschema has a bug where $ref'd schemas do not always get
		// loaded. When https://github.com/xeipuuv/gojsonschema/pull/196 is merged, it will probably
		// be fixed. This means that the backend config validation will not validate settings, but
		// that is OK because specifying settings here is discouraged anyway.
		var v map[string]interface{}
		if err := json.Unmarshal(input, &v); err != nil {
			return nil, err
		}
		delete(v, "settings")
		var err error
		input, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}

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
	mustValidate("Cluster", confdefaults.Cluster)
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
