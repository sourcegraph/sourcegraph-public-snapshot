// @flow weak

import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as SearchActions from "sourcegraph/search/SearchActions";

export class SearchStore extends Store {
	reset(data?: {results: any}) {
		this.results = deepFreeze({
			content: data && data.results ? data.results : {},
			get(query) {
				return this.content[query] || null;
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
					[action.query]: action.defs,
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
