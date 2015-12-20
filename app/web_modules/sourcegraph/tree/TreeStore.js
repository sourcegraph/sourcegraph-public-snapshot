import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as TreeActions from "sourcegraph/tree/TreeActions";

function keyFor(repo, rev, path) {
	return `${repo}#${rev}#${path}`;
}

export class TreeStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.commits = deepFreeze({
			content: {},
			get(repo, rev, path) {
				return this.content[keyFor(repo, rev, path)] || null;
			},
		});
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case TreeActions.CommitFetched:
			this.commits = deepFreeze(Object.assign({}, this.commits, {
				content: Object.assign({}, this.commits.content, {
					[keyFor(action.repo, action.rev, action.path)]: action.commit,
				}),
			}));
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new TreeStore(Dispatcher);
