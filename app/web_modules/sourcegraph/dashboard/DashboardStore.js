import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";

import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

export class DashboardStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
	}

	toJSON() {
		return {
			repos: this.repos,
			remoteRepos: this.remoteRepos,
		};
	}

	reset(data) {
		this.repos = data && data.repos ? data.repos : null;
		this.remoteRepos = data && data.remoteRepos ? data.remoteRepos : null;
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DashboardActions.ReposFetched:
			this.repos = deepFreeze((action.data && action.data.Repos) || []);
			break;

		case DashboardActions.RemoteReposFetched:
			this.remoteRepos = deepFreeze((action.data && action.data.RemoteRepos) || []);
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new DashboardStore(Dispatcher.Stores);
