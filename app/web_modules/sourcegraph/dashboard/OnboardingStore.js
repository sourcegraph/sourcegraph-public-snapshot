import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as OnboardingActions from "sourcegraph/dashboard/OnboardingActions";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

import update from "react/lib/update";

export class OnboardingStore extends Store {
	reset() {
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
		case OnboardingActions.AdvanceProgressToStep:
			this.progress = update(this.progress, {
				currentStep: {$set: action.step},
			});
			break;
		case DashboardActions.MirrorReposAdded:
			if (this.progress.currentStep === 2) {
				this.progress = update(this.progress, {
					currentStep: {$set: 3},
				});
			}
			break;
		case DashboardActions.UsersInvited:
			if (this.progress.currentStep === 3) {
				this.progress = update(this.progress, {
					currentStep: {$set: 4},
				});
			}
			break;
		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new OnboardingStore(Dispatcher);
