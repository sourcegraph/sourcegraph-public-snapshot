import {Store} from "flux/utils";

import Dispatcher from "./Dispatcher";
import * as CodeActions from "./CodeActions";

function keyFor(repo, rev, tree) {
	return `${repo}#${rev}#${tree}`;
}

export class CodeStoreClass extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.files = {
			content: {},
			get(repo, rev, tree) {
				return this.content[keyFor(repo, rev, tree)];
			},
		};
		this.highlightedDef = null;
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case CodeActions.FileFetched:
			this.files.content[keyFor(action.repo, action.rev, action.tree)] = action.file;
			this.__emitChange();
			break;

		case CodeActions.HighlightDef:
			this.highlightedDef = action.def;
			this.__emitChange();
			break;

		}
	}
}

export default new CodeStoreClass(Dispatcher);
