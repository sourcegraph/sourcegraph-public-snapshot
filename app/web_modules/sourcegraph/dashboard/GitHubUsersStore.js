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

export default new GitHubUsersStore(Dispatcher.Stores);
