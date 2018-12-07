package authz

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func init() {
	// Warn about usage of auth providers that are not enabled by the license.
	graphqlbackend.AlertFuncs = append(graphqlbackend.AlertFuncs, func(args graphqlbackend.AlertFuncArgs) []*graphqlbackend.Alert {
		// Only site admins can act on this alert, so only show it to site admins.
		if !args.IsSiteAdmin {
			return nil
		}

		if licensing.IsFeatureEnabledLenient(licensing.FeatureACLs) {
			return nil
		}

		var authzTypes []string
		for _, g := range conf.Get().Github {
			if g.Authorization != nil {
				authzTypes = append(authzTypes, "GitHub")
				break
			}
		}
		for _, g := range conf.Get().Gitlab {
			if g.Authorization != nil {
				authzTypes = append(authzTypes, "GitLab")
				break
			}
		}
		if len(authzTypes) > 0 {
			return []*graphqlbackend.Alert{{
				TypeValue:    graphqlbackend.AlertTypeError,
				MessageValue: fmt.Sprintf("A Sourcegraph license is required to enable repository permissions for the following code hosts: %s. [**Get a license.**](/site-admin/license)", strings.Join(authzTypes, ", ")),
			}}
		}
		return nil
	})
}
