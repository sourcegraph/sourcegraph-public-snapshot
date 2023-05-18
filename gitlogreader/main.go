package main

import (
	"bufio"
	"io"
	_ "net/http/pprof"
	"strings"
	"sync"

	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
)

func main() {
	mapCommits()
}

func mapCommits() {
	logger := log.Scoped("gitlogreader", "")

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "git", "log")

	// dir := server.GitDir("/Users/dhanush/one-big-repo/the-great-pass/.git")
	dir := server.GitDir("/Users/dhanush/sourcegraph/sourcegraph/.git")
	// dir := server.GitDir("/Users/dhanush/github.com/probable-happiness/.git")

	logFormatWithoutRefs := "--format=format:%H %b"
	cmd.Args = append(cmd.Args, logFormatWithoutRefs, "-n", "5")
	dir.Set(cmd)

	pr, pw := io.Pipe()
	defer pw.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
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

	output, err := server.RunWith(ctx, wrexec.Wrap(ctx, logger, cmd), false, nil)
	if err != nil {
		fmt.Println(&server.GitCommandError{Err: err, Output: string(output)})
		os.Exit(1)
	}

	fmt.Println("waiting")
	wg.Wait()
}
