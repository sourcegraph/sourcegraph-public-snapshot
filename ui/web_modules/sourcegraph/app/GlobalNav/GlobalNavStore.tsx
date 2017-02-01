import * as Dispatcher from "sourcegraph/Dispatcher";
import { Store } from "sourcegraph/Store";

export type Action =
	SetQuickOpenVisible | SetShortcutMenuVisible;

export class SetQuickOpenVisible {
	constructor(public quickOpenVisible: boolean) {
		this.quickOpenVisible = quickOpenVisible;
	}
}

export class SetShortcutMenuVisible {
	constructor(public shortcutMenuVisisble: boolean) {
		this.shortcutMenuVisisble = shortcutMenuVisisble;
	}
}

class GlobalNavStoreClass extends Store<any> {
	quickOpenVisible: boolean;
	shortcutMenuVisible: boolean;
	reset(): void {
		this.quickOpenVisible = false;
		this.shortcutMenuVisible = false;
	}

	toJSON(): any {
		return {
			quickOpenVisible: this.quickOpenVisible,
		};
	}

	__onDispatch(action: Action): void {
		if (action.constructor === SetQuickOpenVisible) {
			this.quickOpenVisible = (action as SetQuickOpenVisible).quickOpenVisible;
		} else if (action.constructor === SetShortcutMenuVisible) {
			this.shortcutMenuVisible = (action as SetShortcutMenuVisible).shortcutMenuVisisble;
		} else {
			return;
		}
		this.__emitChange();
	}
}

export const GlobalNavStore = new GlobalNavStoreClass(Dispatcher.Stores);
