package msp

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/urfave/cli/v2"
	"golang.org/x/exp/maps"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/clouddeploy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/operationdocs/diagram"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/operationdocs/terraform"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/stacks/cloudrun"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/terraformcloud"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	msprepo "github.com/sourcegraph/sourcegraph/dev/sg/msp/repo"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// useServiceArgument retrieves the service spec corresponding to the first
// argument.
//
// 'exact' indicates that no additional arguments are expected.
func useServiceArgument(c *cli.Context, exact bool) (*spec.Spec, error) {
	// If we can successfully load the list of services, provide the
	// list as feedback for the user
	allServices, _ := msprepo.ListServices()

	serviceID := c.Args().First()
	if serviceID == "" {
		if len(allServices) > 0 {
			return nil, errors.Newf("argument service is required, available services: [%s]",
				strings.Join(allServices, ", "))
		}
		return nil, errors.New("argument service is required")
	}
	serviceSpecPath := msprepo.ServiceYAMLPath(serviceID)

	s, err := spec.Open(serviceSpecPath)
	if err != nil {
		if errors.Is(err, spec.ErrServiceDoesNotExist) {
			if len(allServices) > 0 {
				return nil, errors.Newf("service %q does not exist, available services: [%s]",
					serviceID, strings.Join(allServices, ", "))
			}
			return nil, errors.Newf("service %q does not exist", serviceID)
		}
		return nil, errors.Wrapf(err, "load service %q", serviceID)
	}

	// Arg 0 is service, arg 1 is environment - any additional arguments are
	// unexpected if we are getting exact arguments.
	if exact && c.Args().Get(1) != "" {
		return s, errors.Newf("got unexpected additional arguments %q - note that flags must be placed BEFORE arguments, i.e. '<flags> <arguments>'",
			strings.Join(c.Args().Slice()[1:], " "))
	}

	return s, nil
}

// useServiceAndEnvironmentArguments retrieves the service and environment specs
// corresponding to the first and second arguments respectively. It should only
// be used if both arguments are required.
func useServiceAndEnvironmentArguments(c *cli.Context, exact bool) (*spec.Spec, *spec.EnvironmentSpec, error) {
	svc, err := useServiceArgument(c, false)
	if err != nil {
		return nil, nil, err
	}

	environmentID := c.Args().Get(1)
	if environmentID == "" {
		return svc, nil, errors.Newf("second argument <environment ID> is required, available environments for service %q: [%s]",
			svc.Service.ID, strings.Join(svc.ListEnvironmentIDs(), ", "))
	}

	env := svc.GetEnvironment(environmentID)
	if env == nil {
		return svc, nil, errors.Newf("environment %q not found in the %q service spec, available environments: [%s]",
			environmentID, svc.Service.ID, strings.Join(svc.ListEnvironmentIDs(), ", "))
	}

	// Arg 0 is service, arg 1 is environment - any additional arguments are
	// unexpected if we are getting exact arguments.
	if exact && c.Args().Get(2) != "" {
		return svc, env, errors.Newf("got unexpected additional arguments %q - note that flags must be placed BEFORE arguments, i.e. '<flags> <arguments>'",
			strings.Join(c.Args().Slice()[2:], " "))
	}

	return svc, env, nil
}

func syncEnvironmentWorkspaces(c *cli.Context, tfc *terraformcloud.Client, service spec.ServiceSpec, env spec.EnvironmentSpec) error {
	if c.Bool("delete") {
		if !pointers.DerefZero(env.AllowDestroys) {
			return errors.Newf("environments[%s].allowDestroys must be 'true' to delete workspaces", env.ID)
		}

		std.Out.Promptf("[%s] Deleting workspaces for environment %q - are you sure? (y/N) ",
			service.ID, env.ID)
		var input string
		if _, err := fmt.Scan(&input); err != nil {
			return err
		}
		if input != "y" {
			return errors.New("aborting")
		}

		pending := std.Out.Pending(output.Styledf(output.StylePending,
			"[%s] Deleting Terraform Cloud workspaces for environment %q", service.ID, env.ID))

		// Destroy stacks in reverse order
		stacks := managedservicesplatform.StackNames()
		slices.Reverse(stacks)
		if errs := tfc.DeleteWorkspaces(c.Context, service, env, stacks); len(errs) > 0 {
			for _, err := range errs {
				std.Out.WriteWarningf(err.Error())
			}
			return errors.New("some errors occurred when deleting workspaces")
		}
		pending.Complete(output.Styledf(output.StyleSuccess,
			"[%s] Deleting Terraform Cloud workspaces for environment %q", service.ID, env.ID))

		return nil // exit early for deletion, we are done
	}

	pending := std.Out.Pending(output.Styledf(output.StylePending,
		"[%s] Synchronizing Terraform Cloud workspaces for environment %q", service.ID, env.ID))
	workspaces, err := tfc.SyncWorkspaces(c.Context, service, env, managedservicesplatform.StackNames())
	if err != nil {
		return errors.Wrap(err, "sync Terraform Cloud workspace")
	}
	pending.Complete(output.Styledf(output.StyleSuccess,
		"[%s] Synchronized Terraform Cloud workspaces for environment %q", service.ID, env.ID))

	var summary strings.Builder
	for _, ws := range workspaces {
		summary.WriteString(fmt.Sprintf("- %s: %s", ws.Name, ws.URL()))
		if ws.Created {
			summary.WriteString(" (created)")
		} else {
			summary.WriteString(" (updated)")
		}
		summary.WriteString("\n")
	}
	return std.Out.WriteMarkdown(summary.String())
}

