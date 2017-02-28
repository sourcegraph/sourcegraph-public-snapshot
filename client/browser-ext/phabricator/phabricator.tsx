/**
 * This is done before all other imports to ensure that the event logger is set ahead of time.
 */
import { InPageEventLogger } from "../app/tracking/InPageEventLogger";
import { setEventLogger, setSourcegraphUrl } from "../app/utils/context";
setEventLogger(new InPageEventLogger("SourcegraphExtension", "PhabricatorExtension"));
setSourcegraphUrl("http://node.aws.sgdev.org:30000");

import { getDomainUsername } from "../app/utils";
import { eventLogger } from "../app/utils/context";
import { injectBackgroundApp } from "../app/utils/injectBackgroundApp";
import { expanderListen, getPhabricatorUsername, metaClickOverride, setupPageLoadListener } from "../app/utils/phabricator";
import { injectPhabricatorBlobAnnotators } from "../app/utils/phabricator_inject";

// fragile and not great
const phabricatorUsername = getPhabricatorUsername();
if (phabricatorUsername !== null) {
	(eventLogger as InPageEventLogger).setUserName(getDomainUsername("sgdev_phabricator", phabricatorUsername));
}

/**
 * This is the main entry point for the phabricator in-page JavaScript plugin.
 */
if (global && global.window && global.window.localStorage && !(global.window.localStorage.SOURCEGRAPH_DISABLED === "true")) {
	document.addEventListener("phabPageLoaded", ev => {
		expanderListen();
		metaClickOverride();
		injectModules();
		setTimeout(injectModules, 1000); // extra data may be loaded asynchronously; reapply after timeout
		setTimeout(injectModules, 5000); // extra data may be loaded asynchronously; reapply after timeout
	});
	setupPageLoadListener();
} else {
	// tslint:disable-next-line
	console.log(`Sourcegraph on Phabricator is disabled because window.localStorage.SOURCEGRAPH_DISABLED is set to ${global.window.localStorage.SOURCEGRAPH_DISABLED}.`);
}

// NOTE: injectModules is idempotent, so safe to call multiple times on the same page.
function injectModules(): void {
	// TODO(uforic): We probably don't need to do this for Phabricator, since we don't make use of it.
	injectBackgroundApp(null);
	injectPhabricatorBlobAnnotators();
}
