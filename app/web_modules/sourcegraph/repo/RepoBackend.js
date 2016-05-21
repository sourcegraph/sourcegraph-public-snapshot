// @flow weak

import * as RepoActions from "sourcegraph/repo/RepoActions";
import RepoStore from "sourcegraph/repo/RepoStore";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import {trackPromise} from "sourcegraph/app/status";
import {singleflightFetch} from "sourcegraph/util/singleflightFetch";
import {updateRepoCloning} from "sourcegraph/repo/cloning";
import EventLogger from "sourcegraph/util/EventLogger";

const RepoBackend = {
	fetch: singleflightFetch(defaultFetch),

	__onDispatch(action) {
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
						RepoBackend.fetch(`/.api/repos/${action.repo}/-/resolve`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => {
								Dispatcher.Stores.dispatch(new RepoActions.RepoResolved(action.repo, data.Error ? data : data.Data));
								if (data.IncludedRepo) {
									// Optimistically included by httpapi.serveRepoResolve.
									Dispatcher.Stores.dispatch(new RepoActions.FetchedRepo(action.repo, data.IncludedRepo));
								}
							})
					);
				}
				break;
			}

		case RepoActions.WantResolveRev:
			{
				let commitID = RepoStore.resolvedRevs.get(action.repo, action.rev);
				if (commitID === null) {
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
				let body = {Op: {FromGitHubID: action.remoteRepo.GitHubID}};
				if (!action.remoteRepo.GitHubID) { // non-GitHub repositories
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
							EventLogger.logEvent("AddRepo", eventProps);
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
						RepoBackend.fetch(`/.api/repos/${action.repo}/-/branches`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => {
								Dispatcher.Stores.dispatch(new RepoActions.FetchedBranches(action.repo, [], true));
							})
							.then((data) => Dispatcher.Stores.dispatch(new RepoActions.FetchedBranches(action.repo, data.Branches || [])))
					);
				}
				break;
			}

		case RepoActions.WantTags:
			{
				let tags = RepoStore.tags.list(action.repo);
				if (tags === null) {
					trackPromise(
						RepoBackend.fetch(`/.api/repos/${action.repo}/-/tags`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => {
								Dispatcher.Stores.dispatch(new RepoActions.FetchedTags(action.repo, [], true));
							})
							.then((data) => Dispatcher.Stores.dispatch(new RepoActions.FetchedTags(action.repo, data.Tags || [])))
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
