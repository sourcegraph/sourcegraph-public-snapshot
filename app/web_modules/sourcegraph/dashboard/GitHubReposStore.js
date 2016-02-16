import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

export class GitHubReposStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		// Annotate repos with IsPrivate flag.
		if (window.mirrorData) {
			Object.keys(window.mirrorData.ReposByOrg).forEach(org => {
				(window.mirrorData.ReposByOrg[org].PublicRepos || []).forEach(repo => {
					repo.Repo.IsPrivate = false;
				});
				(window.mirrorData.ReposByOrg[org].PrivateRepos || []).forEach(repo => {
					repo.Repo.IsPrivate = true;
				});
			});
			this.onWaitlist = window.onWaitlist;
		}
		this.reposByOrg = deepFreeze({
			repos: window.mirrorData ? window.mirrorData.ReposByOrg : {},
			get(org) {
				let orgRepos = this.repos[org];
				return [].concat.apply(orgRepos.PublicRepos || [], orgRepos.PrivateRepos || []);
			},
		});

		// Store the state of which organizations mirrored repos can come from.
		if (!(window.mirrorData && window.mirrorData.ReposByOrg)) {
			this.orgs = {};
		} else {
			this.orgs = Object.keys(window.mirrorData.ReposByOrg);
		}
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DashboardActions.MirrorReposAdded:
			// Set ExistsLocaly for the repos which have been added.
			break;


		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new GitHubReposStore(Dispatcher);
