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

	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/conc/pool"

	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
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
		panic(fmt.Sprintf("unknown input type %T", i))
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
	tr, ctx := trace.New(ctx, "comby", "Run")
	tr.TagFields(log.Object("args", args))
	defer func() {
		tr.LogFields(log.Int("num_matches", len(results)))
		tr.SetError(err)
		tr.Finish()
	}()

	cmd, stdin, stdout, err := SetupCmdWithPipes(ctx, args)
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

	p.Go(func() error {
		return StartAndWaitForCompletion(cmd)
	})

	return results, p.Wait()
}

func SetupCmdWithPipes(ctx context.Context, args Args) (cmd *exec.Cmd, stdin io.WriteCloser, stdout io.ReadCloser, err error) {
	if !Exists() {
		return nil, nil, nil, errors.New("comby is not installed")
	}

	rawArgs := rawArgs(args)

	cmd = exec.CommandContext(ctx, combyPath, rawArgs...)
	// Ensure forked child processes are killed
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdin, err = cmd.StdinPipe()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to connect to comby command stdin")
	}
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to connect to comby command stdout")
	}

	return cmd, stdin, stdout, nil
}

func StartAndWaitForCompletion(cmd *exec.Cmd) error {
	var b bytes.Buffer
	cmd.Stderr = &b

	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start comby command")
	}

	if err := cmd.Wait(); err != nil {
		if len(b.Bytes()) > 0 {
			return errors.Wrapf(err, "failed to wait for comby command: stderr: %s", b.String())
		}
		var stderr string
		var e *exec.ExitError
		if errors.As(err, &e) {
			stderr = string(e.Stderr)
		}
		return errors.Wrapf(err, "failed to wait for executing comby command: stderr: %s", stderr)
	}
	return nil
}

// Matches returns all matches in all files for which comby finds matches.
func Matches(ctx context.Context, args Args) ([]*FileMatch, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Comby.Matches") //nolint:staticcheck // OT is deprecated
	defer span.Finish()

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
func Replacements(ctx context.Context, args Args) ([]*FileReplacement, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Comby.Replacements") //nolint:staticcheck // OT is deprecated
	defer span.Finish()

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
func Outputs(ctx context.Context, args Args) (string, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Comby.Outputs") //nolint:staticcheck // OT is deprecated
	defer span.Finish()

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
