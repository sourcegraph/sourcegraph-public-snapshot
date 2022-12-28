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
	"sync"
	"syscall"

	"github.com/inconshreveable/log15"

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
		log15.Error("unrecognized input type", "type", i)
		panic("unreachable")
	}

	return rawArgs
}

type unmarshaller func([]byte) Result

func ToCombyFileMatchWithChunks(b []byte) Result {
	var m *FileMatchWithChunks
	if err := json.Unmarshal(b, &m); err != nil {
		log15.Warn("ToCombyFileMatchWithChunks() comby error: skipping unmarshaling error", "err", err.Error())
		return nil
	}
	return m
}

func ToFileMatch(b []byte) Result {
	var m *FileMatch
	if err := json.Unmarshal(b, &m); err != nil {
		log15.Warn("toFileMatch() comby error: skipping unmarshaling error", "err", err.Error())
		return nil
	}
	return m
}

func toFileReplacement(b []byte) Result {
	var r *FileReplacement
	if err := json.Unmarshal(b, &r); err != nil {
		log15.Warn("toFileReplacement() comby error: skipping unmarshaling error", "err", err.Error())
		return nil
	}
	return r
}

func toOutput(b []byte) Result {
	return &Output{Value: b}
}

func Run(ctx context.Context, args Args, unmarshal unmarshaller) (results []Result, err error) {
	cmd, stdin, stdout, err := SetupCmdWithPipes(ctx, args)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}

	if bts, ok := args.Input.(FileContent); ok && len(bts) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer stdin.Close()

			_, err := stdin.Write(bts)
			if err != nil {
				log15.Error("failed to write comby input to stdin", "error", err.Error())
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer stdout.Close()

		scanner := bufio.NewScanner(stdout)
		// increase the scanner buffer size for potentially long lines
		scanner.Buffer(make([]byte, 100), 10*bufio.MaxScanTokenSize)
		for scanner.Scan() {
			b := scanner.Bytes()
			if err := scanner.Err(); err != nil {
				// warn on scanner errors and skip
				log15.Warn("comby error: skipping scanner error line", "err", err.Error())
				continue
			}
			if r := unmarshal(b); r != nil {
				results = append(results, r)
			}
		}
	}()

	err = StartAndWaitForCompletion(cmd)
	if err != nil {
		return nil, err
	}

	wg.Wait()

	if len(results) > 0 {
		log15.Info("comby invocation", "num_matches", strconv.Itoa(len(results)))
	}
	return results, nil
}

func SetupCmdWithPipes(ctx context.Context, args Args) (cmd *exec.Cmd, stdin io.WriteCloser, stdout io.ReadCloser, err error) {
	if !Exists() {
		log15.Error("comby is not installed (it could not be found on the PATH)")
		return nil, nil, nil, errors.New("comby is not installed")
	}

	rawArgs := rawArgs(args)
	log15.Info("preparing to run comby", "args", args.String())

	cmd = exec.CommandContext(ctx, combyPath, rawArgs...)
	// Ensure forked child processes are killed
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdin, err = cmd.StdinPipe()
	if err != nil {
		log15.Error("could not connect to comby command stdin", "error", err.Error())
		return nil, nil, nil, errors.Wrap(err, "failed to connect to comby command stdin")
	}
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		log15.Error("could not connect to comby command stdout", "error", err.Error())
		return nil, nil, nil, errors.Wrap(err, "failed to connect to comby command stdout")
	}

	return cmd, stdin, stdout, nil
}

func StartAndWaitForCompletion(cmd *exec.Cmd) error {
	log15.Info("starting comby")

	var b bytes.Buffer
	cmd.Stderr = &b

	if err := cmd.Start(); err != nil {
		log15.Error("failed to start comby command", "error", err.Error())
		return errors.Wrap(err, "failed to start comby command")
	}

	if err := cmd.Wait(); err != nil {
		if len(b.Bytes()) > 0 {
			log15.Error("failed to execute comby command", "error", b.String())
			msg := fmt.Sprintf("failed to wait for executing comby command: comby error: %s", b.String())
			return errors.Wrap(err, msg)
		}
		var stderr string
		var e *exec.ExitError
		if errors.As(err, &e) {
			stderr = string(e.Stderr)
		}
		log15.Error("failed to wait for executing comby command", "error", stderr)
		return errors.Wrap(err, "failed to wait for executing comby command")
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
