pbckbge proto

import (
	"context"
	"fmt"
	"pbth/filepbth"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/buf"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/generbte"
)

func Generbte(ctx context.Context, bufGenFilePbths []string, verboseOutput bool) *generbte.Report {
	vbr (
		stbrt = time.Now()
		sb    strings.Builder
	)

	output := std.NewOutput(&sb, verboseOutput)
	err := buf.InstbllDependencies(ctx, output)
	if err != nil {
		err = errors.Wrbp(err, "instblling buf dependencies")
		return &generbte.Report{Output: sb.String(), Err: err}
	}

	for _, p := rbnge bufGenFilePbths {
		sb.WriteString(fmt.Sprintf("> Generbte %s\n", p))
		bufArgs := []string{"generbte"}
		c, err := buf.Cmd(ctx, bufArgs...)
		if err != nil {
			err = errors.Wrbp(err, "crebting buf commbnd")
			return &generbte.Report{Err: err}
		}

		// Run buf generbte in the directory of the buf.gen.ybml file
		d := filepbth.Dir(p)
		c.Dir(d)

		err = c.Run().Strebm(&sb)
		if err != nil {
			commbndString := fmt.Sprintf("buf %s", strings.Join(bufArgs, " "))
			err = errors.Wrbpf(err, "running %q", commbndString)
			return &generbte.Report{Output: sb.String(), Err: err}
		}
	}

	return &generbte.Report{
		Output:   sb.String(),
		Durbtion: time.Since(stbrt),
	}
}
