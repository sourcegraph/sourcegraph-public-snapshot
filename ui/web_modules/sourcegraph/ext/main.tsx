import URI from "vs/base/common/uri";
import { IThreadService } from "vs/workbench/services/thread/common/threadService";

import { registerContribution as registerExtHostContribution } from "sourcegraph/ext/extHost.contribution.override";
import { MainThreadService } from "sourcegraph/ext/mainThreadService";
import { InitializationOptions } from "sourcegraph/ext/protocol";
import { makeBlobURL } from "sourcegraph/init/worker";
import { listEnabled as listEnabledFeatures } from "sourcegraph/util/features";
import { Services } from "sourcegraph/workbench/services";

let inited = false;

/**
 * init initializes the extension host. If it is already initialized,
 * it is a noop. There is only one extension host and it hosts all
 * extensions.
 *
 * TODO(sqs): This means that only the first workspace loaded upon page
 * load can be used in the extension host. This is a known limitation
 * and will be addressed before the vscode extensions API support is
 * un-feature-flagged.
 */
export function init(workspace: URI): void {
	if (inited) {
		console.log("DEV NOTE: The vscode extension API support only works on the first repository/commit that you visit after loading the page. This limitation will be addressed before release."); // tslint:disable-line no-console
		return;
	}
	inited = true;

	registerExtHostContribution();

	const opts: InitializationOptions = {
		workspace: workspace.toString(),
		features: listEnabledFeatures(),
	};

	// Add our current URL in the worker location's fragment so the
	// worker can determine the current workspace.
	//
	// TODO(sqs): This limits the extension API to working only on the workspace
	// that was in use when the page loaded. This will be fixed before we release
	// the vscode extension API support.
	const worker = new Worker(makeBlobURL(require("raw-loader!inline-worker-loader!sourcegraph/ext/host")) + `#${encodeURIComponent(JSON.stringify(opts))}`);
	(Services.get(IThreadService) as MainThreadService).attachWorker(worker);
}
