package ci

import (
	"fmt"
	"time"

	"github.com/Masterminds/semver"

	bk "github.com/sourcegraph/sourcegraph/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/dev/ci/internal/ci/changed"
	"github.com/sourcegraph/sourcegraph/dev/ci/internal/ci/operations"
)

// CoreTestOperationsOptions should be used ONLY to adjust the behaviour of specific steps,
// e.g. by adding flags, and not as a condition for adding steps or commands.
type CoreTestOperationsOptions struct {
	// for clientChromaticTests
	ChromaticShouldAutoAccept bool
	MinimumUpgradeableVersion string
	ForceReadyForReview       bool

	CacheBundleSize      bool // for addWebAppEnterpriseBuild
	CreateBundleSizeDiff bool // for addWebAppEnterpriseBuild

	IsMainBranch bool

	// AspectWorkflows is set to true when we generate steps as part of the Aspect Workflows pipeline
	AspectWorkflows bool
}

// CoreTestOperations is a core set of tests that should be run in most CI cases. More
// notably, this is what is used to define operations that run on PRs. Please read the
// following notes:
//
//   - opts should be used ONLY to adjust the behaviour of specific steps, e.g. by adding
//     flags and not as a condition for adding steps or commands.
//   - be careful not to add duplicate steps.
//
// If the conditions for the addition of an operation cannot be expressed using the above
// arguments, please add it to the switch case within `GeneratePipeline` instead.
func CoreTestOperations(buildOpts bk.BuildOptions, diff changed.Diff, opts CoreTestOperationsOptions) *operations.Set {
	// Base set
	ops := operations.NewSet()

	// Simple, fast-ish linter checks
	ops.Append(BazelOperations(buildOpts, opts)...)
	linterOps := operations.NewNamedSet("Linters and static analysis")
	if targets := changed.GetLinterTargets(diff); len(targets) > 0 {
		linterOps.Append(addSgLints(targets))
	}
	ops.Merge(linterOps)

	if diff.Has(changed.Client | changed.GraphQL) {
		// If there are any Graphql changes, they are impacting the client as well.
		clientChecks := operations.NewNamedSet("Client checks",
			clientChromaticTests(opts),
			addJetBrainsUnitTests, // ~2.5m
			addStylelint,
		)
		ops.Merge(clientChecks)
	}

	return ops
}

func addJetBrainsUnitTests(pipeline *bk.Pipeline) {
	pipeline.AddStep(":java: Build (client/jetbrains)",
		withPnpmCache(),
		bk.Cmd("pnpm install --frozen-lockfile --fetch-timeout 60000"),
		bk.Cmd("pnpm generate"),
		bk.Cmd("pnpm --filter @sourcegraph/jetbrains run build"),
	)
}

// Adds a Buildkite pipeline "Wait".
func wait(pipeline *bk.Pipeline) {
	pipeline.AddWait()
}

func triggerReleaseBranchHealthchecks(minimumUpgradeableVersion string) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		version := semver.MustParse(minimumUpgradeableVersion)

		// HACK: we can't just subtract a single minor version once we roll over to 4.0,
		// so hard-code the previous minor version.
		previousMinorVersion := fmt.Sprintf("%d.%d", version.Major(), version.Minor()-1)
		if version.Major() == 4 && version.Minor() == 0 {
			previousMinorVersion = "3.43"
		} else if version.Major() == 5 && version.Minor() == 0 {
			previousMinorVersion = "4.5"
		}

		for _, branch := range []string{
			// Most recent major.minor
			fmt.Sprintf("%d.%d", version.Major(), version.Minor()),
			previousMinorVersion,
		} {
			name := fmt.Sprintf(":stethoscope: Trigger %s release branch healthcheck build", branch)
			pipeline.AddTrigger(name, "sourcegraph",
				bk.Async(false),
				bk.Build(bk.BuildOptions{
					Branch:  branch,
					Message: time.Now().Format(time.RFC1123) + " healthcheck build",
				}),
			)
		}
	}
}

func codeIntelQA(candidateTag string) operations.Operation {
	return func(p *bk.Pipeline) {
		p.AddStep(":bazel::docker::brain: Code Intel QA",
			bk.SlackStepNotify(&bk.SlackStepNotifyConfigPayload{
				Message:     ":alert: :noemi-handwriting: Code Intel QA Flake detected <@Noah S-C>",
				ChannelName: "code-intel-buildkite",
				Conditions: bk.SlackStepNotifyPayloadConditions{
					Failed: true,
				},
			}),
			// Run tests against the candidate server image
			bk.DependsOn(candidateImageStepKey("server")),
			bk.Agent("queue", "bazel"),
			bk.Env("CANDIDATE_VERSION", candidateTag),
			bk.Env("SOURCEGRAPH_BASE_URL", "http://127.0.0.1:7080"),
			bk.Env("SOURCEGRAPH_SUDO_USER", "admin"),
			bk.Env("TEST_USER_EMAIL", "test@sourcegraph.com"),
			bk.Env("TEST_USER_PASSWORD", "supersecurepassword"),
			bk.Cmd("dev/ci/integration/code-intel/run.sh"),
			bk.ArtifactPaths("./*.log"),
			bk.SoftFail(1))
	}
}
