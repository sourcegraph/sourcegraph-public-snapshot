import * as GlobalNavStore from "sourcegraph/app/GlobalNav/GlobalNavStore";
import * as Dispatcher from "sourcegraph/Dispatcher";

export const GlobalNavBackend = {
	__onDispatch(payload: GlobalNavStore.Action): void {
		if (payload instanceof GlobalNavStore.ToggleQuickOpen) {
			setTimeout(() => Dispatcher.Stores.dispatch(payload), 1);
		}
	},
};

Dispatcher.Backends.register(GlobalNavBackend.__onDispatch);
