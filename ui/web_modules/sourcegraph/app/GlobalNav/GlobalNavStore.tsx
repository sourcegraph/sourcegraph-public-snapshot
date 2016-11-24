import * as Dispatcher from "sourcegraph/Dispatcher";
import { Store } from "sourcegraph/Store";

export type Action =
	SetQuickOpenVisible;

export class SetQuickOpenVisible {
	quickOpenVisible: boolean;
	constructor(quickOpenVisible: boolean) {
		this.quickOpenVisible = quickOpenVisible;
	}
}

class GlobalNavStoreClass extends Store<any> {
	quickOpenVisible: boolean;
	reset(): void {
		this.quickOpenVisible = false;
	}

	toJSON(): any {
		return {
			quickOpenVisible: this.quickOpenVisible,
		};
	}

	__onDispatch(action: Action): void {
		switch (action.constructor) {
			case SetQuickOpenVisible:
				this.quickOpenVisible = action.quickOpenVisible;
				break;
			default:
				return; // don't emit change
		}
		this.__emitChange();
	}
}

export const GlobalNavStore = new GlobalNavStoreClass(Dispatcher.Stores);
