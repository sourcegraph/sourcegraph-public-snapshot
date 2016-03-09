import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as RepoActions from "sourcegraph/repo/RepoActions";

function keyFor(repo) {
	return repo;
}

export class RepoStore extends Store {
	reset() {
		this.branches = deepFreeze({
			content: {},
			errors: {},
			error(repo) { return this.errors[keyFor(repo)] || null; },
			list(repo) {
				return this.content[keyFor(repo)] || null;
			},
		});
		this.tags = deepFreeze({
			content: {},
			errors: {},
			error(repo) { return this.errors[keyFor(repo)] || null; },
			list(repo) {
				return this.content[keyFor(repo)] || null;
			},
		});
	}

	__onDispatch(action) {
		switch (action.constructor) {

		case RepoActions.FetchedBranches:
			this.branches = deepFreeze(Object.assign({}, this.branches, {
				content: Object.assign({}, this.branches.content, {
					[keyFor(action.repo)]: action.branches,
				}),
				errors: Object.assign({}, this.branches.errors, {
					[keyFor(action.repo)]: action.err,
				}),
			}));
			break;

		case RepoActions.FetchedTags:
			this.tags = deepFreeze(Object.assign({}, this.tags, {
				content: Object.assign({}, this.tags.content, {
					[keyFor(action.repo)]: action.tags,
				}),
				errors: Object.assign({}, this.tags.errors, {
					[keyFor(action.repo)]: action.err,
				}),
			}));
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new RepoStore(Dispatcher);
