import {AnnotationList, TreeEntry} from "sourcegraph/api";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {Store} from "sourcegraph/Store";
import {deepFreeze} from "sourcegraph/util/deepFreeze";

// keyFor must stay in sync with the key func in
// (*ui.BlobStore).AddFile.
export function keyForFile(repo: string, commitID: string | null, path: string): string {
	return `${repo}#${commitID || ""}#${path}`;
}

// keyForAnns must stay in sync with the key func in
// (*ui.BlobStore).AddAnnotations.
export function keyForAnns(repo: string, commitID: string, path: string, startByte?: number | null, endByte?: number | null): string {
	return `${repo}#${commitID}#${path}#${startByte || 0}#${endByte || 0}`;
}

class BlobStoreClass extends Store<any> {
	files: {[key: string]: TreeEntry};
	annotations: {[key: string]: AnnotationList};

	reset(): void {
		this.files = deepFreeze({});

		// annotations are assumed to be sorted & prepared (with Annotations.prepareAnnotations).
		this.annotations = deepFreeze({});
	}

	toJSON(): any {
		return {files: this.files, annotations: this.annotations};
	}

	__onDispatch(action: BlobActions.FileFetched | BlobActions.AnnotationsFetched): void {
		if (action instanceof BlobActions.FileFetched) {
			this.files = deepFreeze(Object.assign({}, this.files, {[keyForFile(action.repo, action.commitID, action.path)]: action.file}));
		} else if (action instanceof BlobActions.AnnotationsFetched) {
			this.annotations = deepFreeze(Object.assign({}, this.annotations, {[keyForAnns(action.repo, action.commitID, action.path, action.startByte, action.endByte)]: action.annotations}));
		} else {
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export const BlobStore = new BlobStoreClass(Dispatcher.Stores);
