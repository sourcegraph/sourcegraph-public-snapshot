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
		if (typeof window !== "undefined") { // TODO(autotest) support document object.
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
		} else {
			this.onWaitlist = false;
			this.remoteRepos = deepFreeze({
				repos: [],
				getDashboard() { return []; },
			});
		}

		// Store the state of which organizations mirrored repos can come from by finding unique orgs
		if (typeof window === "undefined" || !(window.mirrorData && window.mirrorData.RemoteRepos)) {
			this.orgs = {};
			this.reposByOrg = {};
		} else {
			let u = {};
			this.remoteRepos.repos.forEach(repo => u[repo.Owner.Login] = 1);
			this.orgs = Object.keys(u);
			this.reposByOrg = {};
			this.remoteRepos.repos.map(repo => this.reposByOrg.hasOwnProperty(repo.Owner.Login) ? this.reposByOrg[repo.Owner.Login].push(repo) : this.reposByOrg[repo.Owner.Login] = [].concat.apply(repo || []));
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
