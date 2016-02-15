import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

import update from "react/lib/update";

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
		// The currentOrg is a filter for widget components.
		this.orgs = window.mirrorData ? Object.keys(window.mirrorData.ReposByOrg) : {};
		this.currentOrg = this.orgs.length > 0 ? this.orgs[0] : null;

		// Store the state of which repos are currently selected (e.g. to mirror).
		// This is for widget components.
		this.selectAll = false;
		this.selectedRepos = deepFreeze({});
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
				action.repos.forEach(repoURI => updateQuery[repoURI] = {$set: action.selectAll});
				this.selectedRepos = update(this.selectedRepos, updateQuery);
				this.selectAll = action.selectAll;
				break;
			}

		case DashboardActions.SelectRepo:
			{
				let updateQuery = {};
				updateQuery[action.repoURI] = {$set: action.select};
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
