import URI from "vs/base/common/uri";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { IThreadService } from "vs/workbench/services/thread/common/threadService";

import { context } from "sourcegraph/app/context";
import { URIUtils } from "sourcegraph/core/uri";
import { registerContribution as registerExtHostContribution } from "sourcegraph/ext/extHost.contribution.override";
import { MainThreadService } from "sourcegraph/ext/mainThreadService";
import { InitializationOptions } from "sourcegraph/ext/protocol";
import { makeBlobURL } from "sourcegraph/init/worker";
import { listEnabled as listEnabledFeatures } from "sourcegraph/util/features";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";
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
export function init(workspace: URI, revState?: { zapRev?: string, zapRef?: string, commitID?: string, branch?: string }): void {
	registerExtHostContribution();
	setupWorker(workspace, revState);
	(Services.get(IWorkspaceContextService)).onWorkspaceUpdated(w => setupWorker(w.resource, w.revState ? w.revState : undefined));
}

let seqId = 0;

export function setupWorker(workspace: URI, revState?: { zapRev?: string, zapRef?: string, commitID?: string, branch?: string }): void {
	if (workspaces.has(workspace.toString())) {
		return;
	}

	if (revState && revState.zapRef && !/^branch\//.test(revState.zapRef)) {
		throw new Error(`invalid Zap ref: ${JSON.stringify(revState.zapRef)} (no 'branch/' prefix)`);
	}

	workspaces.add(workspace.toString());

	const { repo, rev } = URIUtils.repoParams(workspace);

	// Try to get repo language inventory before initializing extension host.
	// These langs are passed as initialization options to the extension host and
	// allow the host to short-circuit unnecessary LSP work for unused languages.
	// If there is an error fetching inventory, the extension host can make no
	// assumptions about which languages are used in the repository.
	const getInventory = fetchGraphQLQuery(`query RepoInventory($repo: String, $rev: String) {
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
		.then(query => {
			try {
				return query.root.repository!.commit.commit!.languages;
			} catch (e) {
				console.warn(e);
				return undefined;
			}
		})
		.catch(err => undefined);

	getInventory.then(langs => {
		const opts: InitializationOptions = {
			seqId: ++seqId,
			workspace: workspace.toString(),
			features: listEnabledFeatures(),
			revState,
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
			(Services.get(IThreadService) as MainThreadService).attachWorker(worker, workspace);
		};
		reader.readAsText(blob);
	});
}
