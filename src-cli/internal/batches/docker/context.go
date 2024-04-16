package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/kballard/go-shellquote"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// withFastCommandContext wraps the given context with a timeout appropriate for
// invoking Docker commands that are expected to be fast, such as `docker info`
// and `docker image inspect`. This defaults to 5 seconds, but can be overridden
// by the undocumented $SRC_DOCKER_FAST_COMMAND_TIMEOUT environment variable.
//
// If the context deadline is exceeded, the code using the context can pass the
// context and Docker arguments to newFastCommandTimeoutError to get a nicely
// formatted error for the user.
func withFastCommandContext(ctx context.Context) (context.Context, context.CancelFunc, error) {
	timeout, err := fastCommandTimeout()
	if err != nil {
		return nil, nil, err
	}

	fctx, cancel := context.WithTimeout(
		context.WithValue(ctx, fastCommandTimeoutEnv, timeout),
		timeout,
	)
	return fctx, cancel, nil
}

type fastCommandTimeoutError struct {
	args    []string
	timeout time.Duration
}

func newFastCommandTimeoutError(ctx context.Context, args ...string) error {
	// Attempt to extract the timeout from the context.
	timeout, ok := ctx.Value(fastCommandTimeoutEnv).(time.Duration)
	if !ok {
		return errors.Newf(
			"additional error found when attempting to create fastCommandTimeoutError: "+
				"no timeout was set within the context, so the context probably wasn't wrapped "+
				"with withFastCommandContext (please file a bug report on src-cli!): "+
				"the original error involved invoking docker with these args: %q",
			args,
		)
	}

	return &fastCommandTimeoutError{
		args:    args,
		timeout: timeout,
	}
}

func (e *fastCommandTimeoutError) Error() string {
	return fmt.Sprintf(
		"`docker %s` failed to respond within %s; "+
			"please verify that Docker has been started and is responding normally",
		shellquote.Join(e.args...), e.timeout,
	)
}

func (*fastCommandTimeoutError) Timeout() bool { return true }

type fastCommandTimeoutKey string

const (
	fastCommandTimeoutDefault                       = 5 * time.Second
	fastCommandTimeoutEnv     fastCommandTimeoutKey = "SRC_DOCKER_FAST_COMMAND_TIMEOUT"
)

var fastCommandTimeoutData = struct {
	once    sync.Once
	timeout time.Duration
	err     error
}{
	timeout: fastCommandTimeoutDefault,
	err:     nil,
}

func fastCommandTimeout() (time.Duration, error) {
	fastCommandTimeoutData.once.Do(func() {
		if userTimeout, ok := os.LookupEnv(string(fastCommandTimeoutEnv)); ok {
			parsed, err := time.ParseDuration(userTimeout)
			if err != nil {
				fastCommandTimeoutData.err = errors.Wrapf(err, "parsing timeout duration from environment variable %s", fastCommandTimeoutEnv)
			} else {
				fastCommandTimeoutData.timeout = parsed
			}
		}
	})

	return fastCommandTimeoutData.timeout, fastCommandTimeoutData.err
}

// executeFastCommand creates a fastCommandContext used to execute docker commands
// with a timeout for docker commands that are supposed to be fast (e.g docker info).
func executeFastCommand(ctx context.Context, args ...string) ([]byte, error) {
	dctx, cancel, err := withFastCommandContext(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()

	out, err := exec.CommandContext(dctx, "docker", args...).CombinedOutput()
	if errors.IsDeadlineExceeded(err) || errors.IsDeadlineExceeded(dctx.Err()) {
		return nil, newFastCommandTimeoutError(dctx, args...)
	}

	return out, err
}
