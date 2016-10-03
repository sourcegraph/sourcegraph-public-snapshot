import * as Dispatcher from "sourcegraph/Dispatcher";
import {Store} from "sourcegraph/Store";

export type Action =
	ToggleQuickOpen;

export class ToggleQuickOpen {
	quickOpenVisible: boolean;
	constructor(quickOpenVisible: boolean = false) {
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

	__onDispatch(action: any): void {
		switch (action.constructor) {
			case ToggleQuickOpen:
				this.quickOpenVisible = (action as ToggleQuickOpen).quickOpenVisible;
				break;
			default:
				return; // don't emit change
		}
		this.__emitChange();
	}
}

export const GlobalNavStore = new GlobalNavStoreClass(Dispatcher.Stores);
