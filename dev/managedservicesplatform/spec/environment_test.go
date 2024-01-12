package spec

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func errorMessages(errs []error) []string {
	var messages []string
	for _, e := range errs {
		messages = append(messages, e.Error())
	}
	return messages
}

func TestEnvironmentResourcePostgreSQLSpecValidate(t *testing.T) {
	for _, tc := range []struct {
		name       string
		spec       *EnvironmentResourcePostgreSQLSpec
		wantErrors autogold.Value
	}{
		{
			name:       "nil",
			spec:       nil,
			wantErrors: nil,
		},
		{
			name:       "defaults",
			spec:       &EnvironmentResourcePostgreSQLSpec{},
			wantErrors: nil,
		},
		{
			name: "odd CPU",
			spec: &EnvironmentResourcePostgreSQLSpec{
				CPU: pointers.Ptr(3),
			},
			wantErrors: autogold.Expect([]string{"postgreSQL.cpu must be 1 or a multiple of 2"}),
		},
		{
			name: "too little memory for CPU",
			spec: &EnvironmentResourcePostgreSQLSpec{
				CPU:      pointers.Ptr(12),
				MemoryGB: pointers.Ptr(4),
			},
			wantErrors: autogold.Expect([]string{"postgreSQL.memoryGB must be >= postgreSQL.cpu"}),
		},
		{
			name: "too much memory for CPU",
			spec: &EnvironmentResourcePostgreSQLSpec{
				MemoryGB: pointers.Ptr(12),
			},
			wantErrors: autogold.Expect([]string{"postgreSQL.memoryGB must be <= 6*postgreSQL.cpu"}),
		},
		{
			name: "odd CPU, too much memory for CPU",
			spec: &EnvironmentResourcePostgreSQLSpec{
				CPU:      pointers.Ptr(5),
				MemoryGB: pointers.Ptr(50),
			},
			wantErrors: autogold.Expect([]string{
				"postgreSQL.cpu must be 1 or a multiple of 2",
				"postgreSQL.memoryGB must be <= 6*postgreSQL.cpu",
			}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.spec.Validate()
			if tc.wantErrors == nil {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
				tc.wantErrors.Equal(t, errorMessages(errs))
			}
		})
	}
}

func TestEnvironmentInstancesResourcesSpecValdiate(t *testing.T) {
	for _, tc := range []struct {
		name       string
		spec       *EnvironmentInstancesResourcesSpec
		wantErrors autogold.Value
	}{
		{
			name:       "nil",
			spec:       nil,
			wantErrors: nil,
		},
		{
			name: "cpu, memory too low",
			spec: &EnvironmentInstancesResourcesSpec{
				CPU:    0,
				Memory: "256MiB",
			},
			wantErrors: autogold.Expect([]string{"resources.cpu must be >= 1", "resources.memory must be >= 512MiB"}),
		},
		{
			name: "cpu, memory too high",
			spec: &EnvironmentInstancesResourcesSpec{
				CPU:    10,
				Memory: "60GiB",
			},
			wantErrors: autogold.Expect([]string{
				"resources.cpu > 8 not supported - considering decreasing scaling.maxRequestConcurrency and increasing scaling.maxCount instead",
				"resources.memory > 32GiB not supported - considering decreasing scaling.maxRequestConcurrency and increasing scaling.maxCount instead",
			}),
		},
		{
			name: "cpu too high for memory",
			spec: &EnvironmentInstancesResourcesSpec{
				CPU:    8,
				Memory: "1GiB",
			},
			wantErrors: autogold.Expect([]string{"resources.cpu > 6 requires resources.memory >= 4GiB"}),
		},
		{
			name: "memory too high for cpu",
			spec: &EnvironmentInstancesResourcesSpec{
				CPU:    1,
				Memory: "32GiB",
			},
			wantErrors: autogold.Expect([]string{"resources.memory > 24GiB requires resources.cpu >= 8"}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.spec.Validate()
			if tc.wantErrors == nil {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
				tc.wantErrors.Equal(t, errorMessages(errs))
			}
		})
	}
}
