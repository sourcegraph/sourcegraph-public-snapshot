import * as Dispatcher from "sourcegraph/Dispatcher";
import { Store } from "sourcegraph/Store";

export type Action = SetShortcutMenuVisible;

export class SetShortcutMenuVisible {
	constructor(public shortcutMenuVisisble: boolean) {
		this.shortcutMenuVisisble = shortcutMenuVisisble;
	}
}

class GlobalNavStoreClass extends Store<any> {
	shortcutMenuVisible: boolean;
	reset(): void {
		this.shortcutMenuVisible = false;
	}

	__onDispatch(action: Action): void {
		if (action.constructor === SetShortcutMenuVisible) {
			this.shortcutMenuVisible = (action as SetShortcutMenuVisible).shortcutMenuVisisble;
			this.__emitChange();
		}
	}
}

export const GlobalNavStore = new GlobalNavStoreClass(Dispatcher.Stores);
