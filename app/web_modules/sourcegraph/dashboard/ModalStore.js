import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

export class OnboardingStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.reset();
	}

	reset() {
		this.showReposModal = false;
		this.showUsersModal = false;
	}

	__onDispatch(action) {
		switch (action.constructor) {

		case DashboardActions.DismissReposModal:
			this.showReposModal = false;
			break;

		case DashboardActions.DismissUsersModal:
			this.showUsersModal = false;
			break;

		case DashboardActions.OpenAddUsersModal:
			this.showUsersModal = true;
			break;

		case DashboardActions.OpenAddReposModal:
			this.showReposModal = true;
			break;

		case DashboardActions.MirrorReposAdded:
			this.showReposModal = false;
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new OnboardingStore(Dispatcher);
