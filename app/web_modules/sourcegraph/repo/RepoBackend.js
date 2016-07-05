// @flow weak

import * as RepoActions from "sourcegraph/repo/RepoActions";
import RepoStore from "sourcegraph/repo/RepoStore";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import {trackPromise} from "sourcegraph/app/status";
import {singleflightFetch} from "sourcegraph/util/singleflightFetch";
import {updateRepoCloning} from "sourcegraph/repo/cloning";
import {sortBranches, sortTags} from "sourcegraph/repo/vcs";
import EventLogger from "sourcegraph/util/EventLogger";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

export const Origin_GitHub = 0; // Origin.ServiceType enum value for GitHub origin

const RepoBackend = {
	fetch: singleflightFetch(defaultFetch),

	__onDispatch(action) {
		if (action instanceof RepoActions.WantRemoteRepos) {
			let repos;
			let queryString = "";
			if (action.opt) {
				repos = RepoStore.remoteRepos.getOpt(action.opt);
				queryString = `?IncludeDeps=${Boolean(action.opt.deps)}&Private=${Boolean(action.opt.private)}`;
			} else {
				repos = RepoStore.remoteRepos.list();
			}
			if (repos === null) {
				RepoBackend.fetch(`/.api/remote-repos${queryString}`)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => Dispatcher.Stores.dispatch(new RepoActions.RemoteReposFetched(action.opt, data)));
			}
			return;
		} else if (action instanceof RepoActions.WantCommit) {
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
			return;
		}

		switch (action.constructor) {

		case RepoActions.WantRepo:
			{
				let repo = RepoStore.repos.get(action.repo);
				if (repo === null) {
					trackPromise(
						RepoBackend.fetch(`/.api/repos/${action.repo}`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => {
								Dispatcher.Stores.dispatch(new RepoActions.FetchedRepo(action.repo, data));
							})
					);
				}
				break;
			}

		case RepoActions.WantResolveRepo:
			{
				let resolution = RepoStore.resolutions.get(action.repo);
				if (resolution === null) {
					trackPromise(
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
							})
					);
				}
				break;
			}

		case RepoActions.WantResolveRev:
			{
				let commitID = RepoStore.resolvedRevs.get(action.repo, action.rev);
				if (commitID === null || action.force) {
					const revPart = action.rev ? `@${action.rev}` : "";
					trackPromise(
						RepoBackend.fetch(`/.api/repos/${action.repo}${revPart}/-/rev`)
							.then(updateRepoCloning(action.repo))
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => {
								Dispatcher.Stores.dispatch(new RepoActions.ResolvedRev(action.repo, action.rev, data));
							})
					);
				}
				break;
			}

		case RepoActions.WantCreateRepo:
			{
				let body;
				if (action.remoteRepo.GitHubID) {
					body = {
						Op: {Origin: {ID: action.remoteRepo.GitHubID.toString(), Service: Origin_GitHub}},
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

				trackPromise(
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
						}
					})
				);
				break;
			}

		case RepoActions.WantBranches:
			{
				let branches = RepoStore.branches.list(action.repo);
				if (branches === null) {
					trackPromise(
						RepoBackend.fetch(`/.api/repos/${action.repo}/-/branches?IncludeCommit=true&PerPage=1000`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => {
								Dispatcher.Stores.dispatch(new RepoActions.FetchedBranches(action.repo, [], true));
							})
							.then((data) => Dispatcher.Stores.dispatch(new RepoActions.FetchedBranches(action.repo, sortBranches(data.Branches) || [])))
					);
				}
				break;
			}

		case RepoActions.WantTags:
			{
				let tags = RepoStore.tags.list(action.repo);
				if (tags === null) {
					trackPromise(
						RepoBackend.fetch(`/.api/repos/${action.repo}/-/tags?PerPage=1000`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => {
								Dispatcher.Stores.dispatch(new RepoActions.FetchedTags(action.repo, [], true));
							})
							.then((data) => Dispatcher.Stores.dispatch(new RepoActions.FetchedTags(action.repo, sortTags(data.Tags) || [])))
					);
				}
				break;
			}

		case RepoActions.WantInventory:
			{
				let inventory = RepoStore.inventory.get(action.repo, action.commitID);
				if (inventory === null) {
					trackPromise(
						RepoBackend.fetch(`/.api/repos/${action.repo}@${action.commitID}/-/inventory`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => Dispatcher.Stores.dispatch(new RepoActions.FetchedInventory(action.repo, action.commitID, data)))
					);
				}
				break;
			}

		case RepoActions.RefreshVCS:
			{
				trackPromise(
					RepoBackend.fetch(`/.api/repos/${action.repo}/-/refresh`, {method: "POST"})
						.then(checkStatus)
				);
				break;
			}
		}
	},
};

Dispatcher.Backends.register(RepoBackend.__onDispatch);

export default RepoBackend;
