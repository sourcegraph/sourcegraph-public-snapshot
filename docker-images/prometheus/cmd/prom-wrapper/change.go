package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/inconshreveable/log15"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

type ChangeContext struct {
	AMConfig *amconfig.Config
}

// ChangeResult indicates output from a Change
type ChangeResult struct {
	Problems conf.Problems
}

// Change implements a change to configuration
type Change func(ctx context.Context, log log15.Logger, change ChangeContext, newConfig *subscribedSiteConfig) (result ChangeResult)

// changeReceivers applies `observability.alerts` as Alertmanager receivers.
func changeReceivers(ctx context.Context, log log15.Logger, change ChangeContext, newConfig *subscribedSiteConfig) (result ChangeResult) {
	// convenience function for creating a prefixed problem - this reflects the relevant site configuration fields
	newProblem := func(err error) {
		result.Problems = append(result.Problems, conf.NewSiteProblem(fmt.Sprintf("`observability.alerts`: %v", err)))
	}

	// generate new notifiers configuration
	change.AMConfig.Receivers = append(newReceivers(newConfig.Alerts, newProblem), &amconfig.Receiver{
		// stub receiver
		Name: alertmanagerNoopReceiver,
	})

	// make sure alerts are routed appropriately
	change.AMConfig.Route = &amconfig.Route{
		Receiver: alertmanagerNoopReceiver,
		Routes: []*amconfig.Route{
			{
				Receiver: alertmanagerWarningReceiver,
				GroupBy:  []model.LabelName{"description"},
				Match: map[string]string{
					"level": "warning",
				},
			},
			{
				Receiver: alertmanagerCriticalReceiver,
				GroupBy:  []model.LabelName{"description"},
				Match: map[string]string{
					"level": "critical",
				},
			},
		},
	}

	return result
}

// changeSMTP applies SMTP server configuration.
func changeSMTP(ctx context.Context, log log15.Logger, change ChangeContext, newConfig *subscribedSiteConfig) (result ChangeResult) {
	if change.AMConfig.Global == nil {
		change.AMConfig.Global = &amconfig.GlobalConfig{}
	}

	email := newConfig.Email
	change.AMConfig.Global.SMTPFrom = email.Address
	change.AMConfig.Global.SMTPHello = email.SMTP.Domain
	change.AMConfig.Global.SMTPSmarthost.Host = email.SMTP.Host
	change.AMConfig.Global.SMTPSmarthost.Port = strconv.Itoa(email.SMTP.Port)
	change.AMConfig.Global.SMTPAuthUsername = email.SMTP.Username
	switch email.SMTP.Authentication {
	case "PLAIN":
		change.AMConfig.Global.SMTPAuthPassword = amconfig.Secret(email.SMTP.Password)
	case "CRAM-MD5":
		change.AMConfig.Global.SMTPAuthSecret = amconfig.Secret(email.SMTP.Password)
	}

	return
}
