package docker

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"

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

	digest     string
	digestErr  error
	digestOnce sync.Once

	ensureErr  error
	ensureOnce sync.Once

	uidGid     UIDGID
	uidGidErr  error
	uidGidOnce sync.Once
}

// Digest gets and returns the content digest for the image. Note that this is
// different from the "distribution digest" (which is what you can use to
// specify an image to `docker run`, as in `my/image@sha256:xxx`). We need to
// use the content digest because the distribution digest is only computed for
// images that have been pulled from or pushed to a registry. See
// https://windsock.io/explaining-docker-image-ids/ under "A Final Twist" for a
// good explanation.
func (image *image) Digest(ctx context.Context) (string, error) {
	image.digestOnce.Do(func() {
		image.digest, image.digestErr = func() (string, error) {
			if err := image.Ensure(ctx); err != nil {
				return "", err
			}

			// TODO!(sqs): is image id the right thing to use here? it is NOT
			// the digest. but the digest is not calculated for all images
			// (unless they are pulled/pushed from/to a registry), see
			// https://github.com/moby/moby/issues/32016.
			out, err := exec.CommandContext(ctx, "docker", "image", "inspect", "--format", "{{.Id}}", "--", image.name).CombinedOutput()
			if err != nil {
				return "", errors.Wrapf(err, "inspecting docker image: %s", string(bytes.TrimSpace(out)))
			}
			id := string(bytes.TrimSpace(out))
			if id == "" {
				return "", errors.Errorf("unexpected empty docker image content ID for %q", image.name)
			}
			return id, nil
		}()
	})

	return image.digest, image.digestErr
}

// Ensure ensures that the image has been pulled by Docker. Note that it does
// not attempt to pull a newer version of the image if it exists locally.
func (image *image) Ensure(ctx context.Context) error {
	image.ensureOnce.Do(func() {
		image.ensureErr = func() error {
			// docker image inspect will return a non-zero exit code if the image and
			// tag don't exist locally, regardless of the format.
			if err := exec.CommandContext(ctx, "docker", "image", "inspect", "--format", "1", image.name).Run(); err != nil {
				// Let's try pulling the image.
				if err := exec.CommandContext(ctx, "docker", "image", "pull", image.name).Run(); err != nil {
					return errors.Wrap(err, "pulling image")
				}
			}

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
