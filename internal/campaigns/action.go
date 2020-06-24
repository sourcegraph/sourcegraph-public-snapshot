package campaigns

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/src-cli/schema"
	"github.com/xeipuuv/gojsonschema"
)

type Action struct {
	ScopeQuery string        `json:"scopeQuery,omitempty"`
	Steps      []*ActionStep `json:"steps"`
}

type ActionStep struct {
	Type      string   `json:"type"`            // "command"
	Image     string   `json:"image,omitempty"` // Docker image
	CacheDirs []string `json:"cacheDirs,omitempty"`
	Args      []string `json:"args,omitempty"`

	// ImageContentDigest is an internal field that should not be set by users.
	ImageContentDigest string
}

type PatchInput struct {
	Repository   string `json:"repository"`
	BaseRevision string `json:"baseRevision"`
	BaseRef      string `json:"baseRef"`
	Patch        string `json:"patch"`
}

type ActionRepo struct {
	ID      string
	Name    string
	Rev     string
	BaseRef string
}

func ValidateActionDefinition(def []byte) error {
	sl := gojsonschema.NewSchemaLoader()
	sc, err := sl.Compile(gojsonschema.NewStringLoader(schema.ActionSchemaJSON))
	if err != nil {
		return errors.Wrapf(err, "failed to compile actions schema")
	}

	normalized, err := jsonxToJSON(string(def))
	if err != nil {
		return err
	}

	res, err := sc.Validate(gojsonschema.NewBytesLoader(normalized))
	if err != nil {
		return errors.Wrap(err, "failed to validate config against schema")
	}

	errs := &multierror.Error{ErrorFormat: formatValidationErrs}
	for _, err := range res.Errors() {
		e := err.String()
		// Remove `(root): ` from error formatting since these errors are
		// presented to users.
		e = strings.TrimPrefix(e, "(root): ")
		errs = multierror.Append(errs, errors.New(e))
	}

	return errs.ErrorOrNil()
}

func formatValidationErrs(es []error) string {
	points := make([]string, len(es))
	for i, err := range es {
		points[i] = fmt.Sprintf("- %s", err)
	}

	return fmt.Sprintf(
		"Validating action definition failed:\n%s\n",
		strings.Join(points, "\n"))
}

func PrepareAction(ctx context.Context, action Action, logger *ActionLogger) error {
	// Build any Docker images.
	for _, step := range action.Steps {
		if step.Type == "docker" {
			// Set digests for Docker images so we don't cache action runs in 2 different images with
			// the same tag.
			var err error
			step.ImageContentDigest, err = getDockerImageContentDigest(ctx, step.Image, logger)
			if err != nil {
				return errors.Wrap(err, "Failed to get Docker image content digest")
			}
		}
	}

	return nil
}

// getDockerImageContentDigest gets the content digest for the image. Note that this
// is different from the "distribution digest" (which is what you can use to specify
// an image to `docker run`, as in `my/image@sha256:xxx`). We need to use the
// content digest because the distribution digest is only computed for images that
// have been pulled from or pushed to a registry. See
// https://windsock.io/explaining-docker-image-ids/ under "A Final Twist" for a good
// explanation.
func getDockerImageContentDigest(ctx context.Context, image string, logger *ActionLogger) (string, error) {
	// TODO!(sqs): is image id the right thing to use here? it is NOT the
	// digest. but the digest is not calculated for all images (unless they are
	// pulled/pushed from/to a registry), see
	// https://github.com/moby/moby/issues/32016.
	out, err := exec.CommandContext(ctx, "docker", "image", "inspect", "--format", "{{.Id}}", "--", image).CombinedOutput()
	if err != nil {
		if !strings.Contains(string(out), "No such image") {
			return "", fmt.Errorf("error inspecting docker image %q: %s", image, bytes.TrimSpace(out))
		}
		logger.Infof("Pulling Docker image %q...\n", image)
		pullCmd := exec.CommandContext(ctx, "docker", "image", "pull", image)
		prefix := fmt.Sprintf("docker image pull %s", image)
		pullCmd.Stdout = logger.InfoPipe(prefix)
		pullCmd.Stderr = logger.ErrorPipe(prefix)

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

// jsonxToJSON converts jsonx to plain JSON.
func jsonxToJSON(text string) ([]byte, error) {
	data, errs := jsonx.Parse(text, jsonx.ParseOptions{Comments: true, TrailingCommas: true})
	if len(errs) > 0 {
		return data, fmt.Errorf("failed to parse JSON: %v", errs)
	}
	return data, nil
}
