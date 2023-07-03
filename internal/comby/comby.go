//go:build !windows
// +build !windows

package comby

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/conc/pool"

	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const combyPath = "comby"

func Exists() bool {
	_, err := exec.LookPath(combyPath)
	return err == nil
}

func rawArgs(args Args) (rawArgs []string) {
	rawArgs = append(rawArgs, args.MatchTemplate, args.RewriteTemplate)

	if args.Rule != "" {
		rawArgs = append(rawArgs, "-rule", args.Rule)
	}

	if len(args.FilePatterns) > 0 {
		rawArgs = append(rawArgs, "-f", strings.Join(args.FilePatterns, ","))
	}
	rawArgs = append(rawArgs, "-json-lines", "-match-newline-at-toplevel")

	switch args.ResultKind {
	case MatchOnly:
		rawArgs = append(rawArgs, "-match-only")
	case Diff:
		rawArgs = append(rawArgs, "-json-only-diff")
	case NewlineSeparatedOutput:
		rawArgs = append(rawArgs, "-stdout", "-newline-separated")
	case Replacement:
		// Output contains replacement data in rewritten_source of JSON.
	}

	if args.NumWorkers == 0 {
		rawArgs = append(rawArgs, "-sequential")
	} else {
		rawArgs = append(rawArgs, "-jobs", strconv.Itoa(args.NumWorkers))
	}

	if args.Matcher != "" {
		rawArgs = append(rawArgs, "-matcher", args.Matcher)
	}

	switch i := args.Input.(type) {
	case ZipPath:
		rawArgs = append(rawArgs, "-zip", string(i))
	case DirPath:
		rawArgs = append(rawArgs, "-directory", string(i))
	case FileContent:
		rawArgs = append(rawArgs, "-stdin")
	case Tar:
		rawArgs = append(rawArgs, "-tar", "-chunk-matches", "0")
	default:
		log15.Error("unrecognized input type", "type", i)
		panic("unreachable")
	}

	return rawArgs
}

type unmarshaller func([]byte) (Result, error)

func ToCombyFileMatchWithChunks(b []byte) (Result, error) {
	var m FileMatchWithChunks
	err := json.Unmarshal(b, &m)
	return &m, errors.Wrap(err, "unmarshal JSON")
}

func ToFileMatch(b []byte) (Result, error) {
	var m FileMatch
	err := json.Unmarshal(b, &m)
	return &m, errors.Wrap(err, "unmarshal JSON")
}

func toFileReplacement(b []byte) (Result, error) {
	var r FileReplacement
	err := json.Unmarshal(b, &r)
	return &r, errors.Wrap(err, "unmarshal JSON")
}

func toOutput(b []byte) (Result, error) {
	return &Output{Value: b}, nil
}

func Run(ctx context.Context, args Args, unmarshal unmarshaller) (results []Result, err error) {
	cmd, stdin, stdout, stderr, err := SetupCmdWithPipes(ctx, args)
	if err != nil {
		return nil, err
	}

	p := pool.New().WithErrors()

	if bts, ok := args.Input.(FileContent); ok && len(bts) > 0 {
		p.Go(func() error {
			defer stdin.Close()
			_, err := stdin.Write(bts)
			return errors.Wrap(err, "write to stdin")
		})
	}

	p.Go(func() error {
		defer stdout.Close()

		scanner := bufio.NewScanner(stdout)
		// increase the scanner buffer size for potentially long lines
		scanner.Buffer(make([]byte, 100), 10*bufio.MaxScanTokenSize)
		for scanner.Scan() {
			b := scanner.Bytes()
			r, err := unmarshal(b)
			if err != nil {
				return err
			}
			results = append(results, r)
		}

		return errors.Wrap(scanner.Err(), "scan")
	})

	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "start comby")
	}

	// Wait for readers and writers to complete before calling Wait
	// because Wait closes the pipes.
	if err := p.Wait(); err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, InterpretCombyError(err, stderr)
	}

	if len(results) > 0 {
		log15.Info("comby invocation", "num_matches", strconv.Itoa(len(results)))
	}
	return results, nil
}

func InterpretCombyError(err error, stderr *bytes.Buffer) error {
	if len(stderr.Bytes()) > 0 {
		log15.Error("failed to execute comby command", "error", stderr.String())
		msg := fmt.Sprintf("failed to wait for executing comby command: comby error: %s", stderr.String())
		return errors.Wrap(err, msg)
	}
	var stderrString string
	var e *exec.ExitError
	if errors.As(err, &e) {
		stderrString = string(e.Stderr)
	}
	log15.Error("failed to wait for executing comby command", "error", stderrString)
	return errors.Wrap(err, "failed to wait for executing comby command")
}

func SetupCmdWithPipes(ctx context.Context, args Args) (cmd *exec.Cmd, stdin io.WriteCloser, stdout io.ReadCloser, stderr *bytes.Buffer, err error) {
	if !Exists() {
		log15.Error("comby is not installed (it could not be found on the PATH)")
		return nil, nil, nil, nil, errors.New("comby is not installed")
	}

	rawArgs := rawArgs(args)
	log15.Info("preparing to run comby", "args", args.String())

	cmd = exec.CommandContext(ctx, combyPath, rawArgs...)
	// Ensure forked child processes are killed
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdin, err = cmd.StdinPipe()
	if err != nil {
		log15.Error("could not connect to comby command stdin", "error", err.Error())
		return nil, nil, nil, nil, errors.Wrap(err, "failed to connect to comby command stdin")
	}
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		log15.Error("could not connect to comby command stdout", "error", err.Error())
		return nil, nil, nil, nil, errors.Wrap(err, "failed to connect to comby command stdout")
	}

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	return cmd, stdin, stdout, &stderrBuf, nil
}

// Matches returns all matches in all files for which comby finds matches.
func Matches(ctx context.Context, args Args) (_ []*FileMatch, err error) {
	tr, ctx := trace.DeprecatedNew(ctx, "comby", "matches")
	defer tr.FinishWithErr(&err)

	args.ResultKind = MatchOnly
	results, err := Run(ctx, args, ToFileMatch)
	if err != nil {
		return nil, err
	}
	var matches []*FileMatch
	for _, r := range results {
		matches = append(matches, r.(*FileMatch))
	}
	return matches, nil
}

// Replacements performs in-place replacement for match and rewrite template.
func Replacements(ctx context.Context, args Args) (_ []*FileReplacement, err error) {
	tr, ctx := trace.DeprecatedNew(ctx, "comby", "replacements")
	defer tr.FinishWithErr(&err)

	results, err := Run(ctx, args, toFileReplacement)
	if err != nil {
		return nil, err
	}
	var matches []*FileReplacement
	for _, r := range results {
		matches = append(matches, r.(*FileReplacement))
	}
	return matches, nil
}

// Outputs performs substitution of all variables captured in a match
// pattern in a rewrite template and outputs the result, newline-sparated.
func Outputs(ctx context.Context, args Args) (_ string, err error) {
	tr, ctx := trace.DeprecatedNew(ctx, "comby", "outputs")
	defer tr.FinishWithErr(&err)

	results, err := Run(ctx, args, toOutput)
	if err != nil {
		return "", err
	}
	var values []string
	for _, r := range results {
		values = append(values, string(r.(*Output).Value))
	}
	return strings.Join(values, "\n"), nil
}
