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
