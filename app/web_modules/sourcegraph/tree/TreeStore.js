import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as TreeActions from "sourcegraph/tree/TreeActions";

function keyFor(repo, rev, path) {
	return `${repo}#${rev}#${path || ""}`;
}

export class TreeStore extends Store {
	reset() {
		this.commits = deepFreeze({
			content: {},
			get(repo, rev, path) {
				return this.content[keyFor(repo, rev, path)] || null;
			},
		});
		this.fileLists = deepFreeze({
			content: {},
			get(repo, rev) {
				return this.content[keyFor(repo, rev)] || null;
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

		case TreeActions.FileListFetched:
			this.fileLists = deepFreeze(Object.assign({}, this.fileLists, {
				content: Object.assign({}, this.fileLists.content, {
					[keyFor(action.repo, action.rev)]: action.fileList,
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
