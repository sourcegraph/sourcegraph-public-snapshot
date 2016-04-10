// @flow weak

import * as RepoActions from "sourcegraph/repo/RepoActions";
import RepoStore from "sourcegraph/repo/RepoStore";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import {trackPromise} from "sourcegraph/app/status";
import {singleflightFetch} from "sourcegraph/util/singleflightFetch";

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
