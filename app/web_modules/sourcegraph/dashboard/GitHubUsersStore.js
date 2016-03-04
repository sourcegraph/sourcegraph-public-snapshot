import update from "react/lib/update";

import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

export class GitHubUsersStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.reset();
	}

	reset() {
		if (typeof window !== "undefined") { // TODO(autotest) support document object.
			// The users which are available to invite.
			this.users = deepFreeze({
				users: window.teammates ? (window.teammates.Users || []) : [],
				getUnique() {
					let seen = {};
					return this.users.filter(user => seen.hasOwnProperty(user.RemoteAccount.UID) ? false : (seen[user.RemoteAccount.UID] = true));
				},
			});
		} else {
			this.users = deepFreeze({
				users: [],
				getUnique() { return []; },
			});
		}
		// Store the state of which organizations mirrored users can come from.
		// The currentOrg is a filter for widget components.
		this.getByOrg = {};
		if (typeof window === "undefined" || !window.teammates) { // TODO support document object.
			this.orgs = {};
		} else {
			this.orgs = window.teammates.Organizations;
			this.users.users.map(user => this.getByOrg.hasOwnProperty(user.Organization) ? this.getByOrg[user.Organization].push(user) : this.getByOrg[user.Organization] = [].concat.apply(user || []));
		}
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DashboardActions.UsersInvited:
			this.users = update(this.users, {
				users: {$set: action.teammates ? action.teammates.Users : {}},
			});
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new GitHubUsersStore(Dispatcher);
