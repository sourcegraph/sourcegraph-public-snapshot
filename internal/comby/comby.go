package comby

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"
)

const combyPath = "comby"

func exists() bool {
	_, err := exec.LookPath(combyPath)
	return err == nil
}

func rawArgs(args Args) (rawArgs []string) {
	rawArgs = append(rawArgs, args.MatchTemplate, args.RewriteTemplate)
	rawArgs = append(rawArgs, args.FilePatterns...)
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

func PipeTo(args Args, w io.Writer) (err error) {
	if !exists() {
		log15.Error("comby is not installed (it could not be found on the PATH)")
		return errors.New("comby is not installed")
	}

	rawArgs := rawArgs(args)
	log15.Info("running comby", "args", strings.Join(rawArgs, " "))

	cmd := exec.Command(combyPath, rawArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log15.Error("could not connect to comby command stdout", "error", err.Error())
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log15.Error("could not connect to comby command stderr", "error", err.Error())
		return err
	}

	if err := cmd.Start(); err != nil {
		log15.Error("failed to start comby command", "error", err.Error())
		return errors.New(err.Error())
	}

	_, err = io.Copy(w, stdout)
	if err != nil {
		log15.Error("failed to copy comby output to writer", "error", err.Error())
		return err
	}

	stderrMsg, _ := ioutil.ReadAll(stderr)

	if err := cmd.Wait(); err != nil {
		if stderrMsg != nil {
			log15.Error("failed to execute comby command", "error", string(stderrMsg))
			return fmt.Errorf("comby error: %s", string(stderrMsg))
		}
		log15.Error("failed to wait for executing comby command", "error", string(err.(*exec.ExitError).Stderr))
		return err
	}

	return nil
}
