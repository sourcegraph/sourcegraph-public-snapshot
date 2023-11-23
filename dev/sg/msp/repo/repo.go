package repo

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/cliutil/completions"
)

// UseManagedServicesRepo is a cli.BeforeFunc that enforces that we are in the
// sourcegraph/managed-services repository by setting the current working
// directory.
func UseManagedServicesRepo(c *cli.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	repoRoot, err := repositoryRoot(cwd)
	if err != nil {
		return err
	}
	if repoRoot != cwd {
		std.Out.WriteSuggestionf("Using repo root %s as working directory", repoRoot)
		return os.Chdir(repoRoot)
	}
	return nil
}

func listServicesFromRoot(root string) ([]string, error) {
	var services []string
	return services, filepath.Walk(filepath.Join(root, "services"), func(path string, info fs.FileInfo, err error) error {
		if info == nil || info.Name() == "services" {
			return nil
		}
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if _, err := os.Stat(filepath.Join(root, ServiceYAMLPath(info.Name()))); err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		services = append(services, info.Name())
		return nil
	})
}

// ListServices returns a list of services, assuming MSP conventions in the
// working directory. Expected to be run after UseManagedServicesRepo() in a
// command context.
func ListServices() ([]string, error) {
	return listServicesFromRoot(".")
}

func ServicesCompletion() cli.BashCompleteFunc {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	repoRoot, err := repositoryRoot(cwd)
	if err != nil {
		return nil
	}
	return completions.CompleteOptions(func() (options []string) {
		services, _ := listServicesFromRoot(repoRoot)
		return services
	})
}

func ServiceYAMLPath(serviceID string) string {
	return filepath.Join("services", serviceID, "service.yaml")
}
