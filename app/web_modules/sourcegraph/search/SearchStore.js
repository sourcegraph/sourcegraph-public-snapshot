// @flow weak

import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as SearchActions from "sourcegraph/search/SearchActions";

function keyForResults(query: string, repos: ?Array<string>, notRepos: ?Array<string>, commitID: ?string, limit: ?number): string {
	return `${query}#${repos ? repos.join(":") : ""}#${notRepos ? notRepos.join(":") : ""}#${commitID ? commitID : ""}#${limit || ""}`;
}

export class SearchStore extends Store {
	content: any;

	reset(data?: {results: any}) {
		this.content = deepFreeze(data && data.results ? data.results : {});
	}

	get(query: string, repos: ?Array<string>, notRepos: ?Array<string>, commitID: ?string, limit: number): ?any {
		return this.content[keyForResults(query, repos, notRepos, commitID, limit)] || null;
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case SearchActions.ResultsFetched: {
			let p: SearchActions.ResultsFetchedPayload = action.p;
			this.content = deepFreeze({
				...this.content,
				[keyForResults(p.query, p.repos, p.notRepos, p.commitID, p.limit)]: p,
			});
			break;
		}
		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

let store_: SearchStore = new SearchStore(Dispatcher.Stores);
export default store_;
