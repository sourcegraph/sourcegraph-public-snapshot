import * as phabricator from "../../app/utils/phabricator";
import { injectPhabricatorBlobAnnotators } from "../../app/utils/phabricator_inject";

export function injectPhabricatorApplication(): void {
	// make sure this is called before javelinPierce
	document.addEventListener(phabricator.PHAB_PAGE_LOAD_EVENT_NAME, ev => {
		injectModules();
		setTimeout(injectModules, 5000); // extra data may be loaded asynchronously; reapply after timeout
	});
	phabricator.javelinPierce(phabricator.setupPageLoadListener, "body");
	phabricator.javelinPierce(phabricator.expanderListen, "body");
	phabricator.javelinPierce(phabricator.metaClickOverride, "body");
}

function injectModules(): void {
	injectPhabricatorBlobAnnotators();
}
