package mock

import (
	"cmp"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/contract"
)

type ServiceMetadata struct {
	name    string
	version string
}

type MockError struct {
	error
}

func (s ServiceMetadata) Name() string    { return cmp.Or(s.name, "mock-name") }
func (s ServiceMetadata) Version() string { return cmp.Or(s.version, "mock-version") }

// NewMockContract returns a new contract instance from the given env. If there is an error parsing the given environment
// a MockError is returned that contains the error.
//
// Otherwise a new contract instance is returned as well as the environment validation result from `env.Validate()`
func NewContract(t *testing.T, mockServiceMeta contract.ServiceMetadataProvider, env ...string) (*contract.Contract, error) {
	t.Helper()
	e, err := contract.ParseEnv(env)
	if err != nil {
		return nil, MockError{err}
	}

	c := contract.New(logtest.Scoped(t), mockServiceMeta, e)
	return &c, e.Validate()
}
