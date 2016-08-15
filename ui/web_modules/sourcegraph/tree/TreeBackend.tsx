import * as Dispatcher from "sourcegraph/Dispatcher";
import {updateRepoCloning} from "sourcegraph/repo/cloning";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import {TreeStore} from "sourcegraph/tree/TreeStore";
import {checkStatus, defaultFetch} from "sourcegraph/util/xhr";

export const TreeBackend = {
	fetch: defaultFetch as any,

	__onDispatch(payload: TreeActions.Action): void {
		if (payload instanceof TreeActions.WantCommit) {
			const action = payload;
			let commit = TreeStore.commits.get(action.repo, action.rev, action.path);
			if (commit === null) {
				TreeBackend.fetch(`/.api/repos/${action.repo}/-/commits?Head=${encodeURIComponent(action.rev)}&Path=${encodeURIComponent(action.path)}&PerPage=1`)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => Dispatcher.Stores.dispatch(new TreeActions.CommitFetched(action.repo, action.rev, action.path, data.Commits[0])));
			}
		}

		if (payload instanceof TreeActions.WantFileList) {
			const action = payload;
			let fileList = TreeStore.fileLists.get(action.repo, action.commitID);
			if (fileList === null) {
				TreeBackend.fetch(`/.api/repos/${action.repo}@${action.commitID}/-/tree-list`)
					.then(updateRepoCloning(action.repo))
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => Dispatcher.Stores.dispatch(new TreeActions.FileListFetched(action.repo, action.commitID, data)));
			}
		}

		if (payload instanceof TreeActions.WantSrclibDataVersion) {
			const action = payload;
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
		}
	},
};

Dispatcher.Backends.register(TreeBackend.__onDispatch);
