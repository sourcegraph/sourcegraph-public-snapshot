import {Store} from "flux/utils";

import Dispatcher from "./Dispatcher";
import * as CodeActions from "./CodeActions";

function keyFor(repo, rev, tree) {
	return `${repo}#${rev}#${tree}`;
}

export class CodeStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.files = {
			content: {},
			generation: 0,
			get(repo, rev, tree) {
				return this.content[keyFor(repo, rev, tree)] || null;
			},
		};
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case CodeActions.FileFetched:
			this.files.content[keyFor(action.repo, action.rev, action.tree)] = action.file;
			this.files.generation++;
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new CodeStore(Dispatcher);
