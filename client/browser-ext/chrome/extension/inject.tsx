import { Domain, getDomain } from "../../app/utils";
import { injectBackgroundApp, injectGitHubApplication } from "./inject_github";

/**
 * Main entry point into browser extension.
 *
 * Depending on the domain, we load one of three different applications.
 */
switch (getDomain(window.location)) {
	case Domain.GITHUB:
		injectGitHubApplication();
		break;
	case Domain.SGDEV_PHABRICATOR:
		break;
	case Domain.SOURCEGRAPH:
		injectSourcegraphBackgroundApp();
		break;
	default:
		break;
}

function injectSourcegraphBackgroundApp(): void {
	injectBackgroundApp(null);
}
