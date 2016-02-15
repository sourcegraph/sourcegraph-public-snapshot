import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

import update from "react/lib/update";

export class GitHubUsersStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		// The users which are available to invite.
		this.usersByOrg = deepFreeze({
			users: window.teammates ? window.teammates.UsersByOrg : {},
			get(org) {
				let orgUsers = this.users[org];
				return orgUsers ? orgUsers.Users : [];
			},
		});

		// Store the state of which organizations mirrored users can come from.
		// The currentOrg is a filter for widget components.
		this.orgs = window.teammates ? Object.keys(window.teammates.UsersByOrg) : {};
		this.currentOrg = this.orgs.length > 0 ? this.orgs[0] : null;

		// Store the state of which users are currently selected (e.g. to add).
		// This is for widget components.
		this.selectAll = false;
		this.selectedUsers = deepFreeze({});
	}

	__onDispatch(action) {
		switch (action.constructor) {

		case DashboardActions.UsersAdded:
			// TODO: remove them from this store?
			break;

		case DashboardActions.SelectUserOrg:
			this.currentOrg = action.org;
			this.selectAll = false;
			break;

		case DashboardActions.SelectUsers:
			{
				let updateQuery = {};
				action.users.forEach(login => updateQuery[login] = {$set: action.selectAll});
				this.selectedUsers = update(this.selectedUsers, updateQuery);
				this.selectAll = action.selectAll;
				console.log(this.selectedUsers);
				break;
			}

		case DashboardActions.SelectUser:
			{
				let updateQuery = {};
				updateQuery[action.login] = {$set: action.select};
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
