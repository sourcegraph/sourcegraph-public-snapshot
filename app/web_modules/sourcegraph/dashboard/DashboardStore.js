import update from "react/lib/update";

import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

export class DashboardStore extends Store {
	reset() {
		this.repos = deepFreeze(window.repos);
		this.users = deepFreeze(window.users);
		this.isMothership = deepFreeze(window.isMothership);
		this.allowMirrors = Boolean(window.allowMirrors);
	}

	__onDispatch(action) {
		switch (action.constructor) {

		case DashboardActions.RepoCreated:
			this.repos = action.repos;
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
