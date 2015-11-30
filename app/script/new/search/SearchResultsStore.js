import {Store} from "flux/utils";

import Dispatcher from "../Dispatcher";
import deepFreeze from "../util/deepFreeze";
import * as SearchActions from "./SearchActions";

function keyFor(repo, rev, query, type, page) {
	return `${repo}#${rev}#${query}#${type}#${page}`;
}

export class SearchResultsStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.results = deepFreeze({
			content: {},
			get(repo, rev, query, type, page) {
				return this.content[keyFor(repo, rev, query, type, page)] || null;
			},
		});
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case SearchActions.ResultsFetched:
			this.results = deepFreeze(Object.assign({}, this.results, {
				content: Object.assign({}, this.results.content, {
					[keyFor(action.repo, action.rev, action.query, action.type, action.page)]: action.results,
				}),
			}));
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new SearchResultsStore(Dispatcher);
