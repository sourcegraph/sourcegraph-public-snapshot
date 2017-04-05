import { IWorkspace, IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { IThreadService } from "vs/workbench/services/thread/common/threadService";

import { context } from "sourcegraph/app/context";
import { registerContribution as registerExtHostContribution } from "sourcegraph/ext/extHost.contribution.override";
import { MainThreadService } from "sourcegraph/ext/mainThreadService";
import { InitializationOptions } from "sourcegraph/ext/protocol";
import { makeBlobURL } from "sourcegraph/init/worker";
import { listEnabled as listEnabledFeatures } from "sourcegraph/util/features";
import { fetchGQL } from "sourcegraph/util/gqlClient";
import { Services } from "sourcegraph/workbench/services";
import { getURIContext } from "sourcegraph/workbench/utils";

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
export function init(workspace: IWorkspace): void {
	registerExtHostContribution();
	setupWorker(workspace);
	(Services.get(IWorkspaceContextService)).onWorkspaceUpdated(w => setupWorker(w));
}

let seqId = 0;

export function setupWorker(workspace: IWorkspace): void {
	if (workspaces.has(workspace.resource.toString())) {
		return;
	}

	if (workspace.revState && workspace.revState.zapRef && !/^branch\//.test(workspace.revState.zapRef)) {
		throw new Error(`invalid Zap ref: ${JSON.stringify(workspace.revState.zapRef)} (no 'branch/' prefix)`);
	}

	workspaces.add(workspace.resource.toString());
	(Services.get(IWorkspaceContextService)).registerWorkspace(workspace);

	// Skip fetching contents if the resource only contains the scheme. (matches "file://")
	if (workspace.resource.toString() === "file://") {
		const opts: InitializationOptions = {
			seqId: ++seqId,
			workspace: workspace.toString(),
			features: listEnabledFeatures(),
			revState: undefined,
			context,
			langs: undefined,
		};
		const reader = new FileReader();
		const resultWithOpts = `self.extensionHostOptions=${JSON.stringify(opts)};${reader.result}`;
		const worker = new Worker(makeBlobURL(resultWithOpts));
		(Services.get(IThreadService) as MainThreadService).attachWorker(worker, workspace.resource);
		return;
	}
	const { repo } = getURIContext(workspace.resource);
	const rev = workspace.revState ? workspace.revState.commitID || "" : "";

	// Try to get repo language inventory before initializing extension host.
	// These langs are passed as initialization options to the extension host and
	// allow the host to short-circuit unnecessary LSP work for unused languages.
	// If there is an error fetching inventory, the extension host can make no
	// assumptions about which languages are used in the repository.
	// TODO(john): check that this works for private code.
	const getInventory = fetchGQL(`query getInventory($repo: String, $rev: String) {
		root {
			repository(uri: $repo) {
				commit(rev: $rev) {
					commit {
						languages
					}
				}
			}
		}
	}`, { repo, rev })
		.then(resp => {
			try {
				return resp.data.root.repository!.commit.commit!.languages;
			} catch (e) {
				console.warn(e);
				return undefined;
			}
		})
		.catch(err => undefined);

	getInventory.then(langs => {
		const opts: InitializationOptions = {
			seqId: ++seqId,
			workspace: workspace.resource.toString(),
			features: listEnabledFeatures(),
			revState: workspace.revState,
			context,
			langs,
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
			(Services.get(IThreadService) as MainThreadService).attachWorker(worker, workspace.resource);
		};
		reader.readAsText(blob);
	});
}
