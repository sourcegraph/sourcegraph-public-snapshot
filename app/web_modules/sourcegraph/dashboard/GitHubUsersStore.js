import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";

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
	}

	__onDispatch(action) {
		switch (action.constructor) {
		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new GitHubUsersStore(Dispatcher);
