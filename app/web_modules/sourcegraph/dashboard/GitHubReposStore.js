import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

import update from "react/lib/update";

export class GitHubReposStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		// The repos which are available to mirror.
		this.mirrorRepos = deepFreeze(window.mirrorRepos);

		// Store the state of which repos are currently selected (e.g. to mirror).
		// This is for widget components.
		this.selectAll = false;
		let selectedRepos = {};
		if (this.mirrorRepos) this.mirrorRepos.forEach(repo => selectedRepos[repo.key] = false);
		this.selectedRepos = deepFreeze(selectedRepos);

		// Store the state of which organizations mirrored repos can come from.
		// The currentOrg is a filter for widget components.
		let orgs = {};
		if (this.mirrorRepos) this.mirrorRepos.forEach(repo => orgs[repo.org] = 1);
		this.orgs = Object.keys(orgs);
		this.currentOrg = this.orgs.length > 0 ? this.orgs[0] : null;
	}

	__onDispatch(action) {
		switch (action.constructor) {

		case DashboardActions.ReposAdded:
			// TODO: remove them from this store?
			break;

		case DashboardActions.SelectRepoOrg:
			this.currentOrg = action.org;
			this.selectAll = false;
			break;

		case DashboardActions.SelectRepos:
			{
				let updateQuery = {};
				action.repos.forEach(repo => updateQuery[repo.key] = {$set: action.selectAll});
				this.selectedRepos = update(this.selectedRepos, updateQuery);
				this.selectAll = action.selectAll;
				break;
			}

		case DashboardActions.SelectRepo:
			{
				let updateQuery = {};
				updateQuery[action.repoKey] = {$set: action.select};
				this.selectedRepos = update(this.selectedRepos, updateQuery);
				break;
			}

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new GitHubReposStore(Dispatcher);
