package repo

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/cliutil/completions"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// UseManagedServicesRepo is a cli.BeforeFunc that enforces that we are in the
// sourcegraph/managed-services repository by setting the current working
// directory.
func UseManagedServicesRepo(*cli.Context) error {
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

// DescribeServicesOptions returns a list of services for use in command
// descriptions, assuming MSP conventions somewhere in the path of the current
// working directory.
func DescribeServicesOptions() string {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("Could not list services: %s", err.Error())
	}
	repoRoot, err := repositoryRoot(cwd)
	if err != nil {
		return fmt.Sprintf("Could not list services: %s", err.Error())
	}
	services, err := listServicesFromRoot(repoRoot)
	if err != nil {
		return fmt.Sprintf("Could not list services: %s", err.Error())
	}
	if len(services) == 0 {
		return fmt.Sprintf("Could not list services: no services found, %s",
			ErrNotInsideManagedServices.Error())
	}
	return fmt.Sprintf("Available services:\n- %s",
		strings.Join(services, "\n- "))
}

// ServicesCompletions provides completions capabilities for commands that accept
// '<service ID>' positional argument. It traverses upwards to repo root and
// attempts to list services from there.
func ServicesCompletions(additionalArgs ...func(args cli.Args) (options []string)) cli.BashCompleteFunc {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	repoRoot, err := repositoryRoot(cwd)
	if err != nil {
		return nil
	}
	return completions.CompletePositionalArgs(append([]func(args cli.Args) (options []string){
		func(args cli.Args) (options []string) {
			services, _ := listServicesFromRoot(repoRoot)
			return services
		},
	}, additionalArgs...)...)
}

// ServicesAndEnvironmentsCompletion provides completions capabilities for
// commands that accept '<service ID> <environment ID>' positional arguments.
// It traverses upwards to repo root and attempts to list services from there.
func ServicesAndEnvironmentsCompletion(additionalArgs ...func(args cli.Args) (options []string)) cli.BashCompleteFunc {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	repoRoot, err := repositoryRoot(cwd)
	if err != nil {
		return nil
	}
	args := []func(cli.Args) []string{
		func(cli.Args) (options []string) {
			services, _ := listServicesFromRoot(repoRoot)
			return services
		},
		func(args cli.Args) (options []string) {
			svc, err := spec.Open(filepath.Join(repoRoot, ServiceYAMLPath(args.First())))
			if err != nil {
				// try to complete services as a fallback
				services, _ := listServicesFromRoot(repoRoot)
				return services
			}
			return svc.ListEnvironmentIDs()
		},
	}
	return completions.CompletePositionalArgs(append(args, additionalArgs...)...)
}

// ServiceYAMLPath returns the relative path to the service.yaml file for the
// given service.
//
// Requires UseManagedServicesRepo to be relevant.
func ServiceYAMLPath(serviceID string) string {
	return filepath.Join("services", serviceID, "service.yaml")
}

// ServiceEnvironmentYAMLPath returns the relative path to the Terraform Stacks
// directory for the given service environment's stack.
//
// Requires UseManagedServicesRepo to be relevant.
func ServiceStackPath(serviceID, envID, stackID string) string {
	return filepath.Join("services", serviceID, "terraform", envID, "stacks", stackID)
}

// ServiceStackTerraformPath returns the relative path to the Terraform CDKTF
// configuration file for the given service environment's stack.
//
// Requires UseManagedServicesRepo to be relevant.
func ServiceStackCDKTFPath(serviceID, envID, stackID string) string {
	return filepath.Join(ServiceStackPath(serviceID, envID, stackID), "cdk.tf.json")
}

// ToolingLockfileVersion retrieves the contents of the sg-msp lockfile for the
// given category (./sg-msp-$CATEGORY.lock).
//
// Requires UseManagedServicesRepo.
func ToolingLockfileVersion(category spec.EnvironmentCategory) (string, error) {
	lockfile := fmt.Sprintf("sg-msp-%s.lock", category)
	if category == "" {
		lockfile = "sg-msp.lock" // fallback to the old format (no category)
	}

	contents, err := os.ReadFile(lockfile)
	if err != nil {
		// Try to fall back to category-less-lockfile
		if v, fallbackErr := ToolingLockfileVersion(""); fallbackErr == nil {
			return v, nil
		}
		// Otherwise, return the error we got.
		return "", errors.Wrapf(err, "read %q", lockfile)
	}

	version := strings.TrimSpace(string(contents))
	if len(version) == 0 {
		return "", errors.Newf("empty %q", lockfile)
	}
	return version, nil
}

// GitRevision gets the revision of the managed-services repository.
//
// Requires UseManagedServicesRepo.
func GitRevision(ctx context.Context) (string, error) {
	return run.Cmd(ctx, "git rev-parse HEAD").
		Environ(append(os.Environ(),
			// Options copy-pasta from dev/sg/internal/run
			// Don't use the system wide git config.
			"GIT_CONFIG_NOSYSTEM=1",
			// And also not any other, because they can mess up output, change defaults, .. which can do unexpected things.
			"GIT_CONFIG=/dev/null")).
		Run().
		String()
}
