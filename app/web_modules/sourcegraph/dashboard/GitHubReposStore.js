import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";

export class GitHubReposStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.reposByOrg = deepFreeze({
			repos: window.mirrorData ? window.mirrorData.ReposByOrg : {},
			get(org) {
				let orgRepos = this.repos[org];
				return [].concat.apply(orgRepos.PublicRepos || [], orgRepos.PrivateRepos || []);
			},
		});

		// Store the state of which organizations mirrored repos can come from.
		this.orgs = window.mirrorData ? Object.keys(window.mirrorData.ReposByOrg) : {};
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
