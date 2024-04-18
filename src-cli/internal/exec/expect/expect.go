// Package expect uses the middleware concept in internal/exec to mock external
// commands.
//
// Generally, you only want to use this package in testing code.
//
// At a high level, this package operates by overriding created commands to
// invoke the current executable with a specific environment variable, which
// points to a temporary file with metadata on the behaviour that the command
// should implement. (This approach is shamelessly stolen from Go's os/exec test
// suite, but the details are somewhat different.)
//
// This means that the main() function of the executable _must_ check the
// relevant environment variable as early as possible, and not perform its usual
// logic if it's found. An implementation of this is provided for TestMain
// functions in the Handle function, which is normally how this package is used.
package expect

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	goexec "os/exec"
	"testing"

	"github.com/gobwas/glob"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/exec"
)

// envBehaviourFile is the environment variable used to define the behaviour
// file.
const envBehaviourFile = "GO_EXEC_TESTING_BEHAVIOUR_FILE"

// Behaviour defines the behaviour of the mocked command.
type Behaviour struct {
	// Stdout defines the data that will be returned on stdout.
	Stdout []byte

	// Stderr defines the data that will be returned on stderr.
	Stderr []byte

	// ExitCode is the exit code that the process will exit with.
	ExitCode int
}

// Success defines a command behaviour that returns a 0 exit code and no other
// output.
var Success = Behaviour{}

// Commands defines a set of expected commands for the given test. Commands may
// be called from any number of nested subtests, but must only be called once
// from a single test function, as it uses (*testing.T).Cleanup to manage
// per-test state.
func Commands(t *testing.T, exp ...*Expectation) {
	t.Helper()

	i := 0
	mc := exec.NewMiddleware(func(ctx context.Context, previous exec.CmdCreator, name string, arg ...string) *goexec.Cmd {
		if i >= len(exp) {
			t.Fatalf("one or more extra commands created in spite of expecting %d command(s); attempting to create %q with arguments %v", len(exp), name, arg)
		}

		if err := exp[i].Validator(name, arg...); err != nil {
			t.Fatalf("unexpected command: %v", err)
		}

		// Create the command using the next level of middleware. (Which is
		// probably eventually os/exec.CommandContext().)
		//
		// The prepending of ./ to the command name looks completely insane, but
		// there's a reason for it: if the name is a bare string like `docker`,
		// then Go will attempt to resolve it using $PATH. If the command
		// doesn't exist (because, say, we're running it in CI), then an error
		// is embedded within the *Cmd that will be returned when it is run,
		// even if we've subsequently rewritten the Path and Args fields to be
		// valid.
		//
		// Prepending ./ means that Go doesn't need to look the command up in
		// the $PATH and no error can be generated that way. Since we're going
		// to overwrite the Path momentarily anyway, that's fine.
		cmd := previous(ctx, "./"+name, arg...)
		if cmd == nil {
			t.Fatalf("unexpected nil *Cmd for %q with arguments %v", name, arg)
		}

		if len(os.Args) == 0 {
			t.Fatalf("unexpected empty os.Args")
		}

		// Now we modify the command to re-invoke this executable instead. We'll
		// also blank out the arguments, since this should be controlled
		// entirely by the presence of the behaviour file environment variable.
		cmd.Path = os.Args[0]
		cmd.Args = []string{}

		// Actually create the behaviour file.
		f, err := os.CreateTemp(os.TempDir(), "behaviour")
		if err != nil {
			t.Fatalf("error creating behaviour file: %v", err)
		}
		defer f.Close()
		file := f.Name()
		t.Cleanup(func() { os.Remove(file) })

		data, err := json.Marshal(&exp[i].Behaviour)
		if err != nil {
			t.Fatalf("error marshalling behaviour data: %v", err)
		}
		if _, err := f.Write(data); err != nil {
			t.Fatalf("writing data failed: %s", err)
		}

		// Set the relevant environment variable.
		cmd.Env = append(cmd.Env, envBehaviourFile+"="+file)

		i++
		return cmd
	})

	t.Cleanup(func() {
		mc.Remove()

		if i != len(exp) {
			t.Fatalf("unexpected number of commands executed: have=%d want=%d", i, len(exp))
		}
	})
}

