// @flow weak

import * as RepoActions from "sourcegraph/repo/RepoActions";
import RepoStore from "sourcegraph/repo/RepoStore";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";

const RepoBackend = {
	fetch: defaultFetch,

	__onDispatch(action) {
		switch (action.constructor) {

		case RepoActions.WantRepo:
			{
				let repo = RepoStore.repos.get(action.repo);
				if (repo === null) {
					RepoBackend.fetch(`/.api/repos/${action.repo}`)
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => {
							console.error(err);
							// TODO Better httpapi error responses.
							return {Error: {Body: err.body, Status: err.response.status}};
						})
						.then((data) => {
							Dispatcher.Stores.dispatch(new RepoActions.FetchedRepo(action.repo, data));
						});
				}
				break;
			}

		case RepoActions.WantBranches:
			{
				let branches = RepoStore.branches.list(action.repo);
				if (branches === null) {
					RepoBackend.fetch(`/.api/repos/${action.repo}/-/branches`)
							.then((resp) => resp.json())
							.catch((err) => {
								Dispatcher.Stores.dispatch(new RepoActions.FetchedBranches(action.repo, [], true));
								console.error(err);
							})
							.then((data) => Dispatcher.Stores.dispatch(new RepoActions.FetchedBranches(action.repo, data.Branches || [])));
				}
				break;
			}

		case RepoActions.WantTags:
			{
				let tags = RepoStore.tags.list(action.repo);
				if (tags === null) {
					RepoBackend.fetch(`/.api/repos/${action.repo}/-/tags`)
							.then((resp) => resp.json())
							.catch((err) => {
								Dispatcher.Stores.dispatch(new RepoActions.FetchedTags(action.repo, [], true));
								console.error(err);
							})
							.then((data) => Dispatcher.Stores.dispatch(new RepoActions.FetchedTags(action.repo, data.Tags || [])));
				}
				break;
			}

		case RepoActions.RefreshVCS:
			{
				RepoBackend.fetch(`/.api/repos/${action.repo}/-/refresh`, {method: "POST"})
					.then(checkStatus)
					.catch((err) => console.error(err));
				break;
			}
		}
	},
};

Dispatcher.Backends.register(RepoBackend.__onDispatch);

export default RepoBackend;
