package images

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sourcegraph/conc/pool"
	"gopkg.in/yaml.v3"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func UpdateComposeManifests(ctx context.Context, registry Registry, path string, op UpdateOperation) error {
	var checked int
	if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".yaml" {
			return nil
		}

		std.Out.WriteNoticef("Checking %q", path)

		composeFile, innerErr := os.ReadFile(path)
		if innerErr != nil {
			return errors.Wrapf(err, "couldn't read %s", path)
		}

		checked++
		newComposeFile, innerErr := updateComposeFile(registry, op, composeFile)
		if innerErr != nil {
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

func updateComposeFile(registry Registry, op UpdateOperation, fileContent []byte) ([]byte, error) {
	var compose map[string]any
	if err := yaml.Unmarshal(fileContent, &compose); err != nil {
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

	checks := pool.NewWithResults[*replace]().WithMaxGoroutines(10).WithErrors()
	for name, entry := range services {
		name := name
		service, ok := entry.(map[string]any)
		if !ok {
			std.Out.WriteWarningf("%s: invalid service", name)
			continue
		}

		checks.Go(func() (*replace, error) {
			imageField, set := service["image"]
			if !set {
				std.Out.Verbosef("%s: no image", name)
				return nil, nil
			}
			originalImage, ok := imageField.(string)
			if !ok {
				std.Out.WriteWarningf("%s: invalid image", name)
				return nil, nil
			}

			r, err := ParseRepository(originalImage)
			if err != nil {
				if errors.Is(err, ErrNoUpdateNeeded) {
					std.Out.WriteLine(output.Styled(output.StyleWarning, fmt.Sprintf("skipping %q", originalImage)))
					return nil, nil
				} else {
					return nil, err
				}
			}

			newR, err := op(registry, r)
			if err != nil {
				if errors.Is(err, ErrNoUpdateNeeded) {
					std.Out.WriteLine(output.Styled(output.StyleWarning, fmt.Sprintf("skipping %q.", r.Ref())))
					return nil, nil
				} else {
					std.Out.WriteLine(output.Styled(output.StyleWarning, fmt.Sprintf("error on %q: %v", originalImage, err)))
					return nil, err
				}
			}

			std.Out.VerboseLine(output.Styledf(output.StylePending, "%s: will update to %s", name, newR.Ref()))
			return &replace{
				original: originalImage,
				new:      newR.Ref(),
			}, nil
		})
	}

	replaceOps, err := checks.Wait()
	if err != nil {
		return nil, err
	}
	var updates int
	for _, r := range replaceOps {
		if r == nil {
			continue
		}
		fileContent = bytes.ReplaceAll(fileContent, []byte(r.original), []byte(r.new))
		updates++
	}
	if updates == 0 {
		return nil, nil
	}

	return fileContent, nil
}
