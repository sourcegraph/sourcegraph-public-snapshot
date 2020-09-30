package campaigns

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/xeipuuv/gojsonschema"

	yamlv3 "gopkg.in/yaml.v3"
)

// SupportedExternalServices are the external service types currently supported
// by the campaigns feature. Repos that are associated with external services
// whose type is not in this list will simply be filtered out from the search
// results.
var SupportedExternalServices = map[string]struct{}{
	extsvc.TypeGitHub:          {},
	extsvc.TypeBitbucketServer: {},
	extsvc.TypeGitLab:          {},
}

// IsRepoSupported returns whether the given ExternalRepoSpec is supported by
// the campaigns feature, based on the external service type.
func IsRepoSupported(spec *api.ExternalRepoSpec) bool {
	_, ok := SupportedExternalServices[spec.ServiceType]
	return ok
}

// IsKindSupported returns whether the given extsvc Kind is supported by
// campaigns.
func IsKindSupported(extSvcKind string) bool {
	_, ok := SupportedExternalServices[extsvc.KindToType(extSvcKind)]
	return ok
}

// Keyer represents items that return a unique key
type Keyer interface {
	Key() string
}

func unixMilliToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

// unmarshalValidate validates the input, which can be YAML or JSON, against
// the provided JSON schema. If the validation is successful is unmarshals the
// validated input into the target.
func unmarshalValidate(schema string, input []byte, target interface{}) error {
	sl := gojsonschema.NewSchemaLoader()
	sc, err := sl.Compile(gojsonschema.NewStringLoader(schema))
	if err != nil {
		return errors.Wrap(err, "failed to compile JSON schema")
	}

	normalized, err := yaml.YAMLToJSONCustom(input, yamlv3.Unmarshal)
	if err != nil {
		return errors.Wrapf(err, "failed to normalize JSON")
	}

	res, err := sc.Validate(gojsonschema.NewBytesLoader(normalized))
	if err != nil {
		return errors.Wrap(err, "failed to validate input against schema")
	}

	var errs *multierror.Error
	for _, err := range res.Errors() {
		e := err.String()
		// Remove `(root): ` from error formatting since these errors are
		// presented to users.
		e = strings.TrimPrefix(e, "(root): ")
		errs = multierror.Append(errs, errors.New(e))
	}

	if err := json.Unmarshal(normalized, target); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
}
