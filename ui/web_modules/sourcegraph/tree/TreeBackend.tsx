import * as Dispatcher from "sourcegraph/Dispatcher";
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
	},
};

Dispatcher.Backends.register(TreeBackend.__onDispatch);
