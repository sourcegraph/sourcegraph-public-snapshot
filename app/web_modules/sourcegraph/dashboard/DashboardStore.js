import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

// import update from "react/lib/update"

export class DashboardStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.repos = deepFreeze(window.repos);
		this.users = deepFreeze(window.users);
		this.mirrorRepos = deepFreeze(window.mirrorRepos);
	}

	__onDispatch(action) {
		switch (action.constructor) {

		case DashboardActions.ReposAdded:
			break;

		case DashboardActions.UsersAdded:
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new DashboardStore(Dispatcher);
