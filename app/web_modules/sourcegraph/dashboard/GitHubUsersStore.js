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
			// TODO: make it easier to get a single user.
			get(login) {
				let allUsers = Object.keys(this.users).reduce(
					(users, org) => users.concat(this.users[org].Users), []
				);
				for (let user of allUsers) {
					if (user.RemoteAccount.Login === login) return user;
				}
				return null;
			},
			getByOrg(org) {
				let orgUsers = this.users[org];
				return orgUsers ? orgUsers.Users : [];
			},
			getAdded() {
				// TODO(rothfels): this is gross and should be cleaned up...but is necessary to show mirrored repos on the dashboard.
				// We should probably build the map from org => user in this store and just have the server return a flat list.
				// let allUsers = (Object.values(this.users) || []).map(orgUsers => orgUsers.Users);
				// allUsers = [].concat.apply([], allUsers);
				console.log(this.users);
				let allUsers = this.users
					.filter(user => user.LocalAccount || user.IsInvited)
					.map(user => {
						if (user.LocalAccount) return user.LocalAccount;
						return update(user.RemoteAccount, {$merge: {IsInvited: true}});
					});
				// Deduplicate users.
				let userMap = {};
				allUsers.forEach(user => {
					// The GitHub UID (for invited users) may collide with the
					// Sourcegraph.com UID.
					if (user.IsInvited) {
						userMap[`invited-${user.UID}`] = user;
					} else {
						userMap[user.UID] = user;
					}
				});
				// Show invited users last.
				return Object.values(userMap).sort((a, b) => {
					if (a.IsInvited && !b.IsInvited) return 1;
					if (!a.IsInvited && b.IsInvited) return -1;
					return 0;
				});
			},
		});

		// Store the state of which organizations mirrored users can come from.
		// The currentOrg is a filter for widget components.
		if (!window.teammates) {
			this.orgs = {};
			this.getByOrg = {};
		} else {
			this.orgs = window.teammates.Organizations;
			this.getByOrg = {};
			// this.users.users.map(user => this.getByOrg[user.Organization] = [].concat.apply(this.getByOrg[user.Organization], user));
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
				users: {$set: action.teammates ? action.teammates.UsersByOrg : {}},
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
