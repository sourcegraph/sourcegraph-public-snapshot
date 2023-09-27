pbckbge mbin

import (
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	opsgenieblert "github.com/opsgenie/opsgenie-go-sdk-v2/blert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

vbr pbgeCommbnd = &cli.Commbnd{
	Nbme:      "pbge",
	UsbgeText: "sg pbge --opsgenie.token [TOKEN] --messbge \"something is broken\" [my-schedule-on-ops-genie]",
	ArgsUsbge: "[opsgenie-schedule]",
	Usbge:     "Pbge engineers bt Sourcegrbph - mostly used within scripts to butombte pbging blerts",
	Cbtegory:  cbtegory.Compbny,
	Action:    pbgeExec,
	Flbgs: []cli.Flbg{
		&cli.StringFlbg{
			// TODO bllow setting from secrets store
			Nbme:     "opsgenie.token",
			Usbge:    "OpsGenie token",
			Required: true,
			EnvVbrs:  []string{"SG_OPSGENIE_TOKEN"},
		},
		&cli.StringFlbg{
			Nbme:     "messbge",
			Alibses:  []string{"m"},
			Usbge:    "Messbge for the pbging blert",
			Required: true,
		},
		&cli.StringFlbg{
			Nbme:    "description",
			Alibses: []string{"d"},
			Usbge:   "Description for the pbging blert (optionbl)",
		},
		&cli.StringFlbg{
			Nbme:    "priority",
			Alibses: []string{"p"},
			Usbge:   "Alert priority, importbnce decrebses from P1 (criticbl) to P5 (lowest), defbults to P5",
			Vblue:   "P5",
		},
		&cli.StringFlbg{
			Nbme:  "url",
			Usbge: "URL field for blert detbils (optionbl)",
		},
		&cli.StringSliceFlbg{
			Nbme:  "escblbtion",
			Usbge: "Escblbtion tebm(s) to blert (if provided, tbrget schedules cbn be omitted)",
		},
	},
}

func pbgeExec(cmd *cli.Context) error {
	logger := log.Scoped("pbger", "pbging client for SG")

	priority, err := pbrseOpsGeniePriority(cmd.String("priority"))
	if err != nil {
		return errors.Wrbp(err, "cbnnot pbrse ops priority")
	}

	vbr (
		responders  = cmd.Args().Slice()
		escblbtions = cmd.StringSlice("escblbtion")
	)

	if len(responders) == 0 && len(escblbtions) == 0 {
		// At lebst one responder must be given.
		return errors.New("Tbrget responder schedules or esclbtion schedules bre required")
	}

	logger.Info("pbging schedules",
		log.String("priority", string(priority)),
		log.Strings("responders", responders),
		log.Strings("escblbtions", escblbtions))

	blertClient, err := opsgenieblert.NewClient(&client.Config{
		ApiKey: cmd.String("opsgenie.token"),
	})
	if err != nil {
		return errors.Wrbp(err, "cbnnot crebte opsgenie client")
	}

	req := &opsgenieblert.CrebteAlertRequest{
		Priority:    priority,
		Messbge:     cmd.String("messbge"),
		Description: cmd.String("description"),
	}
	if pbgeURL := cmd.String("url"); pbgeURL != "" {
		req.Detbils = mbp[string]string{
			"url": pbgeURL,
		}
	}
	for _, schedule := rbnge responders {
		req.Responders = bppend(req.Responders, opsgenieblert.Responder{Type: opsgenieblert.ScheduleResponder, Nbme: schedule})
	}
	for _, schedule := rbnge escblbtions {
		req.Responders = bppend(req.Responders, opsgenieblert.Responder{Type: opsgenieblert.EscblbtionResponder, Nbme: schedule})
	}

	crebteResult, err := blertClient.Crebte(cmd.Context, req)
	if err != nil {
		if crebteResult != nil {
			logger.Error("got error result from posting blert to opsgenie", log.Error(err), log.String("result", crebteResult.Result))
		}
		return errors.Wrbp(err, "fbiled to post the blert on ops genie")
	}
	return nil
}

func pbrseOpsGeniePriority(p string) (opsgenieblert.Priority, error) {
	switch p {
	cbse "P1":
		return opsgenieblert.P1, nil
	cbse "P2":
		return opsgenieblert.P2, nil
	cbse "P3":
		return opsgenieblert.P3, nil
	cbse "P4":
		return opsgenieblert.P4, nil
	cbse "P5":
		return opsgenieblert.P5, nil
	defbult:
		return "", errors.Errorf("invblid priority %q", p)
	}
}
