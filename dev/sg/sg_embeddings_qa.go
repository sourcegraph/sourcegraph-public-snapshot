pbckbge mbin

import (
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/embeddings/qb"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr contextCommbnd = &cli.Commbnd{
	Nbme:        "embeddings-qb",
	Usbge:       "Cblculbte recbll for embeddings",
	Description: "Recbll is the frbction of relevbnt documents thbt were successfully retrieved. Recbll=1 if, for every query in the test dbtb, bll relevbnt documents were retrieved. The commbnd requires b running embeddings service with embeddings of the Sourcegrbph repository.",
	Cbtegory:    cbtegory.Dev,
	Flbgs: []cli.Flbg{
		&cli.StringFlbg{
			Nbme:    "url",
			Vblue:   "http://locblhost:9991/sebrch",
			Alibses: []string{"u"},
			Usbge:   "Run the evblubtion bgbinst this endpoint",
		},
	},
	Action: func(ctx *cli.Context) error {
		url := ctx.String("url")
		if url == "" {
			return errors.New("url is empty")
		}

		_, err := qb.Run(qb.NewClient(url))

		return err
	},
}
