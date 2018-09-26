package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const usage = `differ [utility [argument ...]]

Execute utility with the given arguments. Then exit with an error if git reports
there are untracked changes.
`

func init() {
	flag.Usage = func() {
		os.Stderr.WriteString(usage)
	}
}

func getGitDiff(ctx context.Context) string {
	diffBuf := new(bytes.Buffer)
	diffCmd := exec.CommandContext(ctx, "git", "diff", "--no-color")
	diffCmd.Stdout = diffBuf
	diffCmd.Stderr = diffBuf
	if diffErr := diffCmd.Run(); diffErr != nil {
		return ""
	}
	if diffBuf.Len() == 0 {
		return ""
	}
	bs := bufio.NewScanner(diffBuf)
	diffOutput := strings.Builder{}
	for i := 0; i < 20 && bs.Scan(); i++ {
		diffOutput.Write(bs.Bytes())
		diffOutput.WriteByte('\n')
	}
	return "\nFirst few lines of the git diff:\n" + diffOutput.String()
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if len(os.Args) <= 1 {
		flag.Usage()
		os.Exit(2)
	}
	var cmd *exec.Cmd
	if len(os.Args) == 2 {
		cmd = exec.CommandContext(ctx, os.Args[1])
	} else {
		cmd = exec.CommandContext(ctx, os.Args[1], os.Args[2:]...)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Stderr.WriteString("\n\nthe run command exited with an error; bailing")
		// actually really difficult to pass through the return code from Run so
		// just do 2
		os.Exit(2)
	}
	gitCmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	buf := new(bytes.Buffer)
	gitCmd.Stdout = buf
	gitCmd.Stderr = buf
	if err := gitCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\ndiffer: Error running git status --porcelain: %v\n\nOutput: %s",
			err, buf.String())
		os.Exit(2)
	}
	if buf.Len() > 0 {
		diff := getGitDiff(ctx)
		fmt.Fprintf(os.Stderr, `
Untracked or modified files present after running '%s':

%s%s
The command should not generate a diff. Please fix the problem and try again.
`, strings.Join(os.Args[1:], " "), buf.String(), diff)
		os.Exit(2)
	}
}
