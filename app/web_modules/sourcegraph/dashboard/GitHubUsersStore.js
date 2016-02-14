import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

import update from "react/lib/update";

export class GitHubUsersStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		// The users which are available to invite.
		this.mirrorUsers = deepFreeze(window.mirrorUsers);

		// Store the state of which users are currently selected (e.g. to add).
		// This is for widget components.
		this.selectAll = false;
		let selectedUsers = {};
		if (this.mirrorUsers) this.mirrorUsers.forEach(user => selectedUsers[user.index] = false);
		this.selectedUsers = deepFreeze(selectedUsers);

		// Store the state of which organizations mirrored repos can come from.
		// The currentOrg is a filter for widget components.
		let orgs = {};
		if (this.mirrorUsers) this.mirrorUsers.forEach(user => orgs[user.org] = 1);
		this.orgs = Object.keys(orgs);
		this.currentOrg = this.orgs.length > 0 ? this.orgs[0] : null;
	}

	__onDispatch(action) {
		switch (action.constructor) {

		case DashboardActions.UsersAdded:
			// TODO: remove them from this store?
			break;

		case DashboardActions.SelectRepoOrg:
			this.currentOrg = action.org;
			this.selectAll = false;
			break;

		case DashboardActions.SelectUsers:
			{
				console.log("received the (user) dispatch!");
				let updateQuery = {};
				action.repos.forEach(user => updateQuery[user.index] = {$set: action.selectAll});
				this.selectedUsers = update(this.selectedUsers, updateQuery);
				this.selectAll = action.selectAll;
				break;
			}

		case DashboardActions.SelectRepo:
			{
				console.log("selected a single user!", action.repoIndex, action.select);
				let updateQuery = {};
				updateQuery[action.repoIndex] = {$set: action.select};
				this.selectedUsers = update(this.selectedUsers, updateQuery);
				console.log(this.selectedUsers);
				break;
			}

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new GitHubUsersStore(Dispatcher);
