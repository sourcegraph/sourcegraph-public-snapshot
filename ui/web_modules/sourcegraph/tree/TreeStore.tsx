// tslint:disable: typedef ordered-imports

import {Store} from "sourcegraph/Store";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {deepFreeze} from "sourcegraph/util/deepFreeze";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import "sourcegraph/tree/TreeBackend";

function keyFor(repo, rev, path?) {
	return `${repo}#${rev}#${path || ""}`;
}

class TreeStoreClass extends Store<any> {
	commits: any;

	reset() {
		this.commits = deepFreeze({
			content: {},
			get(repo, rev, path) {
				return this.content[keyFor(repo, rev, path)] || null;
			},
		});
	}

	toJSON(): any {
		return {
			commits: this.commits,
		};
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

export const TreeStore = new TreeStoreClass(Dispatcher.Stores);
