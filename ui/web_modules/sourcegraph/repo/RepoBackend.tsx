import * as Dispatcher from "sourcegraph/Dispatcher";
import {updateRepoCloning} from "sourcegraph/repo/cloning";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import {RepoStore} from "sourcegraph/repo/RepoStore";
import {sortBranches, sortTags} from "sourcegraph/repo/vcs";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";
import {singleflightFetch} from "sourcegraph/util/singleflightFetch";
import {checkStatus, defaultFetch} from "sourcegraph/util/xhr";

export const OriginGitHub = 0; // Origin.ServiceType enum value for GitHub origin

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
					.catch((err) => ({Error: err}))
					.then((data) => {
						Dispatcher.Stores.dispatch(new RepoActions.ReposFetched(action.querystring, data));
					});
			}
		}

		if (payload instanceof RepoActions.WantCommit) {
			const action = payload;
			let commit = RepoStore.commits.get(action.repo, action.rev);
			if (commit === null) {
				RepoBackend.fetch(`/.api/repos/${action.repo}${action.rev ? `@${action.rev}` : ""}/-/commit`)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => {
						Dispatcher.Stores.dispatch(new RepoActions.FetchedCommit(action.repo, action.rev, data));
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
					.catch((err) => ({Error: err}))
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
					.catch((err) => ({Error: err}))
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
					.catch((err) => ({Error: err}))
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
					Op: {Origin: {ID: action.remoteRepo.GitHubID.toString(), Service: OriginGitHub}},
				};
			} else if (action.remoteRepo.Origin) {
				body = {
					Op: {Origin: action.remoteRepo.Origin},
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
			.catch((err) => ({Error: err}))
			.then((data) => {
				Dispatcher.Stores.dispatch(new RepoActions.RepoCreated(action.repo, data));
				if (!data.Error) {
					const eventProps = {language: action.remoteRepo.Language, private: Boolean(action.remoteRepo.Private)};
					EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_REPOSITORY, AnalyticsConstants.ACTION_SUCCESS, "AddRepo", eventProps);
					EventLogger.logIntercomEvent("add-repo", eventProps);
					if (action.refreshVCS) {
						RepoBackend.fetch(`/.api/repos/${action.repo}/-/refresh`, {method: "POST"})
							.then(checkStatus);
					}
				}
			});
		}

		if (payload instanceof RepoActions.WantCreateRepoHook) {
			const action = payload;
			RepoBackend.fetch(`/.api/webhook/enable?uri=${action.repo}`)
			.then(checkStatus);
		}

		if (payload instanceof RepoActions.WantBranches) {
			const action = payload;
			let branches = RepoStore.branches.list(action.repo);
			if (branches === null) {
				RepoBackend.fetch(`/.api/repos/${action.repo}/-/branches?IncludeCommit=true&PerPage=1000`)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => {
						Dispatcher.Stores.dispatch(new RepoActions.FetchedBranches(action.repo, []));
					})
					.then((data) => Dispatcher.Stores.dispatch(new RepoActions.FetchedBranches(action.repo, sortBranches(data.Branches) || [])));
			}
		}

		if (payload instanceof RepoActions.WantTags) {
			const action = payload;
			let tags = RepoStore.tags.list(action.repo);
			if (tags === null) {
				RepoBackend.fetch(`/.api/repos/${action.repo}/-/tags?PerPage=1000`)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => {
						Dispatcher.Stores.dispatch(new RepoActions.FetchedTags(action.repo, []));
					})
					.then((data) => Dispatcher.Stores.dispatch(new RepoActions.FetchedTags(action.repo, sortTags(data.Tags) || [])));
			}
		}

		if (payload instanceof RepoActions.WantInventory) {
			const action = payload;
			let inventory = RepoStore.inventory.get(action.repo, action.commitID);
			if (inventory === null) {
				RepoBackend.fetch(`/.api/repos/${action.repo}@${action.commitID}/-/inventory`)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => Dispatcher.Stores.dispatch(new RepoActions.FetchedInventory(action.repo, action.commitID, data)));
			}
		}

		if (payload instanceof RepoActions.RefreshVCS) {
			const action = payload;
			RepoBackend.fetch(`/.api/repos/${action.repo}/-/refresh`, {method: "POST"})
				.then(checkStatus);
		}

		if (payload instanceof RepoActions.WantSymbols) {
			const action = payload;
			const repos = RepoStore.symbols.list(action.repo, action.rev, action.query);
			if (repos === null) {
				const url = `/.api/repos/${action.repo}${action.rev ? `@${action.rev}` : ""}/-/symbols${action.query ? `?Query=` + encodeURIComponent(action.query) : ""}`;
				RepoBackend.fetch(url)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => {
						Dispatcher.Stores.dispatch(new RepoActions.FetchedSymbols(action.repo, action.rev, action.query, []));
					})
					.then((data) => {
						Dispatcher.Stores.dispatch(
							new RepoActions.FetchedSymbols(action.repo, action.rev, action.query, data.Symbols)
						);
					});
			}
		}
	},
};

Dispatcher.Backends.register(RepoBackend.__onDispatch);
