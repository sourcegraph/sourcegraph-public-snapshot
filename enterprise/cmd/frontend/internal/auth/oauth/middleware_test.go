package oauth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func Test_getExtraScopes(t *testing.T) {
	envvar.MockSourcegraphDotComMode(true)
	u := database.NewStrictMockUserStore()
	u.CurrentUserAllowedExternalServicesFunc.SetDefaultReturn(conf.ExternalServiceModeAll, nil)
	db := database.NewStrictMockDB()
	db.UsersFunc.SetDefaultReturn(u)
	for name, test := range map[string]struct {
		operation, provider string
		scopes              []string
	}{
		"withoutScopes_gitlab": {"", extsvc.TypeGitLab, []string{}},
		"withoutScopes_github": {"", extsvc.TypeGitHub, []string{}},
		"withScopes_gitlab":    {"createCodeHostConnection", extsvc.TypeGitLab, []string{"api"}},
		"withScopes_github":    {"createCodeHostConnection", extsvc.TypeGitHub, []string{"repo"}},
	} {
		t.Run(name, func(t *testing.T) {

			got, err := getExtraScopes(context.Background(), db, test.provider, LoginStateOp(test.operation))
			if err != nil {
				t.Fatal(err)
			}
			if !assert.ElementsMatch(t, got, test.scopes) {
				t.Errorf("Expected %v got %v", test.scopes, got)
			}
		})
	}
}
