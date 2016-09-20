import * as Dispatcher from "sourcegraph/Dispatcher";
import {Store} from "sourcegraph/Store";
import {Settings} from "sourcegraph/user";
import * as UserActions from "sourcegraph/user/UserActions";
import {deepFreeze} from "sourcegraph/util/deepFreeze";

class UserStoreClass extends Store<any> {
	settings: Settings;

	reset(): void {
		if (global.window) {
			let storedUserSettings = window.localStorage.getItem("userSettings");
			if (storedUserSettings) {
				this.settings = deepFreeze(JSON.parse(storedUserSettings) as Settings);
			} else {
				this.settings = deepFreeze({
					search: {
						languages: ["golang"],
						scope: {
							popular: true,
							public: false,
							private: false,
							repo: true,
						},
					},
				});
			}
		}
	}

	__onDispatch(action: UserActions.Action): void {
		if (action instanceof UserActions.UpdateSettings) {
			if (global.window) { window.localStorage.setItem("userSettings", JSON.stringify(action.settings)); }
			this.settings = deepFreeze(Object.assign({}, this.settings, action.settings));

		} else {
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export const UserStore = new UserStoreClass(Dispatcher.Stores);
