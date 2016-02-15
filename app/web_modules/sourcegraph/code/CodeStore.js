import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as CodeActions from "sourcegraph/code/CodeActions";

function keyFor(repo, rev, tree) {
	return `${repo}#${rev}#${tree}`;
}

function keyForAnns(repo, rev, path, startByte, endByte) {
	return `${repo}#${rev}#${path}#${startByte || 0}#${endByte || 0}`;
}

export class CodeStore extends Store {
	reset() {
		this.files = deepFreeze({
			content: {},
			get(repo, rev, tree) {
				return this.content[keyFor(repo, rev, tree)] || null;
			},
		});
		this.annotations = deepFreeze({
			content: {},
			get(repo, rev, path, startByte, endByte) {
				return this.content[keyForAnns(repo, rev, path, startByte, endByte)] || null;
			},
		});
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case CodeActions.FileFetched:
			this.files = deepFreeze(Object.assign({}, this.files, {
				content: Object.assign({}, this.files.content, {
					[keyFor(action.repo, action.rev, action.tree)]: action.file,
				}),
			}));
			break;

		case CodeActions.AnnotationsFetched:
			this.annotations = deepFreeze(Object.assign({}, this.annotations, {
				content: Object.assign({}, this.annotations.content, {
					[keyForAnns(action.repo, action.rev, action.path, action.startByte, action.endByte)]: action.annotations,
				}),
			}));
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new CodeStore(Dispatcher);
