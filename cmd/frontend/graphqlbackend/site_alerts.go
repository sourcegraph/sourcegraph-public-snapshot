package graphqlbackend

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/pkg/updatecheck"

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
		// 🚨 SECURITY: Only the site admin cares about this. The only time a user should receive a site alert is if
		// sourcegraph is very out of date and or basic setup is still needed (https/ external URL)
		if !args.IsSiteAdmin {
			// users will see alerts if Sourcegraph is >4 months out of date
			alert := outOfDateAlert(args.IsSiteAdmin)
			if alert == nil {
				return nil
			}
			return []*Alert{alert}
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

		alerts := make([]*Alert, 0, 3)

		alert := outOfDateAlert(args.IsSiteAdmin)
		if alert == nil {
			alerts = append(alerts, alert)
		}

		if len(problems) == 0 && len(alerts) == 0 {
			return nil
		}

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

func outOfDateAlert(isAdmin bool) *Alert {

	var alert Alert
	globalUpdateStatus := updatecheck.Last()
	if globalUpdateStatus == nil || updatecheck.IsPending() {
		return &alert
	}
	months := globalUpdateStatus.MonthsOutOfDate
	if months <= 0 {
		return &alert
	}
	// online instances will still be prompt site admins to upgrade via site_update_check
	if months < 3 && !globalUpdateStatus.Offline {
		return &alert
	}

	if isAdmin {
		alert.MessageValue = fmt.Sprintf("Sourcegraph is %d months out of date", months)
		key := fmt.Sprintf("months-out-of-date-%d", months)
		switch {
		case months < 3:
			return &Alert{
				TypeValue:                 AlertTypeInfo,
				MessageValue:              alert.MessageValue,
				IsDismissibleWithKeyValue: key,
			}
		case months == 3:
			return &Alert{
				TypeValue:    AlertTypeWarning,
				MessageValue: "Sourcegraph is 3 months out of date, at 4 months users will be warned Sourcegraph is out of date",
			}
		case months == 4:
			return &Alert{
				TypeValue:    AlertTypeWarning,
				MessageValue: "Sourcegraph is 4+ months out of date, for the latest features and bug fixes please upgrade",
			}

		case months == 5:
			return &Alert{
				TypeValue:    AlertTypeError,
				MessageValue: "Sourcegraph is 5+ months out of date, for the latest features and bug fixes please upgrade.",
			}

		default:
			return &Alert{
				TypeValue:    AlertTypeError,
				MessageValue: "Sourcegraph is 6+ months out of date, for the latest features and bug fixes please upgrade.",
			}
		}
	}
	if months <= 3 {
		return &alert
	}
	alert.MessageValue = fmt.Sprintf("Sourcegraph is %d+ months out of date, "+
		"for the latest features and bug fixes ask your site administrator to upgrade.", months)
	key := fmt.Sprintf("months-out-of-date-%d", months)
	switch months {
	case 4:
		return &Alert{
			TypeValue:                 AlertTypeWarning,
			MessageValue:              alert.MessageValue,
			IsDismissibleWithKeyValue: key,
		}
	case 5:
		return &Alert{
			TypeValue:                 AlertTypeError,
			MessageValue:              alert.MessageValue,
			IsDismissibleWithKeyValue: key,
		}

	default:
		return &Alert{
			TypeValue:    AlertTypeError,
			MessageValue: alert.MessageValue,
		}
	}
}
