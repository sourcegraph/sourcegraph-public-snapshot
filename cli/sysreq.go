package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/kr/text"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/sysreq"

	"context"
)

const skipSysReqsEnvVar = "SRC_SKIP_REQS"

var skipSysReqsEnv = env.Get(skipSysReqsEnvVar, "false", "skip system requirement checks")

// skippedSysReqs returns a list of sysreq names to skip (e.g.,
// "Docker").
func skippedSysReqs() []string {
	return strings.Fields(skipSysReqsEnv)
}

// checkSysReqs uses package sysreq to check for the presence of
// system requirements. If any are missing, it prints a message to
// w and returns a non-nil error.
func checkSysReqs(ctx context.Context, w io.Writer) error {
	wrap := func(s string) string {
		const indent = "\t\t"
		return strings.TrimPrefix(text.Indent(text.Wrap(s, 72), "\t\t"), indent)
	}

	var failed []string
	for _, st := range sysreq.Check(ctx, skippedSysReqs()) {
		if st.Failed() {
			failed = append(failed, st.Name)

			fmt.Fprint(w, redbg(" !!!!! "))
			fmt.Fprintf(w, bold(red(" %s is required\n")), st.Name)
			if st.Problem != "" {
				fmt.Fprint(w, bold(red("\tProblem: ")))
				fmt.Fprintln(w, red(wrap(st.Problem)))
			}
			if st.Err != nil {
				fmt.Fprint(w, bold(red("\tError: ")))
				fmt.Fprintln(w, red(wrap(st.Err.Error())))
			}
			if st.Fix != "" {
				fmt.Fprint(w, bold(green("\tPossible fix: ")))
				fmt.Fprintln(w, green(wrap(st.Fix)))
			}
			fmt.Fprintln(w, "\t"+cyan(wrap(fmt.Sprintf("Skip this check by setting the env var %s=%q (separate multiple entries with spaces). Note: Sourcegraph may not function properly without %s.", skipSysReqsEnvVar, st.Name, st.Name))))
		}
	}

	if failed != nil {
		return fmt.Errorf("system requirement checks failed: %v (see above for more information)", failed)
	}
	return nil
}
