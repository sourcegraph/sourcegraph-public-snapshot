import * as Dispatcher from "sourcegraph/Dispatcher";
import * as lsp from "sourcegraph/editor/lsp";
import { inventoryToSearchModes } from "sourcegraph/editor/modes";
import { updateRepoCloning } from "sourcegraph/repo/cloning";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import { RepoStore } from "sourcegraph/repo/RepoStore";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { EventLogger } from "sourcegraph/util/EventLogger";
import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";

export const OriginGitHub = 0; // Origin.ServiceType enum value for GitHub origin

const workspaceSymbolFlights = new Set<string>();

export const RepoBackend = {
	fetch: singleflightFetch(defaultFetch),

	__onDispatch(payload: RepoActions.Action): void {
		if (payload instanceof RepoActions.WantRepos) {
			const action = payload;
			const repos = RepoStore.repos.list(action.querystring);
			if (repos === null) {
				const url = `/.api/repos?${action.querystring}`;
				RepoBackend.fetch(url)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({ Error: err }))
					.then((data) => {
						Dispatcher.Stores.dispatch(new RepoActions.ReposFetched(action.querystring, data, action.isUserRepos || false));
					});
			}
		}

		if (payload instanceof RepoActions.WantRepo) {
			const action = payload;
			let repo = RepoStore.repos.get(action.repo);
			if (repo === null) {
				RepoBackend.fetch(`/.api/repos/${action.repo}`)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({ Error: err }))
					.then((data) => {
						Dispatcher.Stores.dispatch(new RepoActions.FetchedRepo(action.repo, data));
					});
			}
		}

		if (payload instanceof RepoActions.WantResolveRepo) {
			const action = payload;
			let resolution = RepoStore.resolutions.get(action.repo);
			if (resolution === null) {
				RepoBackend.fetch(`/.api/repos/${action.repo}/-/resolve?Remote=true`)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({ Error: err }))
					.then((data) => {
						if (data.IncludedRepo) {
							// Optimistically included by httpapi.serveRepoResolve.
							Dispatcher.Stores.dispatch(new RepoActions.FetchedRepo(action.repo, data.IncludedRepo));
						}
						Dispatcher.Stores.dispatch(new RepoActions.RepoResolved(action.repo, data.Error ? data : data.Data));
					});
			}
		}

		if (payload instanceof RepoActions.WantResolveRev) {
			const action = payload;
			let commitID = RepoStore.resolvedRevs.get(action.repo, action.rev);
			if (commitID === null || action.force) {
				const revPart = action.rev ? `@${action.rev}` : "";
				RepoBackend.fetch(`/.api/repos/${action.repo}${revPart}/-/rev`)
					.then(updateRepoCloning(action.repo))
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({ Error: err }))
					.then((data) => {
						Dispatcher.Stores.dispatch(new RepoActions.ResolvedRev(action.repo, action.rev, data));
					});
			}
		}

		if (payload instanceof RepoActions.WantCreateRepo) {
			const action = payload;
			let body;
			if (action.remoteRepo.GitHubID) {
				body = {
					Op: { Origin: { ID: action.remoteRepo.GitHubID.toString(), Service: OriginGitHub } },
				};
			} else if (action.remoteRepo.Origin) {
				body = {
					Op: { Origin: action.remoteRepo.Origin },
				};
			} else {
				// Non-GitHub repositories.
				body = {
					Op: {
						New: {
							URI: action.remoteRepo.HTTPCloneURL.replace("https://", ""),
							CloneURL: action.remoteRepo.HTTPCloneURL,
							DefaultBranch: "master",
							Mirror: true,
						},
					},
				};
			}

			RepoBackend.fetch(`/.api/repos`, {
				method: "POST",
				body: JSON.stringify(body),
			})
				.then(checkStatus)
				.then((resp) => resp.json())
				.catch((err) => ({ Error: err }))
				.then((data) => {
					Dispatcher.Stores.dispatch(new RepoActions.RepoCreated(action.repo, data));
					if (!data.Error) {
						const eventProps = { language: action.remoteRepo.Language, private: Boolean(action.remoteRepo.Private) };
						AnalyticsConstants.Events.Repository_Added.logEvent(eventProps);
						EventLogger.logIntercomEvent("add-repo", eventProps);
						if (action.refreshVCS) {
							RepoBackend.fetch(`/.api/repos/${action.repo}/-/refresh`, { method: "POST" })
								.then(checkStatus);
						}
					}
				});
		}

		if (payload instanceof RepoActions.WantInventory) {
			const action = payload;
			let inventory = RepoStore.inventory.get(action.repo, action.commitID);
			if (inventory === null) {
				RepoBackend.fetch(`/.api/repos/${action.repo}@${action.commitID}/-/inventory`)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({ Error: err }))
					.then((data) => Dispatcher.Stores.dispatch(new RepoActions.FetchedInventory(action.repo, action.commitID, data)));
			}
		}

		if (payload instanceof RepoActions.RefreshVCS) {
			const action = payload;
			RepoBackend.fetch(`/.api/repos/${action.repo}/-/refresh`, { method: "POST" })
				.then(checkStatus);
		}

		if (payload instanceof RepoActions.WantSymbols) {
			const action = payload;
			let symbols = RepoStore.symbols.list(payload.inventory, payload.repo, payload.rev, payload.query);
			if (symbols.results.length > 0) {
				return;
			}
			inventoryToSearchModes(action.inventory).forEach(mode => {
				const url = `git:\/\/${action.repo}?${action.rev}`;
				const flightKey = `${url}@${mode}?${action.query}`;
				if (workspaceSymbolFlights.has(flightKey)) {
					return;
				}
				workspaceSymbolFlights.add(flightKey);
				lsp.sendExt(url, mode, "workspace/symbol", { query: action.query, limit: 100 })
					.then((r) => {
						let result;
						if (r === null || !r.result || !r.result.length) {
							result = [];
						} else {
							result = r.result;
						}
						Dispatcher.Stores.dispatch(
							new RepoActions.FetchedSymbols(mode, action.repo, action.rev, action.query, result)
						);
						workspaceSymbolFlights.delete(flightKey);
					});
			});
		}
	},
};

Dispatcher.Backends.register(RepoBackend.__onDispatch);
