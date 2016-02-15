import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as OnboardingActions from "sourcegraph/dashboard/OnboardingActions";

import update from "react/lib/update";

export class OnboardingStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.progress = deepFreeze(window.progress);
		this.currentUser = deepFreeze(window.currentUser);
	}

	__onDispatch(action) {
		switch (action.constructor) {

		case OnboardingActions.AdvanceProgressStep:
			this.progress = update(this.progress, {
				currentStep: {$set: this.progress.currentStep + 1},
			});
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new OnboardingStore(Dispatcher);
