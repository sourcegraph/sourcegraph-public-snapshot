package main

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	opsgeniealert "github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

var pageCommand = &cli.Command{
	Name:      "page",
	UsageText: "sg page --opsgenie.token [TOKEN] --message \"something is broken\" [my-schedule-on-ops-genie]",
	ArgsUsage: "[opsgenie-schedule]",
	Usage:     "Page engineers at Sourcegraph - mostly used within scripts to automate paging alerts",
	Category:  CategoryCompany,
	Action:    pageExec,
	Flags: []cli.Flag{
		&cli.StringFlag{
			// TODO allow setting from secrets store
			Name:     "opsgenie.token",
			Usage:    "OpsGenie token",
			Required: true,
			EnvVars:  []string{"SG_OPSGENIE_TOKEN"},
		},
		&cli.StringFlag{
			Name:     "message",
			Aliases:  []string{"m"},
			Usage:    "Message for the paging alert",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "description",
			Aliases: []string{"d"},
			Usage:   "Description for the paging alert (optional)",
		},
		&cli.StringFlag{
			Name:    "priority",
			Aliases: []string{"p"},
			Usage:   "Alert priority, importance decreases from P1 (critical) to P5 (lowest), defaults to P5",
			Value:   "P5",
		},
		&cli.StringFlag{
			Name:  "url",
			Usage: "URL field for alert details (optional)",
		},
		&cli.StringSliceFlag{
			Name:  "escalation",
			Usage: "Escalation team(s) to alert (if provided, target schedules can be omitted)",
		},
	},
}

func pageExec(cmd *cli.Context) error {
	logger := log.Scoped("pager", "paging client for SG")

	priority, err := parseOpsGeniePriority(cmd.String("priority"))
	if err != nil {
		return errors.Wrap(err, "cannot parse ops priority")
	}

	var (
		responders  = cmd.Args().Slice()
		escalations = cmd.StringSlice("escalation")
	)

	if len(responders) == 0 && len(escalations) == 0 {
		// At least one responder must be given.
		return errors.New("Target responder schedules or esclation schedules are required")
	}

	logger.Info("paging schedules",
		log.String("priority", string(priority)),
		log.Strings("responders", responders),
		log.Strings("escalations", escalations))

	alertClient, err := opsgeniealert.NewClient(&client.Config{
		ApiKey: cmd.String("opsgenie.token"),
	})
	if err != nil {
		return errors.Wrap(err, "cannot create opsgenie client")
	}

	req := &opsgeniealert.CreateAlertRequest{
		Priority:    priority,
		Message:     cmd.String("message"),
		Description: cmd.String("description"),
	}
	if pageURL := cmd.String("url"); pageURL != "" {
		req.Details = map[string]string{
			"url": pageURL,
		}
	}
	for _, schedule := range responders {
		req.Responders = append(req.Responders, opsgeniealert.Responder{Type: opsgeniealert.ScheduleResponder, Name: schedule})
	}
	for _, schedule := range escalations {
		req.Responders = append(req.Responders, opsgeniealert.Responder{Type: opsgeniealert.EscalationResponder, Name: schedule})
	}

	createResult, err := alertClient.Create(cmd.Context, req)
	if err != nil {
		if createResult != nil {
			logger.Error("got error result from posting alert to opsgenie", log.Error(err), log.String("result", createResult.Result))
		}
		return errors.Wrap(err, "failed to post the alert on ops genie")
	}
	return nil
}

func parseOpsGeniePriority(p string) (opsgeniealert.Priority, error) {
	switch p {
	case "P1":
		return opsgeniealert.P1, nil
	case "P2":
		return opsgeniealert.P2, nil
	case "P3":
		return opsgeniealert.P3, nil
	case "P4":
		return opsgeniealert.P4, nil
	case "P5":
		return opsgeniealert.P5, nil
	default:
		return opsgeniealert.Priority(""), errors.Errorf("invalid priority %q", p)
	}
}
