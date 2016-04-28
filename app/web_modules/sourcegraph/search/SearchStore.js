// @flow weak

import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as SearchActions from "sourcegraph/search/SearchActions";

function keyForResults(query: string, repos: ?Array<string>, notRepos: ?Array<string>, limit: ?number): string {
	return `${query}#${repos ? repos.join(":") : ""}#${notRepos ? notRepos.join(":") : ""}#${limit || ""}`;
}

export class SearchStore extends Store {
	reset(data?: {results: any}) {
		this.results = deepFreeze({
			content: data && data.results ? data.results : {},
			get(query, repos, notRepos, limit) {
				return this.content[keyForResults(query, repos, notRepos, limit)] || null;
			},
		});
	}

	toJSON() {
		return {results: this.results};
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case SearchActions.ResultsFetched:
			this.results = deepFreeze({
				...this.results,
				content: {
					...this.results.content,
					[keyForResults(action.query, action.repos, action.notRepos, action.limit)]: action.defs,
				},
			});
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new SearchStore(Dispatcher.Stores);
