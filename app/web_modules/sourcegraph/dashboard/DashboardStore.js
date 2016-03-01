import update from "react/lib/update";

import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

// import update from "react/lib/update"

export class DashboardStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		if (typeof window !== "undefined") { // TODO(autotest) support document object.
			this.repos = deepFreeze(window.repos);
			this.users = deepFreeze(window.users);
			this.currentUser = deepFreeze(window.currentUser);
			this.onboarding = deepFreeze(window.onboarding);
			this.isMothership = deepFreeze(window.isMothership);
			this.onWaitlist = deepFreeze(window.onWaitlist);
			this.allowMirrors = Boolean(window.allowMirrors);
		} else {
			this.repos = [];
			this.users = [];
			this.currentUser = {Name: "abc xyz"};
			this.onboarding = {};
			this.isMothership = true;
			this.onWaitlist = true;
			this.allowMirrors = true;
		}
	}

	__onDispatch(action) {
		switch (action.constructor) {

		case DashboardActions.RepoCreated:
			this.repos = action.repos;
			break;

		case DashboardActions.MirrorRepoAdded:
			window.location.href = `/${action.repo.URI}`;
			break;

		case DashboardActions.UserInvited:
			this.users = update(this.users, {$push: [action.user]});
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new DashboardStore(Dispatcher);