type toolingLockfileChecker struct {
	version    string
	categories map[spec.EnvironmentCategory]*sync.Once
}

// checkCategoryVersion performs warning checks for the given environment category's
// tooling version.
//
// Requires UseManagedServicesRepo.
func (c *toolingLockfileChecker) checkCategoryVersion(out *std.Output, category spec.EnvironmentCategory) {
	var categoryOnce *sync.Once
	if o, ok := c.categories[category]; ok {
		categoryOnce = o
	} else {
		categoryOnce = &sync.Once{}
		c.categories[category] = categoryOnce
	}

	categoryOnce.Do(func() {
		lockedSgVersion, err := msprepo.ToolingLockfileVersion(category)
		if err != nil {
			out.WriteWarningf("Unable to determine locked 'sg' version for category %q: %s",
				category, err.Error())
		} else if lockedSgVersion != c.version {
			out.WriteWarningf(`Lockfile for category %q declares 'sg' version %q, you are using %q - generated outputs may differ from what is expected.
If there is a diff in the generated output, try running the following:`,
				category, lockedSgVersion, c.version)
			_ = out.WriteCode("bash", fmt.Sprintf(
				"sg update -release %q &&\n  SG_SKIP_AUTO_UPDATE=true sg msp generate -all -category %q",
				lockedSgVersion, string(category)))
		}
	})
}

type generateTerraformOptions struct {
	// tooling is used to validate the current tooling version matches what
	// is expected, and warn the user if there is a mismatch.
	tooling *toolingLockfileChecker
	// targetEnv generates the specified env only, otherwise generates all
	targetEnv string
	// targetCategory generates the specified category only
	targetCategory spec.EnvironmentCategory
	// stableGenerate disables updating of any values that are evaluated at
	// generation time
	stableGenerate bool
}

