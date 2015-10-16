import {Store} from "flux/utils";

import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";

export class DefStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.defs = {};
		this.highlightedDef = null;
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DefActions.DefFetched:
			this.defs[action.url] = action.def;
			this.__emitChange();
			break;
		}
	}
}

export default new DefStore(Dispatcher);
