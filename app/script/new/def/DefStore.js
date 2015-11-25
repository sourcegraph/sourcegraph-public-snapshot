import {Store} from "flux/utils";

import Dispatcher from "../Dispatcher";
import deepFreeze from "../util/deepFreeze";
import * as DefActions from "./DefActions";

function exampleKeyFor(defURL, index) {
	return `${defURL}#${index}`;
}

export class DefStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.defs = deepFreeze({
			content: {},
			get(url) {
				return this.content[url] || null;
			},
		});
		this.highlightedDef = null;
		this.examples = deepFreeze({
			content: {},
			counts: {},
			get(defURL, index) {
				return this.content[exampleKeyFor(defURL, index)] || null;
			},
			getCount(defURL) {
				return this.counts.hasOwnProperty(defURL) ? this.counts[defURL] : 1000; // high initial value until count is known
			},
		});

		this.defOptionsURLs = null;
		this.defOptionsLeft = 0;
		this.defOptionsTop = 0;
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DefActions.DefFetched:
			this.defs = deepFreeze(Object.assign({}, this.defs, {
				content: Object.assign({}, this.defs.content, {
					[action.url]: action.def,
				}),
			}));
			break;

		case DefActions.HighlightDef:
			this.highlightedDef = action.url;
			break;

		case DefActions.ExampleFetched:
			this.examples = deepFreeze(Object.assign({}, this.examples, {
				content: Object.assign({}, this.examples.content, {
					[exampleKeyFor(action.defURL, action.index)]: action.example,
				}),
			}));
			break;

		case DefActions.NoExampleAvailable:
			this.examples = deepFreeze(Object.assign({}, this.examples, {
				counts: Object.assign({}, this.examples.counts, {
					[action.defURL]: Math.min(this.examples.getCount(action.defURL), action.index),
				}),
			}));
			break;

		case DefActions.SelectMultipleDefs:
			this.defOptionsURLs = action.urls;
			this.defOptionsLeft = action.left;
			this.defOptionsTop = action.top;
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new DefStore(Dispatcher);
