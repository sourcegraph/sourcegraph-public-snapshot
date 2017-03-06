/**
 * set the event logger before anything else proceeds, to avoid logging events before we have it set.
 */
import { ExtensionEventLogger } from "../../app/tracking/ExtensionEventLogger";
import { eventLogger, setEventLogger, setPhabricatorInstance, setSourcegraphUrl } from "../../app/utils/context";
setEventLogger(new ExtensionEventLogger());

import { getDomain } from "../../app/utils";
import { injectBackgroundApp } from "../../app/utils/injectBackgroundApp";
import { Domain } from "../../app/utils/types";
import { injectGitHubApplication } from "./inject_github";
import { injectPhabricatorApplication } from "./inject_phabricator";

import { SGDEV_SOURCEGRAPH_URL, sgDevPhabricatorInstance } from "../../phabricator/sgdev/constants";

/**
 * Main entry point into browser extension.
 *
 * Depending on the domain, we load one of three different applications.
 */
function injectApplication(loc: Location): void {
	switch (getDomain(loc)) {
		case Domain.GITHUB:
			setSourcegraphUrl("https://sourcegraph.com");
			injectGitHubApplication();
			break;
		case Domain.SGDEV_PHABRICATOR:
			setSourcegraphUrl(SGDEV_SOURCEGRAPH_URL);
			setPhabricatorInstance(sgDevPhabricatorInstance);
			injectPhabricatorApplication();
			break;
		case Domain.SOURCEGRAPH:
			setSourcegraphUrl("https://sourcegraph.com");
			injectSourcergaphCloudApplication();
			break;
		default:
			break;
	}
}

function injectSourcergaphCloudApplication(): void {
	injectBackgroundApp(null);
	document.addEventListener("sourcegraph:identify", (ev: CustomEvent) => {
		if (ev && ev.detail) {
			(eventLogger as ExtensionEventLogger).updatePropsForUser(ev.detail);
			chrome.runtime.sendMessage({ type: "setIdentity", identity: ev.detail });
		} else {
			console.error("sourcegraph:identify missing details");
		}
	});
}

injectApplication(window.location);
