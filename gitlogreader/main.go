package main

import (
	"bufio"
	"io"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"

	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/sync/errgroup"
)

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	logger := log.Scoped("gitlogreader", "")

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "git", "log")

	// dir := server.GitDir("/Users/dhanush/one-big-repo/the-great-pass/.git")
	// dir := server.GitDir("/Users/dhanush/sourcegraph/sourcegraph/.git")
	// dir := server.GitDir("/Users/dhanush/github.com/probable-happiness/.git")

	dir := server.GitDir("/Users/dhanush/.sourcegraph/repos_1/perforce.sgdev.org/rhia-depot-test/.git")

	logFormatWithoutRefs := "--format=format:%H %b"
	cmd.Args = append(cmd.Args, logFormatWithoutRefs)
	dir.Set(cmd)

	// mapCommitsFlipped(ctx, logger, cmd)

	// mapCommits(ctx, logger, cmd)

	mapCommitsErrGroup(ctx, logger, cmd)
}

func mapCommitsFlipped(ctx context.Context, logger log.Logger, cmd *exec.Cmd) {

	pr, pw := io.Pipe()
	defer pw.Close()

	go func() {
		output, err := server.RunWith(ctx, wrexec.Wrap(ctx, logger, cmd), false, pw)

		if err != nil {
			fmt.Println(&server.GitCommandError{Err: err, Output: string(output)})
			os.Exit(1)
		}
	}()

	scan := bufio.NewScanner(pr)
	scan.Split(bufio.ScanLines)
	fmt.Println("starting read")
	for scan.Scan() {
		s := scan.Text()
		parts := strings.SplitN(s, " ", 2)
		fmt.Printf("%v\n", parts)
		fmt.Printf("%d", len(parts))
	}

	if err := scan.Err(); err != nil {
		fmt.Println("error in scan")
	}
}

func mapCommits(ctx context.Context, logger log.Logger, cmd *exec.Cmd) {
	pr, pw := io.Pipe()
	defer pw.Close()

	go func() {
		scan := bufio.NewScanner(pr)
		scan.Split(bufio.ScanLines)
		fmt.Println("starting read")
		for scan.Scan() {
			s := scan.Text()
			parts := strings.Split(s, " ")
			fmt.Printf("%v\n", parts)
		}

		if err := scan.Err(); err != nil {
			fmt.Println("error in scan")
		}
	}()

	output, err := server.RunWith(ctx, wrexec.Wrap(ctx, logger, cmd), false, pw)
	if err != nil {
		fmt.Println(&server.GitCommandError{Err: err, Output: string(output)})
		os.Exit(1)
	}
}

func mapCommitsErrGroup(ctx context.Context, logger log.Logger, cmd *exec.Cmd) {
	progressReader, progressWriter := io.Pipe()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		defer progressWriter.Close()

		output, err := server.RunWith(ctx, wrexec.Wrap(ctx, logger, cmd), false, progressWriter)
		if err != nil {
			return &server.GitCommandError{Err: err, Output: string(output)}
		}
		return nil
	})

	logLineResults := make(chan string)

	g.Go(func() error {
		defer close(logLineResults)
		return readGitLogOutput(ctx, logger, progressReader, logLineResults)
	})

	go func() {
		g.Wait()
	}()

	commitMaps := []*types.PerforceChangelist{}
	for line := range logLineResults {
		c, err := parseGitLogLine(line)
		if err != nil {
			fmt.Printf("error in parsing git log line:\n%q\n", err.Error())
			return
		}
		commitMaps = append(commitMaps, c)
	}

	for _, c := range commitMaps {
		fmt.Println(*c)
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("error in g.Wait():\n%q\n", err.Error())
	}
}

func readGitLogOutput(ctx context.Context, logger log.Logger, reader io.Reader, logLineResults chan<- string) error {
	scan := bufio.NewScanner(reader)
	scan.Split(bufio.ScanLines)
	for scan.Scan() {
		line := scan.Text()

		select {
		case logLineResults <- line:
			return errors.New("early exit")
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	err := scan.Err()
	return errors.Wrap(err, "scanning git-log output failed")
}

func parseGitLogLine(line string) (*types.PerforceChangelist, error) {
	// Expected format: "<commitSHA> <commitBody>"
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return nil, errors.Newf("failed to split line %q from git log output into commitSHA and commit body, parts after splitting: %d", line, len(parts))
	}

	parsedCID, err := perforce.GetP4ChangelistID(parts[1])
	if err != nil {
		return nil, err
	}

	cid, err := strconv.ParseInt(parsedCID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse changelist ID to int64")
	}

	return &types.PerforceChangelist{
		CommitSHA:    api.CommitID(parts[0]),
		ChangelistID: cid,
	}, nil
}
