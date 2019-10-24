package comby

import (
	"bufio"
	"encoding/json"
	"errors"
	"os/exec"
	"strconv"

	"gopkg.in/inconshreveable/log15.v2"
)

const combyPath = "comby"
const numWorkers = 8

type Range struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

// {"uri":"/private/tmp/rrrr/doc.go","matches":[{"range":{"start":{"offset":215,"line":1,"column":216},"end":{"offset":222,"line":1,"column":223}},"environment":[],"matched":"package"}]}
type Match struct {
	URI     string  `json:"uri"`
	Matches []Range `json:"matches"`
	Matched string  `json:"matched"`
}

type Diff struct {
	URI  string `json:"uri"`
	Diff string `json:"diff"`
}

// A result is either a match result or a diff result. These members are
// mutually exclusive, but are bundled together so that we can share the
// unmarshalling code.
type Result struct {
	Matches *[]Match
	Diffs   *[]Diff
}

func exists() (_ int, err error) {
	_, err = exec.LookPath("comby")
	if err != nil {
		return -1, errors.New("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'.")
	}
	return 0, nil
}

func run(args []string, matchOnly bool) (result *Result, err error) {
	var matches []Match
	var diffs []Diff
	cmd := exec.Command(combyPath, args...)

	// TODO test stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log15.Info("Could not connect to command stdout: " + err.Error())
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		log15.Info("Error starting command: " + err.Error())
		return nil, errors.New(err.Error())
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		b := scanner.Bytes()
		if err := scanner.Err(); err != nil {
			// skip scanner errors
			continue
		}
		if matchOnly {
			var r *Match
			if err := json.Unmarshal(b, &r); err != nil {
				// skip decode errors
				continue
			}
			matches = append(matches, *r)
		} else {
			var r *Diff
			if err := json.Unmarshal(b, &r); err != nil {
				// skip decode errors
				continue
			}
			diffs = append(diffs, *r)
		}
	}

	if err := cmd.Wait(); err != nil {
		log15.Info("Error after executing command: " + string(err.(*exec.ExitError).Stderr))
		return nil, err
	}

	return &Result{Diffs: &diffs, Matches: &matches}, nil
}

func defaultCommand(matchTemplate string, rewriteTemplate string, matchOnly bool, anonArgs []string, otherArgs ...string) (result *Result, err error) {
	_, err = exists()
	if err != nil {
		return nil, err
	}

	args := []string{
		matchTemplate,
		rewriteTemplate,
		"-json-lines",
		"-jobs", strconv.Itoa(numWorkers),
	}

	if matchOnly {
		args = append(args, "-match-only")
	} else {
		args = append(args, "-json-only-diff")
	}

	args = append(args, otherArgs...)
	args = append(args, anonArgs...)
	return run(args, matchOnly)
}

func MatchesInDir(matchTemplate string, root string, filePatterns []string, jobs int) (results *[]Match, err error) {
	r, err := defaultCommand(matchTemplate, "", true, filePatterns, "-directory", root)
	if err != nil {
		return nil, err
	}
	return r.Matches, nil
}

func DiffsInZip(matchTemplate string, rewriteTemplate string, zipPath string, filePatterns []string, jobs int) (results *[]Diff, err error) {
	r, err := defaultCommand(matchTemplate, rewriteTemplate, true, filePatterns, "-zip", zipPath)
	if err != nil {
		return nil, err
	}
	return r.Diffs, nil
}

func DiffsInDir(matchTemplate string, rewriteTemplate string, root string, filePatterns []string, jobs int) (results *[]Diff, err error) {
	r, err := defaultCommand(matchTemplate, rewriteTemplate, true, filePatterns, "-directory", root)
	if err != nil {
		return nil, err
	}
	return r.Diffs, nil
}
