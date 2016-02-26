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
			getDashboard() {
				return this.repos.map(repo => update(repo.Repo, {$merge: {ExistsLocally: repo.ExistsLocally}}));
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
		case DashboardActions.MirrorRepoAdded:
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
