import * as TreeActions from "sourcegraph/tree/TreeActions";
import TreeStore from "sourcegraph/tree/TreeStore";
import Dispatcher from "sourcegraph/Dispatcher";
import defaultXhr from "xhr";

const TreeBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
		case TreeActions.WantCommit:
			{
				let commit = TreeStore.commits.get(action.repo, action.rev, action.path);
				if (commit === null) {
					TreeBackend.xhr({
						uri: `/.ui/${action.repo}/.commits?Head=${encodeURIComponent(action.rev)}&Path=${encodeURIComponent(action.path)}&PerPage=1`,
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
		}
	},
};

Dispatcher.register(TreeBackend.__onDispatch);

export default TreeBackend;
