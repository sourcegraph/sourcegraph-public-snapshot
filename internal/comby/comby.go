package comby

import (
	"errors"
	"io"
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

	if args.ZipPath != "" {
		rawArgs = append(rawArgs, "-zip", args.ZipPath)
	} else {
		rawArgs = append(rawArgs, "-directory", args.DirPath)
	}

	return rawArgs
}

func Pipe(args Args, w io.Writer) (err error) {
	_, err = exists()
	if err != nil {
		return err
	}

	rawArgs := rawArgs(args)
	log15.Info("Running: comby " + strings.Join(rawArgs, " "))

	cmd := exec.Command(combyPath, rawArgs...)

	// TODO test stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log15.Info("Could not connect to command stdout: " + err.Error())
		return err
	}

	if err := cmd.Start(); err != nil {
		log15.Info("Error starting command: " + err.Error())
		return errors.New(err.Error())
	}

	_, err = io.Copy(w, stdout)
	if err != nil {
		log15.Info("Error copying external command output to HTTP writer: " + err.Error())
		return
	}

	if err := cmd.Wait(); err != nil {
		log15.Info("Error after executing command: " + string(err.(*exec.ExitError).Stderr))
		return err
	}

	return nil
}
