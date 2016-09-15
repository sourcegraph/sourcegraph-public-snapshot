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

	reset(): void {
		this.files = deepFreeze({});
	}

	toJSON(): any {
		return { files: this.files };
	}

	__onDispatch(action: BlobActions.FileFetched): void {
		if (action instanceof BlobActions.FileFetched) {
			this.files = deepFreeze(Object.assign({}, this.files, { [keyForFile(action.repo, action.commitID, action.path)]: action.file }));
		} else {
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export const BlobStore = new BlobStoreClass(Dispatcher.Stores);
