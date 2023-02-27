package runner

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"
)

func TestDesugarOperation(t *testing.T) {
	for _, testCase := range []struct {
		name                 string
		operationType        MigrationOperationType
		appliedVersions      []int
		expectedOperation    MigrationOperation
		expectedErrorMessage string
	}{
		{
			name:          "upgrade",
			operationType: MigrationOperationTypeUpgrade,
			expectedOperation: MigrationOperation{
				SchemaName:     "test",
				Type:           MigrationOperationTypeTargetedUp,
				TargetVersions: []int{10003, 10004},
			},
		},
		{
			name:            "revert",
			operationType:   MigrationOperationTypeRevert,
			appliedVersions: []int{10001, 10002, 10003},
			expectedOperation: MigrationOperation{
				SchemaName:     "test",
				Type:           MigrationOperationTypeTargetedDown,
				TargetVersions: []int{10002},
			},
		},
		{
			name:            "revert (again)`",
			operationType:   MigrationOperationTypeRevert,
			appliedVersions: []int{10001, 10002},
			expectedOperation: MigrationOperation{
				SchemaName:     "test",
				Type:           MigrationOperationTypeTargetedDown,
				TargetVersions: []int{10001},
			},
		},
		{
			name:                 "empty revert",
			operationType:        MigrationOperationTypeRevert,
			appliedVersions:      nil,
			expectedErrorMessage: "nothing to revert",
		},
		{
			name:                 "ambiguous revert",
			operationType:        MigrationOperationTypeRevert,
			appliedVersions:      []int{10001, 10002, 10003, 10004},
			expectedErrorMessage: "ambiguous revert",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			schemaContext := schemaContext{
				logger: logtest.Scoped(t),
				schema: makeTestSchema(t, "well-formed"),
				initialSchemaVersion: schemaVersion{
					appliedVersions: testCase.appliedVersions,
				},
			}
			sourceOperation := MigrationOperation{
				SchemaName: "test",
				Type:       testCase.operationType,
			}

			desugaredOperation, err := desugarOperation(schemaContext, sourceOperation)
			if err != nil {
				if testCase.expectedErrorMessage == "" {
					t.Fatalf("unexpected error: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedErrorMessage) {
					t.Fatalf("unexpected error. want=%q have=%v", testCase.expectedErrorMessage, err)
				}
			} else {
				if testCase.expectedErrorMessage != "" {
					t.Fatalf("expected error")
				}

				if diff := cmp.Diff(testCase.expectedOperation, desugaredOperation); diff != "" {
					t.Errorf("unexpected operation (-want +got):\n%s", diff)
				}
			}
		})
	}
}
