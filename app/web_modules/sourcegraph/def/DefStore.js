// @flow weak

import Store from "sourcegraph/Store";
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DefActions from "sourcegraph/def/DefActions";
import {defPath} from "sourcegraph/def";
import type {Def} from "sourcegraph/def";

function defKey(repo: string, rev: ?string, def: string): string {
	return `${repo}#${rev || ""}#${def}`;
}

function defsListKeyFor(repo: string, rev: string, query: string, filePathPrefix: ?string): string {
	return `${repo}#${rev}#${query}#${filePathPrefix || ""}`;
}

function refsKeyFor(repo: string, rev: ?string, def: string, refRepo: string, refFile: ?string): string {
	return `${defKey(repo, rev, def)}#${refRepo}#${refFile || ""}`;
}

export class DefStore extends Store {
	reset(data?: {defs: any, refs: any}) {
		this.defs = deepFreeze({
			content: data && data.defs ? data.defs.content : {},
			get(repo: string, rev: ?string, def: string): ?Def {
				return this.content[defKey(repo, rev, def)] || null;
			},
			list(repo: string, rev: string, query: string, filePathPrefix: ?string) {
				return this.content[defsListKeyFor(repo, rev, query, filePathPrefix)] || null;
			},
		});
		this.highlightedDef = null;
		this.refs = deepFreeze({
			content: data && data.refs ? data.refs.content : {},
			get(repo: string, rev: ?string, def: string, refRepo: string, refFile: ?string) {
				return this.content[refsKeyFor(repo, rev, def, refRepo, refFile)] || null;
			},
		});
		this.refLocations = deepFreeze({
			content: data && data.refLocations ? data.refLocations.content : {},
			get(repo: string, rev: ?string, def: string) {
				return this.content[defKey(repo, rev, def)] || null;
			},
		});
	}

	toJSON() {
		return {defs: this.defs, refs: this.refs, refLocations: this.refLocations};
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DefActions.DefFetched:
			this.defs = deepFreeze(Object.assign({}, this.defs, {
				content: Object.assign({}, this.defs.content, {
					[defKey(action.repo, action.rev, action.def)]: action.defObj,
				}),
			}));
			break;

		case DefActions.DefsFetched:
			{
				// Store the list of defs AND each def individually so we can
				// perform more operations quickly.
				let data = {
					[defsListKeyFor(action.repo, action.rev, action.query, action.filePathPrefix)]: action.defs,
				};
				if (action.defs && action.defs.Defs) {
					action.defs.Defs.forEach((d) => {
						data[defKey(d.Repo, action.rev, defPath(d))] = d;
					});
				}
				this.defs = deepFreeze(Object.assign({}, this.defs, {
					content: Object.assign({}, this.defs.content, data),
				}));
				break;
			}

		case DefActions.HighlightDef:
			this.highlightedDef = action.url;
			break;

		case DefActions.RefLocationsFetched:
			this.refLocations = deepFreeze(Object.assign({}, this.refLocations, {
				content: Object.assign({}, this.refLocations.content, {
					[defKey(action.repo, action.rev, action.def)]: getRankedRefLocations(action.locations),
				}),
			}));
			break;

		case DefActions.RefsFetched:
			this.refs = deepFreeze(Object.assign({}, this.refs, {
				content: Object.assign({}, this.refs.content, {
					[refsKeyFor(action.repo, action.rev, action.def, action.refRepo, action.refFile)]: action.refs,
				}),
			}));
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

function getRankedRefLocations(locations) {
	if (locations.length <= 2) {
		return locations;
	}
	let dashboardRepos = DashboardStore.repos;
	let repos = [];

	// The first repo of locations is the current repo.
	repos.push(locations[0]);

	let otherRepos = [];
	let i = 1;
	for (; i < locations.length; i++) {
		if (dashboardRepos && locations[i].Repo in dashboardRepos) {
			repos.push(locations[i]);
		} else {
			otherRepos.push(locations[i]);
		}
	}
	Array.prototype.push.apply(repos, otherRepos);
	return repos;
}

export default new DefStore(Dispatcher.Stores);