// Handle should be called from TestMain. It intercepts expected commands and
// implements the expected behaviour.
//
// m is defined as an interface rather than *testing.M to make this usable from
// outside of a testing context.
func Handle(m interface{ Run() int }) int {
	if file := os.Getenv(envBehaviourFile); file != "" {
		panicErr := func(err error) {
			fmt.Fprintf(os.Stderr, "panic with error: %s", err.Error())
			// This exit code is chosen at random: obviously, a test that
			// expects a failure might just swallow this and be happy. We do the
			// best we can.
			os.Exit(255)
		}

		// Load up the expected behaviour of this command.
		data, err := os.ReadFile(file)
		if err != nil {
			panicErr(err)
		}

		var b Behaviour
		if err := json.Unmarshal(data, &b); err != nil {
			panicErr(err)
		}

		// Do it!
		os.Stderr.Write(b.Stderr)
		os.Stdout.Write(b.Stdout)
		os.Exit(b.ExitCode)
	}

	return m.Run()
}

// Expectation represents a single command invocation.
type Expectation struct {
	Validator CommandValidator
	Behaviour Behaviour
}

// CommandValidator is used to validate the command line that is would be
// executed.
type CommandValidator func(name string, arg ...string) error

// NewGlob is a convenience function that creates an Expectation that validates
// commands using a glob validator (as created by NewGlobValidator) and
// implements the given behaviour.
//
// You don't need to use this, but it tends to make Commands() calls more
// readable.
func NewGlob(behaviour Behaviour, wantName string, wantArg ...string) *Expectation {
	return &Expectation{
		Behaviour: behaviour,
		Validator: NewGlobValidator(wantName, wantArg...),
	}
}

// NewGlobValidator creates a validation function that will validate a command
// using glob syntax against the given name and arguments.
func NewGlobValidator(wantName string, wantArg ...string) CommandValidator {
	wantNameGlob := glob.MustCompile(wantName)
	wantArgGlobs := make([]glob.Glob, len(wantArg))
	for i, a := range wantArg {
		wantArgGlobs[i] = glob.MustCompile(a)
	}

	return func(haveName string, haveArg ...string) error {
		var errs errors.MultiError

		if !wantNameGlob.Match(haveName) {
			errs = errors.Append(errs, errors.Errorf("name does not match: have=%q want=%q", haveName, wantName))
		}

		if len(haveArg) != len(wantArgGlobs) {
			errs = errors.Append(errs, errors.Errorf("unexpected number of arguments:\nhave=%v\nwant=%v", haveArg, wantArg))
		} else {
			for i, g := range wantArgGlobs {
				if !g.Match(haveArg[i]) {
					errs = errors.Append(errs, errors.Errorf("unexpected argument at position %d:\nhave=%q\nwant=%q\ndiff=%q", i, haveArg[i], wantArg[i], cmp.Diff(haveArg[i], wantArg[i])))
				}
			}
		}

		return errs
	}
}

// NewLiteral is a convenience function that creates an Expectation that
// validates commands literally.
//
// You don't need to use this, but it tends to make Commands() calls more
// readable.
func NewLiteral(behaviour Behaviour, wantName string, wantArg ...string) *Expectation {
	return &Expectation{
		Behaviour: behaviour,
		Validator: func(haveName string, haveArg ...string) error {
			var errs errors.MultiError

			if wantName != haveName {
				errs = errors.Append(errs, errors.Errorf("name does not match: have=%q want=%q", haveName, wantName))
			}

			if diff := cmp.Diff(haveArg, wantArg); diff != "" {
				errs = errors.Append(errs, errors.Errorf("arguments do not match (-have +want):\n%s", diff))
			}

			return errs
		},
	}
}
