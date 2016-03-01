import update from "react/lib/update";

import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

export class GitHubReposStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.reset();
	}

	reset() {
		this.onWaitlist = window.onWaitlist;
		this.reposByOrg = deepFreeze({
			repos: window.mirrorData && window.mirrorData.ReposByOrg ? window.mirrorData.ReposByOrg : {},
			get(org) {
				let orgRepos = this.repos[org];
				return [].concat.apply(orgRepos.PublicRepos || [], orgRepos.PrivateRepos || []);
			},
			getMirrored() {
				// TODO(rothfels): this is gross and should be cleaned up...but is necessary to show mirrored repos on the dashboard.
				// We should probably build the map from org => repo in this store and just have the server return a flat list.
				let allRepos = (Object.values(this.repos) || []).map(orgRepos => (orgRepos.PublicRepos || []).concat(orgRepos.PrivateRepos || []));
				allRepos = [].concat.apply([], allRepos);
				return allRepos.filter(repo => repo.ExistsLocally).map(repo => repo.Repo);
			},
		});

		// Store the state of which organizations mirrored repos can come from.
		if (!(window.mirrorData && window.mirrorData.ReposByOrg)) {
			this.orgs = {};
		} else {
			this.orgs = Object.keys(window.mirrorData.ReposByOrg);
		}

		this.showLoading = false; // Indicates if a request to the backend to add mirror repos is in progress
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DashboardActions.WantAddMirrorRepos:
			this.showLoading = true;
			break;

		case DashboardActions.MirrorReposAdded:
			this.reposByOrg = update(this.reposByOrg, {
				repos: {$set: action.mirrorData ? action.mirrorData.ReposByOrg : {}},
			});
			this.showLoading = false;
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new GitHubReposStore(Dispatcher);
