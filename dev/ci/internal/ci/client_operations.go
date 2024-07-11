package ci

import (
	"fmt"

	bk "github.com/sourcegraph/sourcegraph/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/dev/ci/internal/ci/operations"
)

func addStylelint(pipeline *bk.Pipeline) {
	pipeline.AddStep(":stylelint: Stylelint (all)",
		withPnpmCache(),
		bk.Cmd("dev/ci/pnpm-run.sh lint:css:all"))
}

var browsers = []string{"chrome"}

func addBrowserExtensionIntegrationTests() operations.Operation {
	return func(pipeline *bk.Pipeline) {
		for _, browser := range browsers {
			pipeline.AddStep(
				fmt.Sprintf(":%s: Puppeteer tests for %s extension", browser, browser),
				withPnpmCache(),
				bk.Env("EXTENSION_PERMISSIONS_ALL_URLS", "true"),
				bk.Env("BROWSER", browser),
				bk.Env("LOG_BROWSER_CONSOLE", "false"),
				bk.Env("SOURCEGRAPH_BASE_URL", "https://sourcegraph.com"),
				bk.Env("POLLYJS_MODE", "replay"), // ensure that we use existing recordings
				bk.Cmd("pnpm install --frozen-lockfile --fetch-timeout 60000"),
				bk.Cmd("pnpm --filter @sourcegraph/browser run build"),
				bk.Cmd("pnpm run test-browser-integration"),
				bk.ArtifactPaths("./puppeteer/*.png"),
			)
		}
	}
}

func recordBrowserExtensionIntegrationTests(pipeline *bk.Pipeline) {
	for _, browser := range browsers {
		pipeline.AddStep(
			fmt.Sprintf(":%s: Puppeteer tests for %s extension", browser, browser),
			withPnpmCache(),
			bk.Env("EXTENSION_PERMISSIONS_ALL_URLS", "true"),
			bk.Env("BROWSER", browser),
			bk.Env("LOG_BROWSER_CONSOLE", "false"),
			bk.Env("SOURCEGRAPH_BASE_URL", "https://sourcegraph.com"),
			bk.Cmd("pnpm install --frozen-lockfile --fetch-timeout 60000"),
			bk.Cmd("pnpm --filter @sourcegraph/browser run build"),
			bk.Cmd("pnpm --filter @sourcegraph/browser run record-integration"),
			// Retry may help in case if command failed due to hitting the rate limit or similar kind of error on the code host:
			// https://docs.github.com/en/rest/reference/search#rate-limit
			bk.AutomaticRetry(1),
			bk.ArtifactPaths("./puppeteer/*.png"),
		)
	}
}

func addBrowserExtensionE2ESteps(pipeline *bk.Pipeline) {
	for _, browser := range []string{"chrome"} {
		// Run e2e tests
		pipeline.AddStep(fmt.Sprintf(":%s: E2E for %s extension", browser, browser),
			withPnpmCache(),
			bk.Env("EXTENSION_PERMISSIONS_ALL_URLS", "true"),
			bk.Env("BROWSER", browser),
			bk.Env("LOG_BROWSER_CONSOLE", "true"),
			bk.Env("SOURCEGRAPH_BASE_URL", "https://sourcegraph.com"),
			bk.Cmd("pnpm install --frozen-lockfile --fetch-timeout 60000"),
			bk.Cmd("pnpm --filter @sourcegraph/browser run build"),
			bk.Cmd("pnpm mocha ./client/browser/src/end-to-end/github.test.ts ./client/browser/src/end-to-end/gitlab.test.ts"),
			bk.ArtifactPaths("./puppeteer/*.png"))
	}
}

// Release the browser extension.
func addBrowserExtensionReleaseSteps(pipeline *bk.Pipeline) {
	addBrowserExtensionE2ESteps(pipeline)

	pipeline.AddWait()

	// Release to the Chrome Webstore
	pipeline.AddStep(":rocket::chrome: Extension release",
		withPnpmCache(),
		bk.Cmd("pnpm install --frozen-lockfile --fetch-timeout 60000"),
		bk.Cmd("pnpm --filter @sourcegraph/browser run build"),
		bk.Cmd("pnpm --filter @sourcegraph/browser release:chrome"))

	// Build and self sign the FF add-on and upload it to a storage bucket
	pipeline.AddStep(":rocket::firefox: Extension release",
		withPnpmCache(),
		bk.Cmd("pnpm install --frozen-lockfile --fetch-timeout 60000"),
		bk.Cmd("pnpm --filter @sourcegraph/browser release:firefox"))

	// Release to npm
	pipeline.AddStep(":rocket::npm: npm Release",
		withPnpmCache(),
		bk.Cmd("pnpm install --frozen-lockfile --fetch-timeout 60000"),
		bk.Cmd("pnpm --filter @sourcegraph/browser run build"),
		bk.Cmd("pnpm --filter @sourcegraph/browser release:npm"))
}
