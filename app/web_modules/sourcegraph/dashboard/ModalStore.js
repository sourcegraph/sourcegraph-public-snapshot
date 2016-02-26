import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

export class OnboardingStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.showUsersModal = false;
	}

	__onDispatch(action) {
		switch (action.constructor) {

		case DashboardActions.DismissUsersModal:
			this.showUsersModal = false;
			break;

		case DashboardActions.OpenAddUsersModal:
			this.showUsersModal = true;
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new OnboardingStore(Dispatcher);
