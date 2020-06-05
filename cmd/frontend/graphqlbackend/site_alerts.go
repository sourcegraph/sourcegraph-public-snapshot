package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/pkg/updatecheck"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
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

// Intentionally named "DISABLE_SECURITY" and not something else, so that anyone considering
// disabling this thinks twice about the risks associated with disabling these and considers
// keeping up-to-date more frequently instead.
var disableSecurity, _ = strconv.ParseBool(env.Get("DISABLE_SECURITY", "false", "disables security upgrade notices"))

func init() {
	conf.ContributeWarning(func(c conf.Unified) (problems conf.Problems) {
		if c.ExternalURL == "" {
			problems = append(problems, conf.NewSiteProblem("`externalURL` is required to be set for many features of Sourcegraph to work correctly."))
		} else if conf.DeployType() != conf.DeployDev && strings.HasPrefix(c.ExternalURL, "http://") {
			problems = append(problems, conf.NewSiteProblem("Your connection is not private. We recommend [configuring Sourcegraph to use HTTPS/SSL](https://docs.sourcegraph.com/admin/nginx)"))
		}

		return problems
	})

	if !disableSecurity {
		// Warn about Sourcegraph being out of date.
		AlertFuncs = append(AlertFuncs, outOfDateAlert)
	} else {
		log15.Warn("WARNING: SECURITY NOTICES DISABLED: this is not recommended, please unset DISABLE_SECURITY=true")
	}

	// Warn about invalid site configuration.
	AlertFuncs = append(AlertFuncs, func(args AlertFuncArgs) []*Alert {
		// 🚨 SECURITY: Only the site admin cares about this. The only time a user should receive a site alert is if
		// sourcegraph is very out of date and or basic setup is still needed (https/ external URL)
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

func outOfDateAlert(args AlertFuncArgs) []*Alert {
	globalUpdateStatus := updatecheck.Last()
	if globalUpdateStatus == nil || updatecheck.IsPending() {
		return nil
	}
	alert := determineOutOfDateAlert(args.IsSiteAdmin, globalUpdateStatus.MonthsOutOfDate, globalUpdateStatus.Offline)
	if alert == nil {
		return nil
	}
	return []*Alert{alert}
}

func determineOutOfDateAlert(isAdmin bool, months int, offline bool) *Alert {
	if months <= 0 {
		return nil
	}
	// online instances will still be prompt site admins to upgrade via site_update_check
	if months < 3 && !offline {
		return nil
	}

	if isAdmin {
		key := fmt.Sprintf("months-out-of-date-%d", months)
		switch {
		case months < 3:
			message := fmt.Sprintf("Sourcegraph is %d+ months out of date, for the latest features and bug fixes please upgrade ([changelog](http://about.sourcegraph.com/changelog))", months)
			return &Alert{TypeValue: AlertTypeInfo, MessageValue: message, IsDismissibleWithKeyValue: key}
		case months == 3:
			message := "Sourcegraph is 3+ months out of date, you may be missing important security or bug fixes. Users will be notified at 4+ months. ([changelog](http://about.sourcegraph.com/changelog))"
			return &Alert{TypeValue: AlertTypeWarning, MessageValue: message}
		case months == 4:
			message := "Sourcegraph is 4+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"
			return &Alert{TypeValue: AlertTypeWarning, MessageValue: message}
		case months == 5:
			message := "Sourcegraph is 5+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"
			return &Alert{TypeValue: AlertTypeError, MessageValue: message}
		default:
			message := fmt.Sprintf("Sourcegraph is %d+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))", months)
			return &Alert{TypeValue: AlertTypeError, MessageValue: message}
		}
	}

	key := fmt.Sprintf("months-out-of-date-%d", months)
	switch months {
	case 0, 1, 2, 3:
		return nil
	case 4, 5:
		message := fmt.Sprintf("Sourcegraph is %d+ months out of date, ask your site administrator to upgrade for the latest features and bug fixes. ([changelog](http://about.sourcegraph.com/changelog))", months)
		return &Alert{TypeValue: AlertTypeWarning, MessageValue: message, IsDismissibleWithKeyValue: key}
	default:
		message := fmt.Sprintf("Sourcegraph is %d+ months out of date, you may be missing important security or bug fixes. Ask your site administrator to upgrade. ([changelog](http://about.sourcegraph.com/changelog))", months)
		return &Alert{TypeValue: AlertTypeError, MessageValue: message, IsDismissibleWithKeyValue: key}
	}
}
