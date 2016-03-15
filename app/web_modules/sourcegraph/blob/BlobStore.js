import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import prepareAnnotations from "sourcegraph/blob/prepareAnnotations";

function keyFor(repo, rev, tree) {
	return `${repo}#${rev}#${tree}`;
}

function keyForAnns(repo, rev, commitID, path, startByte, endByte) {
	return `${repo}#${rev}#${commitID}#${path}#${startByte || 0}#${endByte || 0}`;
}

export class BlobStore extends Store {
	reset() {
		this.files = deepFreeze({
			content: {},
			get(repo, rev, tree) {
				return this.content[keyFor(repo, rev, tree)] || null;
			},
		});

		// annotations are assumed to be sorted (with Annotations.sortAnns) by all callers of BlobStore.
		this.annotations = deepFreeze({
			content: {},
			get(repo, rev, commitID, path, startByte, endByte) {
				return this.content[keyForAnns(repo, rev, commitID, path, startByte, endByte)] || null;
			},
		});
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case BlobActions.FileFetched:
			this.files = deepFreeze(Object.assign({}, this.files, {
				content: Object.assign({}, this.files.content, {
					[keyFor(action.repo, action.rev, action.tree)]: action.file,
				}),
			}));
			break;

		case BlobActions.AnnotationsFetched:
			this.annotations = deepFreeze(Object.assign({}, this.annotations, {
				content: Object.assign({}, this.annotations.content, {
					[keyForAnns(action.repo, action.rev, action.commitID, action.path, action.startByte, action.endByte)]: action.annotations,
				}),
			}));
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

function prepareAnnotationsInPlace(anns) {
	Object.keys(anns).forEach((key) => {
		anns[key].Annotations = prepareAnnotations(anns[key].Annotations);
	});
	return anns;
}

export default new BlobStore(Dispatcher);
