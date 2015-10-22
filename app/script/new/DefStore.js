import {Store} from "flux/utils";

import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";

function exampleKeyFor(defURL, index) {
	return `${defURL}#${index}`;
}

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
		this.examples = {
			content: {},
			counts: {},
			generation: 0,
			get(defURL, index) {
				return this.content[exampleKeyFor(defURL, index)] || null;
			},
			getCount(defURL) {
				return this.counts[defURL] || 1000; // high initial value until count is known
			},
		};
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DefActions.DefFetched:
			this.defs.content[action.url] = action.def;
			break;

		case DefActions.HighlightDef:
			this.highlightedDef = action.url;
			break;

		case DefActions.ExampleFetched:
			this.examples.content[exampleKeyFor(action.defURL, action.index)] = action.example;
			this.examples.generation++;
			break;

		case DefActions.NoExampleAvailable:
			this.examples.counts[action.defURL] = Math.min(this.examples.getCount(action.defURL), action.index);
			this.examples.generation++;
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new DefStore(Dispatcher);
