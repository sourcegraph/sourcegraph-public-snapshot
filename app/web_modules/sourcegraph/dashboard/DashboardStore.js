import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";

export class DashboardStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
	}

	reset() {
		if (typeof window !== "undefined") { // TODO(autotest) support document object.
			this.repos = deepFreeze(window.repos || []);
			this.currentUser = deepFreeze(window._currentUser);
			this.onboarding = deepFreeze(window.onboarding);
		} else {
			this.repos = [];
			this.currentUser = {Name: "abc xyz"};
			this.onboarding = {};
		}
	}

	__onDispatch(action) {
		switch (action.constructor) {

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new DashboardStore(Dispatcher);
