import { TreeEntry } from "sourcegraph/api";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import * as Dispatcher from "sourcegraph/Dispatcher";
import { Store } from "sourcegraph/Store";
import { deepFreeze } from "sourcegraph/util/deepFreeze";

// keyFor must stay in sync with the key func in
// (*ui.BlobStore).AddFile.
export function keyForFile(repo: string, commitID: string | null, path: string): string {
	return `${repo}#${commitID || ""}#${path}`;
}

class BlobStoreClass extends Store<any> {
	files: { [key: string]: TreeEntry };
	toast: string | null;
	_toastTimeout: number | null;

	reset(): void {
		this.files = deepFreeze({});
		this.toast = null;
		this._toastTimeout = null;
	}

	toJSON(): any {
		return { files: this.files, toast: this.toast };
	}

	__onDispatch(action: BlobActions.Action): void {
		if (action instanceof BlobActions.FileFetched) {
			this.files = deepFreeze(Object.assign({}, this.files, { [keyForFile(action.repo, action.commitID, action.path)]: action.file }));

		} else if (action instanceof BlobActions.Toast) {
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
