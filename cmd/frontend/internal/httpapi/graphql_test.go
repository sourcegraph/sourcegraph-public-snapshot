package httpapi

import (
	"context"
	"testing"

	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Test_recordAuditLog(t *testing.T) {
	tests := []struct {
		name                  string
		auditEnabled          bool
		graphQLResponseErrors bool
	}{
		{
			name:         "GraphQL requests aren't audit logged when audit log is not enabled",
			auditEnabled: false,
		},
		{
			name:         "GraphQL requests are audit logged when audit log is enabled",
			auditEnabled: true,
		},
		{
			name:                  "GraphQL requests are marked as failed when the GraphQL response contained errors",
			auditEnabled:          true,
			graphQLResponseErrors: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf.Mock(
				&conf.Unified{
					SiteConfiguration: schema.SiteConfiguration{
						Log: &schema.Log{
							AuditLog: &schema.AuditLog{
								GraphQL: tt.auditEnabled,
							},
						},
					},
				},
			)
			defer conf.Mock(nil)

			logger, exportLogs := logtest.Captured(t)

			ctx := actor.WithActor(context.Background(), actor.FromUser(1))
			recordAuditLog(ctx, logger, traceData{
				queryParams: graphQLQueryParams{
					Query:     `repository(name: "github.com/gorilla/mux") { name }`,
					Variables: map[string]any{"param1": "value1"},
				},
				requestName:   "TestRequest",
				requestSource: "code",
				queryErrors:   makeQueryErrors(tt.graphQLResponseErrors),
			})

			logs := exportLogs()

			if !tt.auditEnabled {
				assert.Equal(t, len(logs), 0)
			} else {
				assert.Equal(t, len(logs), 1)
				auditFields := logs[0].Fields["audit"].(map[string]interface{})
				assert.Equal(t, "GraphQL", auditFields["entity"])
				assert.NotEmpty(t, auditFields["auditId"])

				actorFields := auditFields["actor"].(map[string]interface{})
				assert.NotEmpty(t, actorFields["actorUID"])
				assert.NotEmpty(t, actorFields["ip"])
				assert.NotEmpty(t, actorFields["X-Forwarded-For"])

				requestField := logs[0].Fields["request"].(map[string]interface{})
				assert.Equal(t, requestField["name"], "TestRequest")
				assert.Equal(t, requestField["source"], "code")
				assert.Equal(t, requestField["variables"], `{"param1":"value1"}`)
				assert.Equal(t, requestField["query"], `repository(name: "github.com/gorilla/mux") { name }`)
			}
		})
	}
}

func makeQueryErrors(errors bool) []*gqlerrors.QueryError {
	var result []*gqlerrors.QueryError
	if !errors {
		return result
	}
	result = append(result, &gqlerrors.QueryError{Message: "oops"})
	return result
}
