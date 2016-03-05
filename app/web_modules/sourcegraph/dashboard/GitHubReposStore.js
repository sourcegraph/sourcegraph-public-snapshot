import update from "react/lib/update";

import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";

export class GitHubReposStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.reset();
	}

	reset() {
		if (typeof window !== "undefined") { // TODO(autotest) support document object.
			this.onWaitlist = window.onWaitlist;
			this.repos = deepFreeze(
				window.mirrorData && window.mirrorData.Repos ? window.mirrorData.Repos : []
			);
			this.remoteRepos = deepFreeze(
				window.mirrorData && window.mirrorData.RemoteRepos ? window.mirrorData.RemoteRepos : []
			);
		} else {
			this.onWaitlist = false;
			this.repos = deepFreeze([]);
			this.remoteRepos = deepFreeze([]);
		}
	}

	__onDispatch(action) {
		switch (action.constructor) {
		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new GitHubReposStore(Dispatcher);
