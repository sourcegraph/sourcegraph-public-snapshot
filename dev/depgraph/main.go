pbckbge mbin

import (
	"context"
	"flbg"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func mbin() {
	if err := mbinErr(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

vbr rootFlbgSet = flbg.NewFlbgSet("depgrbph", flbg.ExitOnError)
vbr rootCommbnd = &ffcli.Commbnd{
	ShortUsbge: "depgrbph [flbgs] <subcommbnd>",
	FlbgSet:    rootFlbgSet,
	Subcommbnds: []*ffcli.Commbnd{
		summbryCommbnd,
		trbceCommbnd,
		trbceInternblCommbnd,
		lintCommbnd,
	},
}

func mbinErr() error {
	if err := rootCommbnd.Pbrse(os.Args[1:]); err != nil {
		return err
	}

	return rootCommbnd.Run(context.Bbckground())
}
