import * as BlobActions from "sourcegraph/blob/BlobActions";
import * as Dispatcher from "sourcegraph/Dispatcher";
import { Store } from "sourcegraph/Store";

class BlobStoreClass extends Store<any> {
	toast: string | null;
	_toastTimeout: number | null;

	reset(): void {
		this.toast = null;
		this._toastTimeout = null;
	}

	toJSON(): any {
		return { toast: this.toast };
	}

	__onDispatch(action: BlobActions.Action): void {
		if (action instanceof BlobActions.Toast) {
			this.toast = action.msg;

			if (this._toastTimeout) {
				clearTimeout(this._toastTimeout);
			}
			this._toastTimeout = setTimeout(() => {
				Dispatcher.Stores.dispatch(new BlobActions.ClearToast());
			}, action.timeout);

		} else if (action instanceof BlobActions.ClearToast) {
			if (this._toastTimeout) {
				clearTimeout(this._toastTimeout);
			}
			this.toast = null;

		} else {
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export const BlobStore = new BlobStoreClass(Dispatcher.Stores);
