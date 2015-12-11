import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as CodeActions from "sourcegraph/code/CodeActions";

function keyFor(repo, rev, tree) {
	return `${repo}#${rev}#${tree}`;
}

export class CodeStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.files = deepFreeze({
			content: {},
			get(repo, rev, tree) {
				return this.content[keyFor(repo, rev, tree)] || null;
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

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new CodeStore(Dispatcher);
