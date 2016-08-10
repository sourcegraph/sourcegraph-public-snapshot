// tslint:disable: typedef ordered-imports curly

import {Store} from "sourcegraph/Store";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {deepFreeze} from "sourcegraph/util/deepFreeze";
import * as SearchActions from "sourcegraph/search/SearchActions";

function keyForResults(query: string, repos: string[] | null, notRepos: string[] | null, commitID: string | null, limit: number | null): string {
	return `${query}#${repos ? repos.join(":") : ""}#${notRepos ? notRepos.join(":") : ""}#${commitID ? commitID : ""}#${limit || ""}`;
}

class SearchStoreClass extends Store<any> {
	content: any;

	reset() {
		this.content = deepFreeze({});
	}

	get(query: string, repos: string[] | null, notRepos: string[] | null, commitID: string | null, limit: number): any {
		return this.content[keyForResults(query, repos, notRepos, commitID, limit)] || null;
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case SearchActions.ResultsFetched: {
			let p: SearchActions.ResultsFetchedPayload = action.p;
			this.content = deepFreeze(Object.assign({}, this.content, {
				[keyForResults(p.query, p.repos, p.notRepos, p.commitID || null, p.limit || null)]: p.defs,
			}));
			break;
		}
		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export const SearchStore = new SearchStoreClass(Dispatcher.Stores);
