package main

import (
	"flag"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

var (
	pageToken               string
	pageMessage             string
	pageDesc                string
	pagePriority            string
	pageURL                 string
	pageResponderEscalation string
	pageResponderSchedule   string
)

var pageCommand = &cli.Command{
	Name:      "page",
	UsageText: "sg page --opsgenie.token [TOKEN] --message \"something is broken\" --responder.schedule [my-schedule-on-ops-genie]",
	Usage:     "Utility to page engineers, mostly used within scripts to automate paging alerts",
	Category:  CategoryUtil,
	Action:    pageExec,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "opsgenie.token",
			Usage:       "OpsGenie token",
			Required:    true,
			Destination: &pageToken,
		},
		&cli.StringFlag{
			Name:        "message",
			Usage:       "Message for the paging alert",
			Required:    true,
			Destination: &pageMessage,
		},
		&cli.StringFlag{
			Name:        "description",
			Usage:       "Description for the paging alert",
			Destination: &pageDesc,
		},
		&cli.StringFlag{
			Name:        "priority",
			Usage:       "Alert priority, importance decreases from P1 (critical) to P5 (lowest), defaults to P5",
			Value:       "P5",
			Destination: &pagePriority,
		},
		&cli.StringFlag{
			Name:        "url",
			Usage:       "URL field for alert details (optional)",
			Destination: &pageURL,
		},
		&cli.StringFlag{
			Name:        "responder.schedule",
			Usage:       "Schedule team to alert (at least one responder must be given)",
			Destination: &pageResponderSchedule,
		},
		&cli.StringFlag{
			Name:        "responder.escalation",
			Usage:       "Escalation team to alert (at least one responder must be given)",
			Destination: &pageResponderEscalation,
		},
	},
}

func pageExec(ctx *cli.Context) error {
	logger := log.Scoped("pager", "paging client for SG")
	var err error
	var prio alert.Priority
	if prio, err = parseOpsGeniePriority(pagePriority); err != nil {
		logger.Fatal("cannot parse ops priority", log.Error(err))
	}
	if pageResponderSchedule == "" && pageResponderEscalation == "" {
		// At least one responder must be given.
		flag.Usage()
		return flag.ErrHelp
	}

	alertClient, err := alert.NewClient(&client.Config{
		ApiKey: pageToken,
	})
	if err != nil {
		logger.Fatal("cannot create client", log.Error(err))
	}

	req := &alert.CreateAlertRequest{
		Message:  pageMessage,
		Priority: prio,
	}

	req.Description = ctx.String("description")

	if pageURL != "" {
		req.Details = map[string]string{
			"url": pageURL,
		}
	}
	if pageResponderSchedule != "" {
		req.Responders = append(req.Responders, alert.Responder{Type: alert.ScheduleResponder, Name: pageResponderSchedule})
	}
	if pageResponderEscalation != "" {
		req.Responders = append(req.Responders, alert.Responder{Type: alert.EscalationResponder, Name: pageResponderEscalation})
	}

	createResult, err := alertClient.Create(ctx.Context, req)

	if err != nil {
		if createResult != nil {
			logger.Fatal("failed to post the alert on ops genie", log.Error(err), log.String("result", createResult.Result))
		} else {
			logger.Fatal("failed to post the alert on ops genie", log.Error(err))
		}
	}
	return nil
}

func parseOpsGeniePriority(p string) (alert.Priority, error) {
	switch p {
	case "P1":
		return alert.P1, nil
	case "P2":
		return alert.P2, nil
	case "P3":
		return alert.P3, nil
	case "P4":
		return alert.P4, nil
	case "P5":
		return alert.P5, nil
	default:
		return alert.Priority(""), errors.Errorf("invalid priority %q", p)
	}
}
