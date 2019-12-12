package comby

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"gopkg.in/inconshreveable/log15.v2"
)

const combyPath = "comby"

func exists() bool {
	_, err := exec.LookPath(combyPath)
	return err == nil
}

func rawArgs(args Args) (rawArgs []string) {
	rawArgs = append(rawArgs, args.MatchTemplate, args.RewriteTemplate)
	if len(args.FilePatterns) > 0 {
		rawArgs = append(rawArgs, "-f", strings.Join(args.FilePatterns, ","))
	}
	rawArgs = append(rawArgs, "-json-lines")

	if args.MatchOnly {
		rawArgs = append(rawArgs, "-match-only")
	} else {
		rawArgs = append(rawArgs, "-json-only-diff")
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
	default:
		log15.Error("unrecognized input type: %T", i)
		panic("unreachable")
	}

	return rawArgs
}

func PipeTo(ctx context.Context, args Args, w io.Writer) (err error) {
	if !exists() {
		log15.Error("comby is not installed (it could not be found on the PATH)")
		return errors.New("comby is not installed")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rawArgs := rawArgs(args)
	log15.Info("running comby", "args", strings.Join(rawArgs, " "))

	cmd := exec.CommandContext(ctx, combyPath, rawArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log15.Error("could not connect to comby command stdout", "error", err.Error())
		return errors.Wrap(err, "failed to connect to comby command stdout")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log15.Error("could not connect to comby command stderr", "error", err.Error())
		return errors.Wrap(err, "failed to connect to comby command stderr")
	}

	if err := cmd.Start(); err != nil {
		log15.Error("failed to start comby command", "error", err.Error())
		return errors.Wrap(err, "failed to start comby command")
	}

	// Read stderr in goroutine so we don't potentially block reading stdout
	stderrMsgC := make(chan []byte, 1)
	go func() {
		msg, _ := ioutil.ReadAll(stderr)
		stderrMsgC <- msg
		close(stderrMsgC)
	}()

	_, err = io.Copy(w, stdout)
	if err != nil {
		log15.Error("failed to copy comby output to writer", "error", err.Error())
		return errors.Wrap(err, "failed to copy comby output to writer")
	}

	stderrMsg := <-stderrMsgC

	if err := cmd.Wait(); err != nil {
		if len(stderrMsg) > 0 {
			log15.Error("failed to execute comby command", "error", string(stderrMsg))
			return errors.Errorf("comby error: %s", stderrMsg)
		}
		log15.Error("failed to wait for executing comby command", "error", string(err.(*exec.ExitError).Stderr))
		return errors.Wrap(err, "failed to wait for executing comby command")
	}

	return nil
}

// Matches returns all matches in all files for which comby finds matches.
func Matches(ctx context.Context, args Args) (matches []FileMatch, err error) {
	b := new(bytes.Buffer)
	w := bufio.NewWriter(b)

	args.MatchOnly = true

	err = PipeTo(ctx, args, w)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(b)
	// increase the scanner buffer size for potentially long lines
	scanner.Buffer(make([]byte, 100), 10*bufio.MaxScanTokenSize)
	for scanner.Scan() {
		b := scanner.Bytes()
		if err := scanner.Err(); err != nil {
			// warn on scanner errors and skip
			log15.Warn("comby error: skipping scanner error line: %v", err)
			continue
		}
		var m *FileMatch
		if err := json.Unmarshal(b, &m); err != nil {
			// warn on decode errors and skip
			log15.Warn("comby error: skipping unmarshaling error: %v", err)
			continue
		}
		matches = append(matches, *m)
	}

	if len(matches) > 0 {
		log15.Info("comby invocation", "num_matches", strconv.Itoa(len(matches)))
	}
	return matches, nil
}
