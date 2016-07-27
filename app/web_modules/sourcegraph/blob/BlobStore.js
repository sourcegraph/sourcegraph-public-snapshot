import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as BlobActions from "sourcegraph/blob/BlobActions";

// keyFor must stay in sync with the key func in
// (*ui.BlobStore).AddFile.
function keyFor(repo: string, rev: ?string, path: string) {
	return `${repo}#${rev || ""}#${path}`;
}

// keyForAnns must stay in sync with the key func in
// (*ui.BlobStore).AddAnnotations.
function keyForAnns(repo: string, commitID: string, path: string, startByte: ?number, endByte: ?number) {
	return `${repo}#${commitID}#${path}#${startByte || 0}#${endByte || 0}`;
}

export class BlobStore extends Store {
	reset(data?: {files: any, annotations: any}) {
		this.files = deepFreeze({
			content: data && data.files ? data.files.content : {},
			get(repo, commitID, path) {
				return this.content[keyFor(repo, commitID, path)] || null;
			},
		});

		// annotations are assumed to be sorted & prepared (with Annotations.prepareAnnotations).
		this.annotations = deepFreeze({
			content: data && data.annotations ? data.annotations.content : {},
			get(repo, commitID, path, startByte, endByte) {
				return this.content[keyForAnns(repo, commitID, path, startByte, endByte)] || null;
			},
		});
	}

	toJSON(): any {
		return {files: this.files, annotations: this.annotations};
	}

	__onDispatch(action: BlobActions.FileFetched | BlobActions.AnnotationsFetched) {
		if (action instanceof BlobActions.FileFetched) {
			this.files = deepFreeze(Object.assign({}, this.files, {
				content: {...this.files.content,
					[keyFor(action.repo, action.commitID, action.path)]: action.file,
				},
			}));
		} else if (action instanceof BlobActions.AnnotationsFetched) {
			this.annotations = deepFreeze(Object.assign({}, this.annotations, {
				content: Object.assign({}, this.annotations.content, {
					[keyForAnns(action.repo, action.commitID, action.path, action.startByte, action.endByte)]: action.annotations,
				}),
			}));
		} else {
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new BlobStore(Dispatcher.Stores);
