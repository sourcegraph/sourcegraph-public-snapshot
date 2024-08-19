package run

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"bitbucket.org/creachadair/shell"
)

// BashOpt denotes options for running bash commands. For more options see 'man bash'
type BashOpt string

const (
	// 'pipefail' instructs bash to fail an entire statement if any command in a pipefails.
	BashOptPipeFail BashOpt = "pipefail"
	// 'errexit' lets bash exit with an err exit code if a command fails .
	BashOptErrExit BashOpt = "errexit"
)

// StrictBashOpts contains options that effectively enforce safe execution of bash commands.
var StrictBashOpts = []BashOpt{BashOptPipeFail, BashOptErrExit}

// Command builds a command for execution. Functions modify the underlying command.
type Command struct {
	ctx context.Context

	args    []string
	environ []string
	dir     string

	stdin  io.Reader
	attach attachedOutput

	// buildError represents an error that occured when building this command.
	buildError error
}

// Cmd joins all the parts and builds a command from it.
//
// Arguments are not implicitly quoted - to quote arguments, you can use Arg.
func Cmd(ctx context.Context, parts ...string) *Command {
	args, ok := shell.Split(strings.Join(parts, " "))
	if !ok {
		return &Command{buildError: errors.New("provided parts has unclosed quotes")}
	}

	return &Command{
		ctx:  ctx,
		args: args,
	}
}

// BashWith appends all the given bash options to the bash command with '-o'. The given parts
// is then joined together to be executed with 'bash -c'
//
// The final command will have the following format: bash -o option-1 -c command. For recommended strict bash options
// see StrictBashOpts, which has 'pipefail' and 'errexit' options
func BashWith(ctx context.Context, opts []BashOpt, parts ...string) *Command {
	var bash strings.Builder
	bash.WriteString("bash")
	for _, v := range opts {
		bash.WriteString(" -o ")
		bash.WriteString(string(v))
	}
	bash.WriteString(" -c")

	return Cmd(ctx, bash.String(), Arg(strings.Join(parts, " ")))
}

// Bash joins all the parts and builds a command from it to be run by 'bash -c'.
//
// Arguments are not implicitly quoted - to quote arguemnts, you can use Arg.
func Bash(ctx context.Context, parts ...string) *Command {
	return Cmd(ctx, "bash -c", Arg(strings.Join(parts, " ")))
}

// Run starts command execution and returns Output, which defaults to combined output.
func (c *Command) Run() Output {
	if c.buildError != nil {
		return NewErrorOutput(c.buildError)
	}
	if len(c.args) == 0 {
		return NewErrorOutput(errors.New("Command not instantiated"))
	}

	return attachAndRun(c.ctx, c.attach, c.stdin, ExecutedCommand{
		Args:    c.args,
		Environ: c.environ,
		Dir:     c.dir,
	})
}

// Dir sets the directory this command should be executed in.
func (c *Command) Dir(dir string) *Command {
	c.dir = dir
	return c
}

// Input pipes the given io.Reader to the command. If an input is already set, the given
// input is appended.
func (c *Command) Input(input io.Reader) *Command {
	if c.stdin != nil {
		c.stdin = io.MultiReader(c.stdin, input)
	} else {
		c.stdin = input
	}
	return c
}

// ResetInput sets the command's input to nil.
func (c *Command) ResetInput() *Command {
	c.stdin = nil
	return c
}

// Env adds the given environment variables to the command.
func (c *Command) Env(env map[string]string) *Command {
	for k, v := range env {
		c.environ = append(c.environ, fmt.Sprintf("%s=%s", k, v))
	}
	return c
}

// Environ adds the given strings representing the environment (key=value) to the
// command, for example os.Environ().
func (c *Command) Environ(environ []string) *Command {
	c.environ = append(c.environ, environ...)
	return c
}

// StdOut configures the command Output to only provide StdOut. By default, Output
// includes combined output.
func (c *Command) StdOut() *Command {
	c.attach = attachOnlyStdOut
	return c
}

// StdErr configures the command Output to only provide StdErr. By default, Output
// includes combined output.
func (c *Command) StdErr() *Command {
	c.attach = attachOnlyStdErr
	return c
}
