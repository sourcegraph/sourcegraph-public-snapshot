package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/inconshreveable/log15"
	amclient "github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
	"github.com/prometheus/alertmanager/api/v2/models"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

type ChangeContext struct {
	AMConfig *amconfig.Config // refer to https://prometheus.io/docs/alerting/latest/configuration/
	AMClient *amclient.Alertmanager
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

	// include `alertname` for now to accomodate non-generator alerts - in the long run, we want to remove grouping on `alertname`
	// because all alerts should have some predefined labels
	// https://github.com/sourcegraph/sourcegraph/issues/5370
	groupBy := []string{"alertname", "level", "service_name", "name"}
	// make sure alerts are routed appropriately
	change.AMConfig.Route = &amconfig.Route{
		Receiver:   alertmanagerNoopReceiver,
		GroupByStr: groupBy,
		Routes: []*amconfig.Route{
			{
				Receiver:   alertmanagerWarningReceiver,
				GroupByStr: groupBy,
				Match: map[string]string{
					"level": "warning",
				},
			},
			{
				Receiver:   alertmanagerCriticalReceiver,
				GroupByStr: groupBy,
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

	// assign zero-values to AMConfig SMTP fields if email.SMTP is nil
	if email.SMTP == nil {
		email.SMTP = &schema.SMTPServerConfig{}
	}
	change.AMConfig.Global.SMTPHello = email.SMTP.Domain
	change.AMConfig.Global.SMTPSmarthost = amconfig.HostPort{
		Host: email.SMTP.Host,
		Port: strconv.Itoa(email.SMTP.Port),
	}
	change.AMConfig.Global.SMTPAuthUsername = email.SMTP.Username
	switch email.SMTP.Authentication {
	case "PLAIN":
		change.AMConfig.Global.SMTPAuthPassword = amconfig.Secret(email.SMTP.Password)
	case "CRAM-MD5":
		change.AMConfig.Global.SMTPAuthSecret = amconfig.Secret(email.SMTP.Password)
	}
	change.AMConfig.Global.SMTPRequireTLS = !email.SMTP.DisableTLS

	return
}

func changeSilences(ctx context.Context, log log15.Logger, change ChangeContext, newConfig *subscribedSiteConfig) (result ChangeResult) {
	// convenience function for creating a prefixed problem - this reflects the relevant site configuration fields
	newProblem := func(err error) {
		result.Problems = append(result.Problems, conf.NewSiteProblem(fmt.Sprintf("`observability.alerts`: %v", err)))
	}

	var (
		createdBy      = "src-prom-wrapper"
		comment        = "Applied via `observability.alerts` in site configuration"
		startTime      = strfmt.DateTime(time.Now())
		endTime        = strfmt.DateTime(time.Unix(1<<63-62135596801, 999999999)) // https://stackoverflow.com/questions/25065055/what-is-the-maximum-time-time-in-go/32620397
		activeSilences = map[schema.ObservabilitySilenceAlerts]string{}           // map silence to alertmanager silence ID
	)

	for _, s := range newConfig.SilencedAlerts {
		activeSilences[*s] = ""
	}

	// delete existing silences that should no longer be silenced
	existingSilences, err := change.AMClient.Silence.GetSilences(&silence.GetSilencesParams{})
	if err != nil {
		newProblem(err)
		return
	}
	for _, s := range existingSilences.Payload {
		if *s.CreatedBy == createdBy && s.Comment != nil {
			// if this silence should not exist, delete
			silencedAlert := newSilenceFromMatchers(s.Matchers)
			if _, shouldBeActive := activeSilences[silencedAlert]; shouldBeActive {
				activeSilences[silencedAlert] = *s.ID
			} else {
				uid := strfmt.UUID(*s.ID)
				if _, err := change.AMClient.Silence.DeleteSilence(&silence.DeleteSilenceParams{
					SilenceID: uid,
				}); err != nil {
					newProblem(err)
					return
				}
			}
		}
	}

	// create or update silences
	for alert, existingSilence := range activeSilences {
		s := models.Silence{
			CreatedBy: &createdBy,
			Comment:   &comment,
			StartsAt:  &startTime,
			EndsAt:    &endTime,
			Matchers:  newMatchersFromSilence(alert),
		}
		var err error
		if existingSilence != "" {
			_, err = change.AMClient.Silence.PostSilences(&silence.PostSilencesParams{
				Context: ctx,
				Silence: &models.PostableSilence{
					ID:      existingSilence,
					Silence: s,
				},
			})
		} else {
			_, err = change.AMClient.Silence.PostSilences(&silence.PostSilencesParams{
				Context: ctx,
				Silence: &models.PostableSilence{
					Silence: s,
				},
			})
		}
		if err != nil {
			newProblem(err)
			return
		}
	}

	return result
}