func generateTerraform(service *spec.Spec, opts generateTerraformOptions) error {
	serviceID := service.Service.ID
	serviceSpecPath := msprepo.ServiceYAMLPath(serviceID)

	var envs []spec.EnvironmentSpec
	if opts.targetEnv != "" {
		deployEnv := service.GetEnvironment(opts.targetEnv)
		if deployEnv == nil {
			return errors.Newf("environment %q not found in service spec", opts.targetEnv)
		}
		envs = append(envs, *deployEnv)
	} else {
		envs = service.Environments
	}

	for _, env := range envs {
		env := env

		if opts.targetCategory != "" && env.Category != opts.targetCategory {
			// Quietly skip environments that don't match specified category
			std.Out.WriteLine(output.StyleSuggestion.Linef(
				"[%s] Skipping non-%q environment %q (category %q)",
				serviceID, opts.targetCategory, env.ID, env.Category))
			continue
		}

		// Check tooling version and emit warnings
		opts.tooling.checkCategoryVersion(std.Out, env.Category)

		// Then, start our actual work
		pending := std.Out.Pending(output.StylePending.Linef(
			"[%s] Preparing Terraform for %q environment %q", serviceID, env.Category, env.ID))
		renderer := managedservicesplatform.Renderer{
			OutputDir:      filepath.Join(filepath.Dir(serviceSpecPath), "terraform", env.ID),
			StableGenerate: opts.stableGenerate,
		}

		// CDKTF needs the output dir to exist ahead of time, even for
		// rendering. If it doesn't exist yet, create it
		if f, err := os.Lstat(renderer.OutputDir); err != nil {
			if !os.IsNotExist(err) {
				return errors.Wrap(err, "check output directory")
			}
			if err := os.MkdirAll(renderer.OutputDir, 0755); err != nil {
				return errors.Wrap(err, "prepare output directory")
			}
		} else if !f.IsDir() {
			return errors.Newf("output directory %q is not a directory", renderer.OutputDir)
		}

		// Render environment
		cdktf, err := renderer.RenderEnvironment(*service, env)
		if err != nil {
			return err
		}

		pending.Updatef("[%s] Generating Terraform assets in %q for %q environment %q...",
			serviceID, renderer.OutputDir, env.Category, env.ID)
		if err := cdktf.Synthesize(); err != nil {
			return err
		}

		// Generate additional rollouts assets IFF this is the final stage of
		// a rollout pipeline.
		if rollout := service.BuildRolloutPipelineConfiguration(env); rollout.IsFinalStage() {
			pending.Updatef("[%s] Building rollout pipeline configurations for environment %q...", serviceID, env.ID)

			// We generate skaffold.yaml archive for upload to GCS. See
			// cloudrun.ScaffoldSourceFile docstring for more on why we need
			// to generate this separately. This step will likely always be
			// rquired.
			skaffoldObject, err := clouddeploy.NewCloudRunCustomTargetSkaffoldAssetsArchive()
			if err != nil {
				return errors.Wrap(err, "create Cloud Deploy custom target skaffold YAML archive")
			}
			skaffoldObjectPath := filepath.Join(renderer.OutputDir, "stacks/cloudrun", cloudrun.ScaffoldSourceFile)
			if err := os.WriteFile(skaffoldObjectPath, skaffoldObject.Bytes(), 0644); err != nil {
				return errors.Wrap(err, "write Cloud Run custom target skaffold YAML archive")
			}
		}

		// We must persist diagrams somewhere for reference in Notion, since
		// Notion does not allow us to upload files via API.
		// https://developers.notion.com/docs/working-with-files-and-media#uploading-files-and-media-via-the-notion-api
		pending.Updatef("[%s] Generating architecture diagrams for environment %q...", serviceID, env.ID)
		d, err := diagram.New()
		if err != nil {
			return errors.Wrap(err, "initialize architecture diagram")
		}
		if err = d.Generate(service, env.ID); err != nil {
			return errors.Wrap(err, "generate architecture diagram")
		}
		svg, err := d.Render()
		if err != nil {
			return errors.Wrap(err, "render architecture diagram")
		}

		diagramDir := filepath.Join(filepath.Dir(serviceSpecPath), "diagrams")
		_ = os.MkdirAll(diagramDir, os.ModePerm)

		diagramFileName := fmt.Sprintf("%s.svg", env.ID)
		if err := os.WriteFile(filepath.Join(diagramDir, diagramFileName), svg, 0o644); err != nil {
			return errors.Wrap(err, "write architecture diagram")
		}

		// GitHub file view for SVG sucks, so also generate a Markdown file
		// with nothing but a view of the image, for easier viewing in GitHub
		// and linking from Notion.
		diagramViewFileName := fmt.Sprintf("%s.md", env.ID)
		if err := os.WriteFile(
			filepath.Join(diagramDir, diagramViewFileName),
			[]byte(fmt.Sprintf("![architecture diagram](%s)\n", diagramFileName)),
			0o644,
		); err != nil {
			return errors.Wrap(err, "write architecture diagram view")
		}

		pending.Complete(output.Styledf(output.StyleSuccess,
			"[%s] Category %q environment %q infrastructure assets generated!",
			serviceID, env.Category, env.ID))
	}

	return nil
}

func collectAlertPolicies(svc *spec.Spec) (map[string]terraform.AlertPolicy, error) {
	// Deduplicate alerts across environments into a single map
	collectedAlerts := make(map[string]terraform.AlertPolicy)
	for _, env := range svc.ListEnvironmentIDs() {
		// Parse the generated alert policies to create alerting docs
		monitoringPath := msprepo.ServiceStackCDKTFPath(svc.Service.ID, env, "monitoring")
		monitoring, err := terraform.ParseMonitoringCDKTF(monitoringPath)
		if err != nil {
			return nil, err
		}
		maps.Copy(collectedAlerts, monitoring.ResourceType.GoogleMonitoringAlertPolicy)
	}
	return collectedAlerts, nil
}

// sortSlice sorts a slice of elements and returns it, for ease of chaining.
func sortSlice[S ~[]E, E cmp.Ordered](s S) S {
	slices.Sort(s)
	return s
}

// maybeAddSuggestion adds suggestions to errors that are known to be related to
// problems that can be resolved by referring to the service Notion page. If
// the service doesn't have one, or if the error doesn't match any known patterns,
// the error is returned as-is.
func maybeAddSuggestion(svc spec.ServiceSpec, err error) error {
	if svc.NotionPageID == nil {
		return err
	}
	if strings.Contains(err.Error(), "PermissionDenied") {
		return errors.Wrapf(err, "possible permissions error, ensure you have the prerequisite Entitle grants mentioned in %s",
			svc.GetHandbookPageURL())
	}
	return err
}
