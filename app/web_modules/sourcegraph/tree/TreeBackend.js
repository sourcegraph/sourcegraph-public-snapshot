import * as TreeActions from "sourcegraph/tree/TreeActions";
import TreeStore from "sourcegraph/tree/TreeStore";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";

const TreeBackend = {
	fetch: defaultFetch,

	__onDispatch(action) {
		switch (action.constructor) {
		case TreeActions.WantCommit:
			{
				let commit = TreeStore.commits.get(action.repo, action.rev, action.path);
				if (commit === null) {
					TreeBackend.fetch(`/.api/repos/${action.repo}/-/commits?Head=${encodeURIComponent(action.rev)}&Path=${encodeURIComponent(action.path)}&PerPage=1`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: true}))
							.then((data) => Dispatcher.Stores.dispatch(new TreeActions.CommitFetched(action.repo, action.rev, action.path, data.Commits[0])));
				}
				break;
			}

		case TreeActions.WantFileList:
			{
				let fileList = TreeStore.fileLists.get(action.repo, action.rev);
				if (fileList === null) {
					TreeBackend.fetch(`/.api/repos/${action.repo}@${action.rev}/-/tree-list`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: true}))
							.then((data) => Dispatcher.Stores.dispatch(new TreeActions.FileListFetched(action.repo, action.rev, data)));
				}
				break;
			}

		case TreeActions.WantSrclibDataVersion:
			{
				let version = TreeStore.srclibDataVersions.get(action.repo, action.rev, action.path);
				if (version === null) {
					TreeBackend.fetch(`/.api/repos/${action.repo}@${action.rev}/-/srclib-data-version?Path=${action.path ? encodeURIComponent(action.path) : ""}`)
							.then((resp) => {
								if (resp.status === 404) {
									return Object.assign({}, resp, {json: () => ({})});
								} else if (resp.status === 200) {
									return resp;
								}
								return checkStatus(resp);
							})
							.then((resp) => resp.json())
							.catch((err) => ({Error: true}))
							.then((data) => Dispatcher.Stores.dispatch(new TreeActions.FetchedSrclibDataVersion(action.repo, action.rev, action.path, data)));
				}
				break;
			}
		}
	},
};

Dispatcher.Backends.register(TreeBackend.__onDispatch);

export default TreeBackend;
