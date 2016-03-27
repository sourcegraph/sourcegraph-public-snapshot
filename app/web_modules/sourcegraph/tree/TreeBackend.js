import * as TreeActions from "sourcegraph/tree/TreeActions";
import TreeStore from "sourcegraph/tree/TreeStore";
import Dispatcher from "sourcegraph/Dispatcher";
import defaultXhr from "sourcegraph/util/xhr";

const TreeBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
		case TreeActions.WantCommit:
			{
				let commit = TreeStore.commits.get(action.repo, action.rev, action.path);
				if (commit === null) {
					TreeBackend.xhr({
						uri: `/.api/repos/${action.repo}/.commits?Head=${encodeURIComponent(action.rev)}&Path=${encodeURIComponent(action.path)}&PerPage=1`,
						json: {},
					}, function(err, resp, body) {
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.dispatch(new TreeActions.CommitFetched(action.repo, action.rev, action.path, body));
					});
				}
				break;
			}

		case TreeActions.WantFileList:
			{
				let fileList = TreeStore.fileLists.get(action.repo, action.rev, action.commitID);
				if (fileList === null) {
					TreeBackend.xhr({
						uri: `/.api/repos/${action.repo}@${encodeURIComponent(action.rev)}===${encodeURIComponent(action.commitID)}/.tree-list`,
						json: {},
					}, function(err, resp, body) {
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.dispatch(new TreeActions.FileListFetched(action.repo, action.rev, action.commitID, body));
					});
				}
				break;
			}

		case TreeActions.WantSrclibDataVersion:
			{
				let version = TreeStore.srclibDataVersions.get(action.repo, action.rev, action.commitID, action.path);
				if (version === null) {
					TreeBackend.xhr({
						uri: `/.api/repos/${action.repo}@${encodeURIComponent(action.rev)}===${encodeURIComponent(action.commitID)}/.srclib-data-version?Path=${action.path ? encodeURIComponent(action.path) : ""}`,
						json: {},
					}, function(err, resp, body) {
						if (resp.statusCode === 404) {
							body = {};
							err = null;
						}
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.dispatch(new TreeActions.FetchedSrclibDataVersion(action.repo, action.rev, action.commitID, action.path, body));
					});
				}
				break;
			}
		}
	},
};

Dispatcher.register(TreeBackend.__onDispatch);

export default TreeBackend;
