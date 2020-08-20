package campaigns

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// getDockerImageContentDigest gets the content digest for the image. Note that this
// is different from the "distribution digest" (which is what you can use to specify
// an image to `docker run`, as in `my/image@sha256:xxx`). We need to use the
// content digest because the distribution digest is only computed for images that
// have been pulled from or pushed to a registry. See
// https://windsock.io/explaining-docker-image-ids/ under "A Final Twist" for a good
// explanation.
func getDockerImageContentDigest(ctx context.Context, image string) (string, error) {
	// TODO!(sqs): is image id the right thing to use here? it is NOT the
	// digest. but the digest is not calculated for all images (unless they are
	// pulled/pushed from/to a registry), see
	// https://github.com/moby/moby/issues/32016.
	out, err := exec.CommandContext(ctx, "docker", "image", "inspect", "--format", "{{.Id}}", "--", image).CombinedOutput()
	if err != nil {
		if !strings.Contains(string(out), "No such image") {
			return "", fmt.Errorf("error inspecting docker image %q: %s", image, bytes.TrimSpace(out))
		}
		pullCmd := exec.CommandContext(ctx, "docker", "image", "pull", image)

		err = pullCmd.Start()
		if err != nil {
			return "", fmt.Errorf("error pulling docker image %q: %s", image, err)
		}
		err = pullCmd.Wait()
		if err != nil {
			return "", fmt.Errorf("error pulling docker image %q: %s", image, err)
		}
	}
	out, err = exec.CommandContext(ctx, "docker", "image", "inspect", "--format", "{{.Id}}", "--", image).CombinedOutput()
	// This time, the image MUST be present, so the issue must be something else.
	if err != nil {
		return "", fmt.Errorf("error inspecting docker image %q: %s", image, bytes.TrimSpace(out))
	}
	id := string(bytes.TrimSpace(out))
	if id == "" {
		return "", fmt.Errorf("unexpected empty docker image content ID for %q", image)
	}
	return id, nil
}
