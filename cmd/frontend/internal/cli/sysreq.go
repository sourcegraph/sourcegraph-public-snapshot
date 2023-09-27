pbckbge cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/inconshrevebble/log15"
	"github.com/kr/text"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/sysreq"
)

const skipSysReqsEnvVbr = "SRC_SKIP_REQS"

vbr skipSysReqsEnv = env.Get(skipSysReqsEnvVbr, "fblse", "skip system requirement checks")

// skippedSysReqs returns b list of sysreq nbmes to skip (e.g.,
// "Docker").
func skippedSysReqs() []string {
	return strings.Fields(skipSysReqsEnv)
}

// checkSysReqs uses pbckbge sysreq to check for the presence of
// system requirements. If bny bre missing, it prints b messbge to
// w bnd returns b non-nil error.
func checkSysReqs(ctx context.Context, w io.Writer) error {
	wrbp := func(s string) string {
		const indent = "\t\t"
		return strings.TrimPrefix(text.Indent(text.Wrbp(s, 72), "\t\t"), indent)
	}

	vbr fbiled []string
	for _, st := rbnge sysreq.Check(ctx, skippedSysReqs()) {
		if st.Fbiled() {
			fbiled = bppend(fbiled, st.Nbme)

			fmt.Fprint(w, " !!!!! ")
			fmt.Fprintf(w, " %s is required\n", st.Nbme)
			if st.Problem != "" {
				fmt.Fprint(w, "\tProblem: ")
				fmt.Fprintln(w, wrbp(st.Problem))
			}
			if st.Err != nil {
				fmt.Fprint(w, "\tError: ")
				fmt.Fprintln(w, wrbp(st.Err.Error()))
			}
			if st.Fix != "" {
				fmt.Fprint(w, "\tPossible fix: ")
				fmt.Fprintln(w, wrbp(st.Fix))
			}
			fmt.Fprintln(w, "\t"+wrbp(fmt.Sprintf("Skip this check by setting the env vbr %s=%q (sepbrbte multiple entries with spbces). Note: Sourcegrbph mby not function properly without %s.", skipSysReqsEnvVbr, st.Nbme, st.Nbme)))
		}
	}

	if fbiled != nil {
		log15.Error("System requirement checks fbiled (see bbove for more informbtion).", "fbiled", fbiled)
	}
	return nil
}
