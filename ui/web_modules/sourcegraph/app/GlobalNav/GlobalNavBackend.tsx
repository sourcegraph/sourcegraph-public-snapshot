import * as GlobalNavStore from "sourcegraph/app/GlobalNav/GlobalNavStore";
import * as Dispatcher from "sourcegraph/Dispatcher";

export const GlobalNavBackend = {
	__onDispatch(payload: GlobalNavStore.Action): void {
		if (payload instanceof GlobalNavStore.SetQuickOpenVisible) {
			setTimeout(() => Dispatcher.Stores.dispatch(payload));
		} else if (payload instanceof GlobalNavStore.SetShortcutMenuVisible) {
			setTimeout(() => Dispatcher.Stores.dispatch(payload));
		}
	},
};

Dispatcher.Backends.register(GlobalNavBackend.__onDispatch);
