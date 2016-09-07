// tslint:disable: typedef ordered-imports

import {Def} from "sourcegraph/api";
import {Store} from "sourcegraph/Store";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {deepFreeze} from "sourcegraph/util/deepFreeze";
import * as DefActions from "sourcegraph/def/DefActions";
import {defPath} from "sourcegraph/def";
import {ExamplesKey, RefLocationsKey} from "sourcegraph/def";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import "sourcegraph/def/DefBackend";
import {fastParseDefPath} from "sourcegraph/def";
import {toQuery} from "sourcegraph/util/toQuery";
import {BlobPos} from "sourcegraph/def/DefActions";

function defKey(repo: string, rev: string | null, def: string): string {
	return `${repo}#${rev || ""}#${def}`;
}

function defsListKeyFor(repo: string, rev: string, query: string, filePathPrefix: string | null): string {
	return `${repo}#${rev}#${query}#${filePathPrefix || ""}`;
}

function refsKeyFor(repo: string, rev: string | null, def: string, refRepo: string, refFile: string | null): string {
	return `${defKey(repo, rev, def)}#${refRepo}#${refFile || ""}`;
}

function refLocationsKeyFor(r: RefLocationsKey): string {
	let opts: {Repos?: string[], Page?: number, PerPage?: number, Sorting?: string} = {
		Repos: r.repos,
	};
	let q = toQuery(opts);
	return `/.api/repos/${r.repo}${r.rev ? `@${r.rev}` : ""}/-/def/${r.def}/-/ref-locations?${q}`;
}

function examplesKeyFor(r: ExamplesKey): string {
	return (new DefActions.WantExamples(r)).url();
}

function posKeyFor(pos: BlobPos) {
	return `${pos.repo}#${pos.commit}#${pos.file}#${pos.line}#${pos.character}`;
}

type DefPos = {
	// Mirrors the fields of the same name in Def so that if the whole
	// Def is available, we can just use it as its own DefPos.
	File: string;
	DefStart: number;
	DefEnd: number;
};

class DefStoreClass extends Store<any> {
	authors: any;
	hoverPos: any;
	hoverInfos: any;
	refs: any;
	defs: any;

	_refLocations: any;
	_examples: any;

	getRefLocations(r: RefLocationsKey): any {
		return this._refLocations[refLocationsKeyFor(r)] || null;
	}

	getExamples(r: ExamplesKey): any {
		return this._examples[examplesKeyFor(r)] || null;
	}

	reset() {
		this.defs = deepFreeze({
			content: {},
			pos: {},
			get(repo: string, rev: string | null, def: string): Def | null {
				return this.content[defKey(repo, rev, def)] || null;
			},

			// getPos returns just the DefPos that the def is defined in. It
			// is an optimization over get because sometimes we cheaply can determine
			// just the def's pos (from annotations, for example), which is all we need
			// to support within-the-same-file jump-to-def without loading the full def.
			getPos(repo: string, rev: string | null, def: string): DefPos | null {
				// Prefer fetching from the def, which has the full def's start and end bytes, etc.
				const d = this.get(repo, rev, def);
				if (d && !d.Error) {
					return d;
				}
				return this.pos[defKey(repo, rev, def)] || null;
			},

			list(repo: string, commitID: string, query: string, filePathPrefix: string | null) {
				return this.content[defsListKeyFor(repo, commitID, query, filePathPrefix)] || null;
			},
		});
		this.authors = deepFreeze({
			content: {},
			get(repo: string, commitID: string, def: string): any {
				return this.content[defKey(repo, commitID, def)] || null;
			},
		});
		this.hoverPos = null;
		this.hoverInfos = deepFreeze({
			content: {},
			get(pos: BlobPos): string {
				return this.content[posKeyFor(pos)] || null;
			},
		});
		this.refs = deepFreeze({
			content: {},
			get(repo: string, commitID: string, def: string, refRepo: string, refFile: string | null) {
				return this.content[refsKeyFor(repo, commitID, def, refRepo, refFile)] || null;
			},
		});
		this._refLocations = deepFreeze({});
		this._examples = deepFreeze({});
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

		case DefActions.DefAuthorsFetched:
			this.authors = deepFreeze(Object.assign({}, this.authors, {
				content: Object.assign({}, this.authors.content, {
					[defKey(action.repo, action.commitID, action.def)]: action.authors,
				}),
			}));
			break;

		case DefActions.DefsFetched:
			{
				// Store the list of defs AND each def individually so we can
				// perform more operations quickly.
				let data = {
					[defsListKeyFor(action.repo, action.commitID, action.query, action.filePathPrefix)]: action.defs,
				};
				if (action.defs && action.defs.Defs) {
					action.defs.Defs.forEach((d) => {
						data[defKey(d.Repo, action.commitID, defPath(d))] = d;
					});
				}
				this.defs = deepFreeze(Object.assign({}, this.defs, {
					content: Object.assign({}, this.defs.content, data),
				}));
				break;
			}

		case BlobActions.AnnotationsFetched:
			{
				// For any ref annotations with Def=true, we know their defs are in
				// this file, so we can record that for faster within-same-file jump-to-def.
				if (action.annotations && action.annotations.Annotations) {
					// Needn't complete synchronously since this is an optimization,
					// and deepFreezing so much data actually can take ~1s on a ~1000
					// line file in dev mode, so run this in setTimeout.
					const defPos: {[key: string]: DefPos} = {};
					action.annotations.Annotations.forEach((ann) => {
						if (ann.Def && ann.URL) {
							// All of these defs must be defined in the current repo
							// and commitID (since that's what Def means), so we don't need to
							// call the slower def/index.js routeParams to parse out the
							// def path.
							const def = fastParseDefPath(ann.URL);
							if (def) {
								defPos[defKey(action.repo, action.commitID, def)] = {
									File: action.path,
									// This is just the range for the def's name, not the whole
									// def, but it's better than nothing. The whole def range
									// will be available when the full def loads. In the meantime
									// this lets BlobMain, for example, scroll to the def's name
									// in the file (which is better than not scrolling at all until
									// the full def loads).
									DefStart: ann.StartByte,
									DefEnd: ann.EndByte,
								};
							}
						}
					});
					this.defs = deepFreeze(Object.assign({}, this.defs, {
						pos: Object.assign({}, this.defs.pos, defPos),
					}));
				}
				break;
			}

		case DefActions.Hovering:
			this.hoverPos = action.pos;
			break;

		case DefActions.HoverInfoFetched:
			this.hoverInfos = deepFreeze(Object.assign({}, this.hoverInfos, {
				content: Object.assign({}, this.hoverInfos.content, {
					[posKeyFor(action.pos)]: action.info,
				}),
			}));
			break;

		case DefActions.ExamplesFetched:
			this._examples = deepFreeze(Object.assign({}, this._examples, {
				[examplesKeyFor(action.request.resource)]: action.locations,
			}));
			break;

		case DefActions.LocalRefLocationsFetched:
		case DefActions.RefLocationsFetched:
			{
				let a = (action as DefActions.RefLocationsFetched);
				let r = a.request.resource;
				let updatedContent = {};
				updatedContent[refLocationsKeyFor(r)] = a.locations;

				this._refLocations = deepFreeze(Object.assign({}, this._refLocations, updatedContent));
				break;
			}

		case DefActions.RefsFetched:
			this.refs = deepFreeze(Object.assign({}, this.refs, {
				content: Object.assign({}, this.refs.content, {
					[refsKeyFor(action.repo, action.commitID, action.def, action.refRepo, action.refFile)]: action.refs,
				}),
			}));
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export const DefStore = new DefStoreClass(Dispatcher.Stores);
