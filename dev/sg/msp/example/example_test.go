package example

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
)

func mockNewProjectID(t *testing.T) {
	templateFuncs[newProjectIDFuncKey] = func(s, e string, l int) (string, error) {
		if len(s) == 0 {
			return "", errors.New("service ID is required")
		}
		if len(e) == 0 {
			return "", errors.New("environment ID is required")
		}
		if l == 0 {
			return "", errors.New("expected length > 0")
		}
		return fmt.Sprintf("%s-%s-%s", s, e, strings.Repeat("x", l)), nil
	}
	t.Cleanup(func() { templateFuncs[newProjectIDFuncKey] = spec.NewProjectID })
}

func TestNewService(t *testing.T) {
	mockNewProjectID(t)

	for _, tc := range []struct {
		name     string
		template Template
	}{
		{
			name: "dev",
			template: Template{
				ID:    "msp-example",
				Dev:   true,
				Owner: "core-services",

				ProjectIDSuffixLength: 4,
			},
		},
		{
			name: "prod",
			template: Template{
				ID:    "msp-example",
				Dev:   false,
				Owner: "core-services",

				ProjectIDSuffixLength: 4,
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testRender(t, NewService, tc.template)
		})
	}
}

func TestNewJob(t *testing.T) {
	mockNewProjectID(t)

	for _, tc := range []struct {
		name     string
		template Template
	}{
		{
			name: "dev",
			template: Template{
				ID:    "msp-example",
				Dev:   true,
				Owner: "core-services",

				ProjectIDSuffixLength: 4,
			},
		},
		{
			name: "prod",
			template: Template{
				ID:    "msp-example",
				Dev:   false,
				Owner: "core-services",

				ProjectIDSuffixLength: 4,
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testRender(t, NewJob, tc.template)
		})
	}
}

func testRender(t *testing.T, renderFn func(t Template) ([]byte, error), template Template) {
	f, err := renderFn(template)
	require.NoError(t, err)

	t.Run("spec", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(string(f)))
	})

	t.Run("is valid", func(t *testing.T) {
		var s spec.Spec
		require.NoError(t, yaml.Unmarshal(f, &s))
		assert.Empty(t, s.Validate())
	})

	t.Run("insert environment", func(t *testing.T) {
		e, err := NewEnvironment(EnvironmentTemplate{
			ServiceID:             template.ID,
			EnvironmentID:         "second",
			ProjectIDSuffixLength: template.ProjectIDSuffixLength,
		})
		require.NoError(t, err)

		updatedSpecData, err := spec.AppendEnvironment(f, e)
		require.NoError(t, err)

		autogold.ExpectFile(t, autogold.Raw(string(updatedSpecData)))
	})
}
