import {Store} from "flux/utils";

import Dispatcher from "../Dispatcher";
import * as SearchActions from "./SearchActions";

function keyFor(repo, rev, type, page) {
	return `${repo}#${rev}#${type}#${page}`;
}

export class SearchResultsStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.results = {
			content: {},
			generation: 0,
			get(repo, rev, type, page) {
				return this.content[keyFor(repo, rev, type, page)] || null;
			},
		};
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case SearchActions.ResultsFetched:
			this.results.content[keyFor(action.repo, action.rev, action.type, action.page)] = action.results;
			this.results.generation++;
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new SearchResultsStore(Dispatcher);
