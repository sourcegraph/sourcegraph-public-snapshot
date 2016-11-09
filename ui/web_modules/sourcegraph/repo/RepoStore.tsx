// tslint:disable: typedef

import * as filter from "lodash/filter";
import * as flatten from "lodash/flatten";
import * as some from "lodash/some";

import * as Dispatcher from "sourcegraph/Dispatcher";
import { languagesToSearchModes } from "sourcegraph/editor/modes";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import "sourcegraph/repo/RepoBackend";
import { Store } from "sourcegraph/Store";
import { deepFreeze } from "sourcegraph/util/deepFreeze";

function keyFor(repo: string, rev?: string) {
	return `${repo}@${rev || ""}`;
}

function keyForSymbols(mode: string, repo: string, rev?: string | null, query?: string | null) {
	return `(mode:${mode})${repo}@${rev || ""}${query ? "?" + query : ""}`;
}

class RepoStoreClass extends Store<any> {
	repos: any;
	resolvedRevs: any;
	commits: any;
	resolutions: any;
	symbols: {
		content: any;
		list(languages: string[], repo: string, rev: string | null, query: string): {
			loading: boolean;
			results: any[];
		};
	};

	reset() {
		this.repos = deepFreeze({
			content: {},
			get(repo) {
				return this.content[keyFor(repo)] || null;
			},
			listContent: {},
			list(querystring: string) {
				return this.listContent[querystring] || null;
			},
			cloning: {},
			isCloning(repo) {
				return this.cloning[keyFor(repo)] || false;
			},
		});
		this.resolvedRevs = deepFreeze({
			content: {},
			get(repo, rev) {
				return this.content[keyFor(repo, rev)] || null;
			},
		});
		this.resolutions = deepFreeze({
			content: {},
			get(repo) {
				return this.content[keyFor(repo)] || null;
			},
		});
		this.symbols = deepFreeze({
			content: {},
			list(languages: string[], repo: string, rev: string | null, query: string) {
				const langResults = [];
				languagesToSearchModes(languages).forEach((mode) => langResults.push(this.content[keyForSymbols(mode, repo, rev, query)]));
				const results = flatten(filter(langResults));
				const loading = some(langResults, r => r === undefined);
				return {
					results: results,
					loading: loading,
				};
			},
		});
	}

	toJSON(): any {
		return {
			repos: this.repos,
			resolvedRevs: this.resolvedRevs,
			resolutions: this.resolutions,
			symbols: this.symbols,
		};
	}

	__onDispatch(action: any) {
		if (action instanceof RepoActions.ReposFetched) {
			this.repos = deepFreeze(Object.assign({}, this.repos, {
				listContent: Object.assign({}, this.repos.listContent, {
					[action.querystring]: action.data,
				}),
			}));
			this.__emitChange();
			return;
		} else if (action instanceof RepoActions.ResolvedRev) {
			this.resolvedRevs = deepFreeze(Object.assign({}, this.resolvedRevs, {
				content: Object.assign({}, this.resolvedRevs.content, {
					[keyFor(action.repo, action.rev)]: action.commitID,
				}),
			}));
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
							[keyFor(action.repo)]: { Result: { Repo: action.repoObj.URI } },
						}),
					}));
				}
				break;

			case RepoActions.FetchedSymbols:
				this.symbols = deepFreeze(Object.assign({}, this.symbols, {
					content: Object.assign({}, this.symbols.content, {
						[keyForSymbols(action.mode, action.repo, action.rev, action.query)]: action.symbols,
					}),
				}));
				break;

			default:
				return; // don't emit change
		}

		this.__emitChange();
	}
}

export const RepoStore = new RepoStoreClass(Dispatcher.Stores);
