import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DefActions from "sourcegraph/def/DefActions";

function defsListKeyFor(repo, rev, query) {
	return `${repo}#${rev}#${query}`;
}

function refsKeyFor(defURL, file) {
	return `${defURL}:${file || ""}`;
}

export class DefStore extends Store {
	reset(data) {
		this.defs = deepFreeze({
			content: data && data.defs ? data.defs : {},
			get(url) {
				return this.content[url] || null;
			},
			list(repo, rev, query) {
				return this.content[defsListKeyFor(repo, rev, query)] || null;
			},
		});
		this.activeDef = null;
		this.highlightedDef = null;
		this.refs = deepFreeze({
			content: {},
			get(defURL, file) {
				return this.content[refsKeyFor(defURL, file)] || null;
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

		case DefActions.DefsFetched:
			this.defs = deepFreeze(Object.assign({}, this.defs, {
				content: Object.assign({}, this.defs.content, {
					[defsListKeyFor(action.repo, action.rev, action.query)]: action.defs,
				}),
			}));
			break;

		case DefActions.HighlightDef:
			this.highlightedDef = action.url;
			break;

		case DefActions.RefsFetched:
			this.refs = deepFreeze(Object.assign({}, this.refs, {
				content: Object.assign({}, this.refs.content, {
					[refsKeyFor(action.defURL, action.file)]: action.refs,
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

export default new DefStore(Dispatcher.Stores);
