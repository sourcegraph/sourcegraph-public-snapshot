package docker

import (
	"bytes"
	"context"
	"fmt"
	goexec "os/exec"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/exec"
)

// UIDGID represents a UID:GID pair.
type UIDGID struct {
	UID int
	GID int
}

func (ug UIDGID) String() string {
	return fmt.Sprintf("%d:%d", ug.UID, ug.GID)
}

// Root is a root:root user.
var Root = UIDGID{UID: 0, GID: 0}

// Image represents a Docker image, hopefully stored in the local cache.
type Image interface {
	Digest(context.Context) (string, error)
	Ensure(context.Context) error
	UIDGID(context.Context) (UIDGID, error)
}

type image struct {
	name string

	// There are lots of once fields below: basically, we're going to try fairly
	// hard to prevent performing the same operations on the same image over and
	// over, since some of them are expensive.

	digest string

	ensureErr  error
	ensureOnce sync.Once

	uidGid     UIDGID
	uidGidErr  error
	uidGidOnce sync.Once
}

// Digest returns the content digest for the image. Note that this is different
// from the "distribution digest" (which is what you can use to specify an image
// to `docker run`, as in `my/image@sha256:xxx`). We need to use the content digest
// because the distribution digest is only computed for images that have been pulled
// from or pushed to a registry. See https://windsock.io/explaining-docker-image-ids/
// under "A Final Twist" for a good explanation.
func (image *image) Digest(ctx context.Context) (string, error) {
	ensureErr := image.Ensure(ctx)
	return image.digest, ensureErr
}

// Ensure ensures that the image has been pulled by Docker. Note that it does
// not attempt to pull a newer version of the image if it exists locally.
func (image *image) Ensure(ctx context.Context) error {
	image.ensureOnce.Do(func() {
		image.ensureErr = func() (err error) {
			inspectDigest := func() (string, error) {
				// Since we are only asking Docker for local information, we
				// expect this operation to be quick, and therefore set a
				// relatively low timeout for Docker to respond. This is
				// particularly useful because this function is usually the
				// first non-trivial interaction we have with Docker in a
				// src-cli invocation, and this allows us to catch failure modes
				// that result in the Docker socket still listening and
				// accepting connections, but where dockerd is no longer able to
				// respond to non-trivial requests.
				//
				// Anecdotally, this seems to happen most frequently with Docker
				// Desktop VMs running out of memory, whereupon the Linux
				// kernel's OOM killer sometimes chooses to kill components of
				// Docker instead of processes within containers.
				dctx, cancel, err := withFastCommandContext(ctx)
				if err != nil {
					return "", err
				}
				defer cancel()

				args := []string{"image", "inspect", "--format", "{{ .Id }}", image.name}
				out, err := exec.CommandContext(dctx, "docker", args...).Output()
				id := string(bytes.TrimSpace(out))

				if errors.IsDeadlineExceeded(err) || errors.IsDeadlineExceeded(dctx.Err()) {
					return "", newFastCommandTimeoutError(dctx, args...)
				} else if err != nil {
					return "", err
				}

				return id, nil
			}

			// docker image inspect will return a non-zero exit code if the image and
			// tag don't exist locally, regardless of the format.
			var digest string
			if digest, err = inspectDigest(); errors.HasType(err, &fastCommandTimeoutError{}) {
				// Ensure we immediately propagate a timeout up, rather than
				// trying to tell an unresponsive Docker to pull.
				return err
			} else if err != nil {
				// Let's try pulling the image.
				pullCmd := exec.CommandContext(ctx, "docker", "image", "pull", image.name)
				var stderr bytes.Buffer
				pullCmd.Stderr = &stderr
				if err := pullCmd.Run(); err != nil {
					exitErr := &goexec.ExitError{}
					if errors.As(err, &exitErr) {
						return errors.Newf("failed to pull image: %s\ndocker pull exited with code %d", stderr.String(), exitErr.ExitCode())
					}
					return errors.Wrap(err, "pulling image")
				}
				// And try again to get the image digest.
				digest, err = inspectDigest()
				if err != nil {
					return errors.Wrap(err, "not found after pulling image")
				}
			}

			if digest == "" {
				return errors.Errorf("unexpected empty docker image content ID for %q", image.name)
			}

			image.digest = digest

			return nil
		}()
	})

	return image.ensureErr
}

// UIDGID returns the user and group the container is configured to run as.
func (image *image) UIDGID(ctx context.Context) (UIDGID, error) {
	image.uidGidOnce.Do(func() {
		image.uidGid, image.uidGidErr = func() (UIDGID, error) {
			stdout := new(bytes.Buffer)

			// Digest also implicitly means Ensure has been called.
			digest, err := image.Digest(ctx)
			if err != nil {
				return UIDGID{}, errors.Wrap(err, "getting digest")
			}

			args := []string{
				"run",
				"--rm",
				"--entrypoint", "/bin/sh",
				digest,
				"-c", "id -u; id -g",
			}
			cmd := exec.CommandContext(ctx, "docker", args...)
			cmd.Stdout = stdout

			if err := cmd.Run(); err != nil {
				return UIDGID{}, errors.Wrap(err, "running id")
			}

			// POSIX specifies the output of `id -u` as the effective UID,
			// terminated by a newline. `id -g` is the same, just for the GID.
			raw := strings.TrimSpace(stdout.String())
			var res UIDGID
			_, err = fmt.Sscanf(raw, "%d\n%d", &res.UID, &res.GID)
			if err != nil {
				return res, errors.Wrapf(err, "malformed uid/gid: %q", raw)
			}
			return res, nil
		}()
	})

	return image.uidGid, image.uidGidErr
}
