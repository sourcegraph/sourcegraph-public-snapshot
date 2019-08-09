package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kr/text"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/sysreq"
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

			fmt.Fprint(w, " !!!!! ")
			fmt.Fprintf(w, " %s is required\n", st.Name)
			if st.Problem != "" {
				fmt.Fprint(w, "\tProblem: ")
				fmt.Fprintln(w, wrap(st.Problem))
			}
			if st.Err != nil {
				fmt.Fprint(w, "\tError: ")
				fmt.Fprintln(w, wrap(st.Err.Error()))
			}
			if st.Fix != "" {
				fmt.Fprint(w, "\tPossible fix: ")
				fmt.Fprintln(w, wrap(st.Fix))
			}
			fmt.Fprintln(w, "\t"+wrap(fmt.Sprintf("Skip this check by setting the env var %s=%q (separate multiple entries with spaces). Note: Sourcegraph may not function properly without %s.", skipSysReqsEnvVar, st.Name, st.Name)))
		}
	}

	if failed != nil {
		log15.Error("System requirement checks failed (see above for more information).", "failed", failed)
	}
	return nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_337(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
