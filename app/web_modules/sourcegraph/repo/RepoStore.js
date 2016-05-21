// @flow weak

import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import "sourcegraph/repo/RepoBackend";

function keyFor(repo, rev) {
	return `${repo}@${rev}`;
}

export class RepoStore extends Store {
	reset(data?: {repos: any, resolvedRevs: any, resolutions: any, branches: any, tags: any}) {
		this.repos = deepFreeze({
			content: data && data.repos ? data.repos.content : {},
			get(repo) {
				return this.content[keyFor(repo)] || null;
			},
			cloning: data && data.repos ? data.repos.cloning : {},
			isCloning(repo) {
				return this.cloning[keyFor(repo)] || false;
			},
		});
		this.resolvedRevs = deepFreeze({
			content: data && data.resolvedRevs ? data.resolvedRevs.content : {},
			get(repo, rev) {
				return this.content[keyFor(repo, rev)] || null;
			},
		});
		this.resolutions = deepFreeze({
			content: data && data.resolutions ? data.resolutions.content : {},
			get(repo) {
				return this.content[keyFor(repo)] || null;
			},
		});
		this.inventory = deepFreeze({
			content: data && data.inventory ? data.inventory.content : {},
			get(repo, commitID) {
				return this.content[keyFor(repo, commitID)] || null;
			},
		});
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

	toJSON(): any {
		return {
			repos: this.repos,
			resolvedRevs: this.resolvedRevs,
			resolutions: this.resolutions,
			branches: this.branches,
			tags: this.tags,
			inventory: this.inventory,
		};
	}

	__onDispatch(action) {
		if (action instanceof RepoActions.ResolvedRev) {
			this.resolvedRevs = deepFreeze({...this.resolvedRevs,
				content: {...this.resolvedRevs.content,
					[keyFor(action.repo, action.rev)]: action.commitID,
				},
			});
			this.__emitChange();
			return;
		}

		switch (action.constructor) {

		case RepoActions.FetchedRepo:
			this.repos = deepFreeze(Object.assign({}, this.repos, {
				content: Object.assign({}, this.repos.content, {
					[keyFor(action.repo)]: action.repoObj,
				}),
			}));
			break;

		case RepoActions.FetchedInventory:
			this.inventory = deepFreeze(Object.assign({}, this.inventory, {
				content: Object.assign({}, this.inventory.content, {
					[keyFor(action.repo, action.commitID)]: action.inventory,
				}),
			}));
			break;

		case RepoActions.RepoCloning:
			this.repos = deepFreeze(Object.assign({}, this.repos, {
				cloning: Object.assign({}, this.repos.cloning, {
					[keyFor(action.repo)]: action.isCloning,
				}),
			}));
			break;

		case RepoActions.RepoResolved:
			this.resolutions = deepFreeze(Object.assign({}, this.resolutions, {
				content: Object.assign({}, this.resolutions.content, {
					[keyFor(action.repo)]: action.resolution,
				}),
			}));
			break;

		case RepoActions.RepoCreated:
			this.repos = deepFreeze(Object.assign({}, this.repos, {
				content: Object.assign({}, this.repos.content, {
					[keyFor(action.repo)]: action.repoObj,
				}),
			}));

			if (!action.repoObj.Error) {
				// Update resolution to reflect the newly created repo.
				this.resolutions = deepFreeze(Object.assign({}, this.resolutions, {
					content: Object.assign({}, this.resolutions.content, {
						[keyFor(action.repo)]: {Result: {Repo: {URI: action.repoObj.URI}}},
					}),
				}));
			}
			break;

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

export default new RepoStore(Dispatcher.Stores);
