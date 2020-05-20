package graphqlbackend

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

// Alert implements the GraphQL type Alert.
type Alert struct {
	TypeValue                 string
	MessageValue              string
	IsDismissibleWithKeyValue string
}

func (r *Alert) Type() string    { return r.TypeValue }
func (r *Alert) Message() string { return r.MessageValue }
func (r *Alert) IsDismissibleWithKey() *string {
	if r.IsDismissibleWithKeyValue == "" {
		return nil
	}
	return &r.IsDismissibleWithKeyValue
}

// Constants for the GraphQL enum AlertType.
const (
	AlertTypeInfo    = "INFO"
	AlertTypeWarning = "WARNING"
	AlertTypeError   = "ERROR"
)

// AlertFuncs is a list of functions called to populate the GraphQL Site.alerts value. It may be
// appended to at init time.
//
// The functions are called each time the Site.alerts value is queried, so they must not block.
var AlertFuncs []func(AlertFuncArgs) []*Alert

// AlertFuncArgs are the arguments provided to functions in AlertFuncs used to populate the GraphQL
// Site.alerts value. They allow the functions to customize the returned alerts based on the
// identity of the viewer (without needing to query for that on their own, which would be slow).
type AlertFuncArgs struct {
	IsAuthenticated bool // whether the viewer is authenticated
	IsSiteAdmin     bool // whether the viewer is a site admin
}

func (r *siteResolver) Alerts(ctx context.Context) ([]*Alert, error) {
	args := AlertFuncArgs{
		IsAuthenticated: actor.FromContext(ctx).IsAuthenticated(),
		IsSiteAdmin:     backend.CheckCurrentUserIsSiteAdmin(ctx) == nil,
	}

	var alerts []*Alert
	for _, f := range AlertFuncs {
		alerts = append(alerts, f(args)...)
	}
	return alerts, nil
}

func init() {
	conf.ContributeWarning(func(c conf.Unified) (problems conf.Problems) {
		if c.ExternalURL == "" {
			problems = append(problems, conf.NewSiteProblem("`externalURL` is required to be set for many features of Sourcegraph to work correctly."))
		} else if conf.DeployType() != conf.DeployDev && strings.HasPrefix(c.ExternalURL, "http://") {
			problems = append(problems, conf.NewSiteProblem("Your connection is not private. We recommend [configuring Sourcegraph to use HTTPS/SSL](https://docs.sourcegraph.com/admin/nginx)"))
		}

		return problems
	})

	// Warn about invalid site configuration.
	AlertFuncs = append(AlertFuncs, func(args AlertFuncArgs) []*Alert {
		// 🚨 SECURITY: Only the site admin cares about this. Leaking a boolean wouldn't be a
		// security vulnerability, but just in case this method is changed to return more
		// information, let's lock it down.
		if !args.IsSiteAdmin {
			return nil
		}

		problems, err := conf.Validate(globals.ConfigurationServerFrontendOnly.Raw())
		if err != nil {
			return []*Alert{
				{
					TypeValue:    AlertTypeError,
					MessageValue: `Update [**site configuration**](/site-admin/configuration) to resolve problems: ` + err.Error(),
				},
			}
		}

		warnings, err := conf.GetWarnings()
		if err != nil {
			return []*Alert{
				{
					TypeValue:    AlertTypeError,
					MessageValue: `Update [**critical configuration**](/help/admin/management_console) to resolve problems: ` + err.Error(),
				},
			}
		}
		problems = append(problems, warnings...)

		if len(problems) == 0 {
			return nil
		}

		alerts := make([]*Alert, 0, 2)

		siteProblems := problems.Site()
		if len(siteProblems) > 0 {
			alerts = append(alerts, &Alert{
				TypeValue: AlertTypeWarning,
				MessageValue: `[**Update site configuration**](/site-admin/configuration) to resolve problems:` +
					"\n* " + strings.Join(siteProblems.Messages(), "\n* "),
			})
		}

		externalServiceProblems := problems.ExternalService()
		if len(externalServiceProblems) > 0 {
			alerts = append(alerts, &Alert{
				TypeValue: AlertTypeWarning,
				MessageValue: `[**Update external service configuration**](/site-admin/external-services) to resolve problems:` +
					"\n* " + strings.Join(externalServiceProblems.Messages(), "\n* "),
			})
		}
		return alerts
	})
}

func OutOfDateAlert(months int, isAdmin bool) Alert {

	if isAdmin {
		switch {
		case months <= 0:
			return Alert{}
		case months == 1:
			return Alert{
				TypeValue:                 AlertTypeInfo,
				MessageValue:              "Sourcegraph is 1 month out of date",
				IsDismissibleWithKeyValue: "y", //TODO What does this mean?
			}
		case months == 2:
			return Alert{
				TypeValue:                 AlertTypeInfo,
				MessageValue:              "Sourcegraph is 2 months out of date",
				IsDismissibleWithKeyValue: "y",
			}
		case months == 3:
			return Alert{
				TypeValue:    AlertTypeWarning,
				MessageValue: "Sourcegraph is 3 months out of date, at 4 months users will be warned Sourcegraph is out of date",
			}
		case months == 4:
			return Alert{
				TypeValue:    AlertTypeWarning,
				MessageValue: "Sourcegraph is 4+ months out of date, for the latest features and bug fixes ask your site administrator to upgrade.",
			}

		case months == 5:
			return Alert{
				TypeValue:    AlertTypeError,
				MessageValue: "Sourcegraph is 5+ months out of date, for the latest features and bug fixes ask your site administrator to upgrade.",
			}

		default:
			return Alert{
				TypeValue:    AlertTypeError,
				MessageValue: "Sourcegraph is 6+ months out of date, for the latest features and bug fixes ask your site administrator to upgrade.",
			}
		}
	}
	switch {
	case months <= 3:
		return Alert{}
	case months == 4:
		return Alert{
			TypeValue:                 AlertTypeWarning,
			MessageValue:              "Sourcegraph is 4+ months out of date, for the latest features and bug fixes ask your site administrator to upgrade.",
			IsDismissibleWithKeyValue: "y", //Should only be dismissible by non-admins
		}

	case months == 5:
		return Alert{
			TypeValue:                 AlertTypeError,
			MessageValue:              "Sourcegraph is 5+ months out of date, for the latest features and bug fixes ask your site administrator to upgrade.",
			IsDismissibleWithKeyValue: "y",
		}

	default:
		return Alert{
			TypeValue:    AlertTypeError,
			MessageValue: "Sourcegraph is 6+ months out of date, for the latest features and bug fixes ask your site administrator to upgrade.",
		}
	}
}
