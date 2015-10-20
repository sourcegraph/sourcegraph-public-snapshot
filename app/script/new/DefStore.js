import {Store} from "flux/utils";

import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";

export class DefStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.defs = {
			content: {},
			get(url) {
				return this.content[url] || null;
			},
		};
		this.highlightedDef = null;
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DefActions.DefFetched:
			this.defs.content[action.url] = action.def;
			break;

		case DefActions.HighlightDef:
			this.highlightedDef = action.url;
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new DefStore(Dispatcher);
