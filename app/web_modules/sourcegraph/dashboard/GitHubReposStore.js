import update from "react/lib/update";

import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

export class GitHubReposStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.onWaitlist = window.onWaitlist;
		this.remoteRepos = deepFreeze({
			repos: window.mirrorData && window.mirrorData.RemoteRepos ? window.mirrorData.RemoteRepos : [],
			get(org) {
				return this.repos.filter(repo => repo.Owner.Login === org);
			},
			getMirrored() {
				return this.repos.filter(repo => repo.ExistsLocally).map(repo => repo.Repo);
			},
			getAll() {
				// TODO(rothfels): this is gross and should be cleaned up...but is necessary to show mirrored repos on the dashboard.
				// We should probably build the map from org => repo in this store and just have the server return a flat list.
				let allRepos = (Object.values(this.repos) || []).map(orgRepos => (orgRepos.PublicRepos || []).concat(orgRepos.PrivateRepos || []));
				allRepos = [].concat.apply([], allRepos);
				return allRepos.map(repo => repo.Repo);
			},
		});

		// Store the state of which organizations mirrored repos can come from by finding unique orgs
		if (!(window.mirrorData && window.mirrorData.RemoteRepos)) {
			this.orgs = {};
		} else {
			let u = {};
			let a = [];
			for (let repo of this.remoteRepos.repos) {
				if (!u.hasOwnProperty(repo.Owner.Login)) {
					u[repo.Owner.Login] = 1;
					a.push(repo.Owner.Login);
				}
			}
			this.orgs = a;
		}

		this.showLoading = false; // Indicates if a request to the backend to add mirror repos is in progress
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DashboardActions.WantAddMirrorRepos:
			this.showLoading = true;
			break;

		case DashboardActions.MirrorReposAdded:
			this.remoteRepos = update(this.remoteRepos, {
				repos: {$set: action.mirrorData ? action.mirrorData.RemoteRepos : {}},
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
