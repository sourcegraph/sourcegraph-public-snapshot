package images

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var imageRegex = regexp.MustCompile(``)

func UpdateShellManifests(ctx context.Context, registry Registry, path string, op UpdateOperation) error {
	var checked int
	if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".sh" {
			return nil
		}

		std.Out.WriteNoticef("Checking %q", path)

		shellFile, innerErr := os.ReadFile(path)
		if innerErr != nil {
			return errors.Wrapf(err, "couldn't read %s", path)
		}

		checked++
		newShellFile, innerErr := updateShellFile(registry, op, shellFile)
		if innerErr != nil {
			return err
		}
		if newShellFile == nil {
			std.Out.WriteSkippedf("No updates to make to %s", d.Name())
			return nil
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

func updateShellFile(registry Registry, op UpdateOperation, fileContent []byte) ([]byte, error) {
}
