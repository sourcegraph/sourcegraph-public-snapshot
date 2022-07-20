package images

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker-credential-helpers/credentials"
	"gopkg.in/yaml.v3"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/group"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// UpdateCompose walks all `*docker-compose.yaml` files and updates Sourcegraph images
// in each.
func UpdateCompose(path string, creds credentials.Credentials, pinTag string) error {
	var checked int
	if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if !strings.Contains(d.Name(), "docker-compose.yaml") {
			std.Out.WriteWarningf("%s is not a docker-compose.yaml file but we will still try to update it anyway", path)
		}

		std.Out.WriteNoticef("Checking %q", path)

		composeFile, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "couldn't read %s", path)
		}

		checked++
		newComposeFile, err := updateComposeFile(composeFile, creds, pinTag)
		if err != nil {
			return err
		}
		if newComposeFile == nil {
			std.Out.WriteSkippedf("No updates to make to %s", d.Name())
			return nil
		}

		if err := os.WriteFile(path, newComposeFile, 0644); err != nil {
			return errors.Newf("WriteFile: %w", err)
		}

		std.Out.WriteSuccessf("%s updated!", path)
		return nil
	}); err != nil {
		return err
	}
	if checked == 0 {
		return errors.New("no valid docker-compose files found")
	}

	return nil
}

// updateComposeFile updates composeFile data and returns it.
func updateComposeFile(composeFile []byte, creds credentials.Credentials, pinTag string) ([]byte, error) {
	var compose map[string]any
	if err := yaml.Unmarshal(composeFile, &compose); err != nil {
		return nil, err
	}
	services, ok := compose["services"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid services")
	}

	type replace struct {
		original string
		new      string
	}
	checks := group.NewWithResults[*replace]().WithMaxConcurrency(10)
	for name, entry := range services {
		name := name
		service, ok := entry.(map[string]any)
		if !ok {
			std.Out.WriteWarningf("%s: invalid service", name)
			continue
		}

		checks.Go(func() *replace {
			originalImage, ok := service["image"].(string)
			if !ok {
				std.Out.WriteWarningf("%s: invalid image", name)
				return nil
			}

			newImage, err := getUpdatedSourcegraphImage(originalImage, creds, pinTag)
			if err != nil {
				std.Out.WriteWarningf("%s: %s", name, err)
				return nil
			}

			std.Out.VerboseLine(output.Styledf(output.StylePending, "%s: will update to %s", name, newImage))
			return &replace{
				original: originalImage,
				new:      newImage,
			}
		})
	}

	replaceOps := checks.Wait()
	var updates int
	for _, r := range replaceOps {
		if r == nil {
			continue
		}
		composeFile = bytes.ReplaceAll(composeFile, []byte(r.original), []byte(r.new))
		updates++
	}
	if updates == 0 {
		return nil, nil
	}

	return composeFile, nil
}
