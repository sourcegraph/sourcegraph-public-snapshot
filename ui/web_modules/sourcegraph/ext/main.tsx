import URI from "vs/base/common/uri";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { IThreadService } from "vs/workbench/services/thread/common/threadService";

import { context } from "sourcegraph/app/context";
import { registerContribution as registerExtHostContribution } from "sourcegraph/ext/extHost.contribution.override";
import { MainThreadService } from "sourcegraph/ext/mainThreadService";
import { InitializationOptions } from "sourcegraph/ext/protocol";
import { makeBlobURL } from "sourcegraph/init/worker";
import { listEnabled as listEnabledFeatures } from "sourcegraph/util/features";
import { Services } from "sourcegraph/workbench/services";

/**
 * workspaces is a set of URIs for which we have already bootstrapped
 * an extension host.
 */
const workspaces = new Set<string>();

/**
 * init initializes an extension host for the initial workspace and
 * sets up a listener to initialize a new extension host when workspace
 * is updated.
 *
 * TODO(john): there is currently no cleanup of unused extension hosts / web workers.
 */
export function init(workspace: URI, zapRef?: string): void {
	registerExtHostContribution();
	setupWorker(workspace, zapRef);
	(Services.get(IWorkspaceContextService)).onWorkspaceUpdated(w => setupWorker(w.resource, w.revState ? w.revState.zapRef : undefined));
}

let seqId = 0;

export function setupWorker(workspace: URI, zapRef?: string): void {
	if (workspaces.has(workspace.toString())) {
		return;
	}

	seqId += 1;
	workspaces.add(workspace.toString());

	const opts: InitializationOptions = {
		seqId,
		workspace: workspace.toString(),
		features: listEnabledFeatures(),
		zapRef: zapRef,
		context,
	};

	const blob = new Blob([require("raw-loader!inline-worker-loader!sourcegraph/ext/host")]);
	const reader = new FileReader();

	// We need to provide properties (e.g. current workspace) to the worker.
	// We could add those properties to the worker location's fragment; however,
	// there are browser compatibility issues with loading a blob URI with a
	// fragment (fails on Safari). Instead, we read the blob contents and append
	// properties to the worker script's global namespace. This is safe to do
	// because the worker's namespace is sandboxed.
	reader.onload = event => {
		const resultWithOpts = `self.extensionHostOptions=${JSON.stringify(opts)};${reader.result}`;
		const worker = new Worker(makeBlobURL(resultWithOpts));
		(Services.get(IThreadService) as MainThreadService).attachWorker(worker, workspace);
	};
	reader.readAsText(blob);
}
