// tslint:disable: typedef ordered-imports

import * as TreeActions from "sourcegraph/tree/TreeActions";
import {TreeStore} from "sourcegraph/tree/TreeStore";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import {updateRepoCloning} from "sourcegraph/repo/cloning";

export const TreeBackend = {
	fetch: defaultFetch as any,

	__onDispatch(action) {
		switch (action.constructor) {
		case TreeActions.WantCommit:
			{
				let commit = TreeStore.commits.get(action.repo, action.rev, action.path);
				if (commit === null) {
					TreeBackend.fetch(`/.api/repos/${action.repo}/-/commits?Head=${encodeURIComponent(action.rev)}&Path=${encodeURIComponent(action.path)}&PerPage=1`)
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => ({Error: err}))
						.then((data) => Dispatcher.Stores.dispatch(new TreeActions.CommitFetched(action.repo, action.rev, action.path, data.Commits[0])));
				}
				break;
			}

		case TreeActions.WantFileList:
			{
				let fileList = TreeStore.fileLists.get(action.repo, action.commitID);
				if (fileList === null) {
					TreeBackend.fetch(`/.api/repos/${action.repo}@${action.commitID}/-/tree-list`)
						.then(updateRepoCloning(action.repo))
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => ({Error: err}))
						.then((data) => Dispatcher.Stores.dispatch(new TreeActions.FileListFetched(action.repo, action.commitID, data)));
				}
				break;
			}

		case TreeActions.WantSrclibDataVersion:
			{
				let version = TreeStore.srclibDataVersions.get(action.repo, action.commitID, action.path);
				if (version === null || action.force) {
					TreeBackend.fetch(`/.api/repos/${action.repo}@${action.commitID}/-/srclib-data-version?Path=${action.path ? encodeURIComponent(action.path) : ""}`)
						.then((resp) => {
							if (resp.status === 404) {
								return Object.assign({}, resp, {json: () => ({})});
							} else if (resp.status === 200) {
								return resp;
							}
							return checkStatus(resp);
						})
						.then((resp) => resp.json())
						.catch((err) => ({Error: err}))
						.then((data) => Dispatcher.Stores.dispatch(new TreeActions.FetchedSrclibDataVersion(action.repo, action.commitID, action.path, data)));
				}
				break;
			}
		}
	},
};

Dispatcher.Backends.register(TreeBackend.__onDispatch);
