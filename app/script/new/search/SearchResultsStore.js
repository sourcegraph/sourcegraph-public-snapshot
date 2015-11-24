import {Store} from "flux/utils";

import Dispatcher from "../Dispatcher";

export class SearchResultsStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
	}

	__onDispatch(action) {
		this.__emitChange();
	}
}

export default new SearchResultsStore(Dispatcher);
