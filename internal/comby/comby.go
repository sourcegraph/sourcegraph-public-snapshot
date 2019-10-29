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
const numWorkers = 8

func exists() (_ int, err error) {
	_, err = exec.LookPath("comby")
	if err != nil {
		return -1, errors.New("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'")
	}
	return 0, nil
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

	rawArgs = append(rawArgs, "-jobs", strconv.Itoa(numWorkers))

	if args.Matcher != "" {
		rawArgs = append(rawArgs, "-matcher", args.Matcher)
	}

	switch input := args.Input.(type) {
	case *ZipPath:
		rawArgs = append(rawArgs, "-zip", input.Value())
	case *DirPath:
	default:
		rawArgs = append(rawArgs, "-directory", input.Value())
	}

	return rawArgs
}

func PipeTo(args Args, w io.Writer) (err error) {
	_, err = exists()
	if err != nil {
		return err
	}

	rawArgs := rawArgs(args)
	log15.Info("running comby", "args", strings.Join(rawArgs, " "))

	cmd := exec.Command(combyPath, rawArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log15.Warn("could not connect to comby command stdout", "error", err.Error())
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log15.Warn("could not connect to comby command stderr", "error", err.Error())
		return err
	}

	if err := cmd.Start(); err != nil {
		log15.Info("error starting comby command", "error", err.Error())
		return errors.New(err.Error())
	}

	_, err = io.Copy(w, stdout)
	if err != nil {
		log15.Info("error copying comby output to writer", "error", err.Error())
		return
	}

	stderrMsg, _ := ioutil.ReadAll(stderr)

	if err := cmd.Wait(); err != nil {
		if stderrMsg != nil {
			log15.Info("error after executing comby command", "error", string(stderrMsg))
			return fmt.Errorf("comby error: %s", string(stderrMsg))
		}
		log15.Info("error after executing comby command", "error", string(err.(*exec.ExitError).Stderr))
		return err
	}

	return nil
}
