package images

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// imageRegex will match a Docker image reference in a shell script of the following format:
//
// abc.domain.tld/something/else/imageName:tag@sha256:1234567890abcdef
var imageRegex = regexp.MustCompile(`[\w-]+\.\w+\.\w+/[:print:]+/.+:.+@sha256:[a-fA-F0-9]+`)

func UpdatePureDockerManifests(ctx context.Context, registry Registry, path string, op UpdateOperation) error {
	var checked int
	if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".sh" {
			return nil
		}

		std.Out.WriteNoticef("Checking %q", path)

		shellFile, innerErr := os.ReadFile(path)
		if innerErr != nil {
			return errors.Wrapf(innerErr, "couldn't read %s", path)
		}

		checked++
		newShellFile, innerErr := updatePureDockerFile(registry, op, shellFile)
		if innerErr != nil {
			if errors.Is(innerErr, ErrNoUpdateNeeded) {
				std.Out.WriteSkippedf("No updates to make to %s", d.Name())
				return nil
			} else {
				return innerErr
			}
		}

		if err := os.WriteFile(path, newShellFile, 0644); err != nil {
			return errors.Newf("WriteFile: %w", err)
		}

		std.Out.WriteSuccessf("%s updated!", path)
		return nil
	}); err != nil {
		return err
	}
	if checked == 0 {
		return errors.New("no valid shell files found")
	}

	return nil
}

func updatePureDockerFile(registry Registry, op UpdateOperation, fileContent []byte) ([]byte, error) {
	var outerErr error
	replaced := imageRegex.ReplaceAllFunc(fileContent, func(repl []byte) []byte {
		repo, err := ParseRepository(string(repl))
		if err != nil {
			if errors.Is(err, ErrNoUpdateNeeded) {
				outerErr = err
				return repl
			}
			outerErr = errors.Append(outerErr, errors.Wrapf(err, "couldn't parse %q", repl))
		}

		resultRepo, err := op(registry, repo)
		if err != nil {
			if errors.Is(err, ErrNoUpdateNeeded) {
				outerErr = err
				return repl
			}
			outerErr = errors.Append(outerErr, errors.Wrapf(err, "couldn't update %q", repl))
		}

		return []byte(resultRepo.Ref())
	})
	return replaced, outerErr
}
