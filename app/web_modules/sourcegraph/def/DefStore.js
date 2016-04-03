// @flow weak

import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DefActions from "sourcegraph/def/DefActions";
import {defPath} from "sourcegraph/def";
import type {Def} from "sourcegraph/def";

function defKey(repo: string, rev: ?string, def: string): string {
	return `${repo}#${rev || ""}#${def}`;
}

function defsListKeyFor(repo, rev, query) {
	return `${repo}#${rev}#${query}`;
}

function refsKeyFor(repo: string, rev: ?string, def: string, file: ?string): string {
	return `${defKey(repo, rev, def)}#${file || ""}`;
}

export class DefStore extends Store {
	reset(data?: {defs: any, refs: any}) {
		this.defs = deepFreeze({
			content: data && data.defs ? data.defs.content : {},
			get(repo: string, rev: ?string, def: string): ?Def {
				return this.content[defKey(repo, rev, def)] || null;
			},
			list(repo, rev, query) {
				return this.content[defsListKeyFor(repo, rev, query)] || null;
			},
		});
		this.highlightedDef = null;
		this.refs = deepFreeze({
			content: data && data.refs ? data.refs.content : {},
			get(repo: string, rev: ?string, def: string, file: ?string) {
				return this.content[refsKeyFor(repo, rev, def, file)] || null;
			},
		});
	}

	toJSON() {
		return {defs: this.defs, refs: this.refs};
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DefActions.DefFetched:
			this.defs = deepFreeze(Object.assign({}, this.defs, {
				content: Object.assign({}, this.defs.content, {
					[defKey(action.repo, action.rev, action.def)]: action.defObj,
				}),
			}));
			break;

		case DefActions.DefsFetched:
			{
				// Store the list of defs AND each def individually so we can
				// perform more operations quickly.
				let data = {
					[defsListKeyFor(action.repo, action.rev, action.query)]: action.defs,
				};
				if (action.defs && action.defs.Defs) {
					action.defs.Defs.forEach((d) => {
						data[defKey(d.Repo, action.rev, defPath(d))] = d;
					});
				}
				this.defs = deepFreeze(Object.assign({}, this.defs, {
					content: Object.assign({}, this.defs.content, data),
				}));
				break;
			}

		case DefActions.HighlightDef:
			this.highlightedDef = action.url;
			break;

		case DefActions.RefsFetched:
			this.refs = deepFreeze(Object.assign({}, this.refs, {
				content: Object.assign({}, this.refs.content, {
					[refsKeyFor(action.repo, action.rev, action.def, action.file)]: action.refs,
				}),
			}));
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new DefStore(Dispatcher.Stores);
