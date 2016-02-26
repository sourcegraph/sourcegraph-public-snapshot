import update from "react/lib/update";

import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

export class GitHubUsersStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		// The users which are available to invite.
		this.users = deepFreeze({
			users: window.teammates ? (window.teammates.Users || {}) : [],
			getUnique() {
				let seen = {};
				return this.users.filter(user => seen.hasOwnProperty(user.RemoteAccount.UID) ? false : (seen[user.RemoteAccount.UID] = true));
			},
		});

		// Store the state of which organizations mirrored users can come from.
		// The currentOrg is a filter for widget components.
		this.getByOrg = {};
		if (!window.teammates) {
			this.orgs = {};
		} else {
			this.orgs = window.teammates.Organizations;
			this.users.users.map(user => this.getByOrg.hasOwnProperty(user.Organization) ? this.getByOrg[user.Organization].push(user) : this.getByOrg[user.Organization] = [].concat.apply(user || []));
		}
		this.showLoading = false; // Indicates if a request to the backend to invite users is in progress
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DashboardActions.WantInviteUsers:
			this.showLoading = true;
			break;

		case DashboardActions.UsersInvited:
			this.users = update(this.users, {
				users: {$merge: {IsInvited: true}},
			});
			this.showLoading = false;
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new GitHubUsersStore(Dispatcher);
