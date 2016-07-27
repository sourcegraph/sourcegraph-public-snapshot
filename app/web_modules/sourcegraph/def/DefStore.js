import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DefActions from "sourcegraph/def/DefActions";
import {defPath} from "sourcegraph/def";
import type {Def, ExamplesKey, RefLocationsKey} from "sourcegraph/def";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import "sourcegraph/def/DefBackend";
import {fastParseDefPath} from "sourcegraph/def";
import toQuery from "sourcegraph/util/toQuery";
import update from "react/lib/update";
import type {BlobPos} from "sourcegraph/def/DefActions";

function defKey(repo: string, rev: ?string, def: string): string {
	return `${repo}#${rev || ""}#${def}`;
}

function defsListKeyFor(repo: string, rev: string, query: string, filePathPrefix: ?string): string {
	return `${repo}#${rev}#${query}#${filePathPrefix || ""}`;
}

function refsKeyFor(repo: string, rev: ?string, def: string, refRepo: string, refFile: ?string): string {
	return `${defKey(repo, rev, def)}#${refRepo}#${refFile || ""}`;
}

function refLocationsKeyFor(r: RefLocationsKey): string {
	let opts: {Repos?: Array<string>, Page?: number, PerPage?: number, Sorting?: string} = {
		Repos: r.repos,
	};
	if (r.page) {
		opts.Page = r.page;
	}
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

export class DefStore extends Store {
	_refLocations: Object;
	_examples: Object;

	getRefLocations(r: RefLocationsKey): ?Object {
		return this._refLocations[refLocationsKeyFor(r)] || null;
	}

	getExamples(r: ExamplesKey): ?Object {
		return this._examples[examplesKeyFor(r)] || null;
	}

	reset(data?: {defs: any, refs: any}) {
		this.defs = deepFreeze({
			content: data && data.defs ? data.defs.content : {},
			pos: data && data.defs ? data.defs.pos : {},
			get(repo: string, rev: ?string, def: string): ?Def {
				return this.content[defKey(repo, rev, def)] || null;
			},

			// getPos returns just the DefPos that the def is defined in. It
			// is an optimization over get because sometimes we cheaply can determine
			// just the def's pos (from annotations, for example), which is all we need
			// to support within-the-same-file jump-to-def without loading the full def.
			getPos(repo: string, rev: ?string, def: string): ?DefPos {
				// Prefer fetching from the def, which has the full def's start and end bytes, etc.
				const d = this.get(repo, rev, def);
				if (d && !d.Error) return d;
				return this.pos[defKey(repo, rev, def)] || null;
			},

			list(repo: string, commitID: string, query: string, filePathPrefix: ?string) {
				return this.content[defsListKeyFor(repo, commitID, query, filePathPrefix)] || null;
			},
		});
		this.authors = deepFreeze({
			content: data && data.authors ? data.authors.content : {},
			get(repo: string, commitID: string, def: string): ?Object {
				return this.content[defKey(repo, commitID, def)] || null;
			},
		});
		this.hoverPos = null;
		this.hoverInfos = deepFreeze({
			content: data && data.hoverInfos ? data.hoverInfos.content : {},
			get(pos: BlobPos): string {
				return this.content[posKeyFor(pos)] || null;
			},
		});
		this.refs = deepFreeze({
			content: data && data.refs ? data.refs.content : {},
			get(repo: string, commitID: string, def: string, refRepo: string, refFile: ?string) {
				return this.content[refsKeyFor(repo, commitID, def, refRepo, refFile)] || null;
			},
		});
		this._refLocations = deepFreeze(data && data.refLocations ? data.refLocations.content : {});
		this._examples = deepFreeze(data && data.examples ? data.examples.content : {});
	}

	toJSON() {
		return {
			defs: this.defs,
			authors: this.authors,
			refs: this.refs,
			refLocations: this.refLocations,
		};
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
					this.defs = deepFreeze({
						...this.defs,
						pos: {...this.defs.pos, ...defPos},
					});
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

		case DefActions.RefLocationsFetched:
			{
				let a = (action: DefActions.RefLocationsFetched);
				let r = a.request.resource;
				let updatedContent = {};
				updatedContent[refLocationsKeyFor(r)] = a.locations;

				// We need to support querying for refs without pagination options.
				// Below we merge contents from any previous fetches with the latest
				// fetch and store that as a new entry omitting pagination options.
				//
				// TODO this is a hack to support streaming pagination for performance
				// reasons. This should be temporary, or we need to implement a
				// properly formatted endpoint for this instead.
				let r2 = Object.assign({}, r);
				delete r2.page;
				// Keep track of the position of the pages we've loaded to prevent
				// saving the same set of refs twice.
				let page = r.page || 0;
				let refsForPage = this.getRefLocations(r);
				let currentRefs = this.getRefLocations(r2);
				let newRepoRefs = a.locations.RepoRefs;
				if (!refsForPage && currentRefs && currentRefs.PagesFetched < page) {
					currentRefs = update(currentRefs, {PagesFetched: {$set: page}});
					if (currentRefs.RepoRefs && newRepoRefs) {
						// If any of our new refs come from the same repo as the last ref
						// in our current list, append those refs to the end of the list.
						let lastRefIdx = currentRefs.RepoRefs.length - 1;
						let lastRef = currentRefs.RepoRefs[lastRefIdx];
						for (let ref of newRepoRefs) {
							if (ref.Repo === lastRef.Repo) {
								let mergedRefs = update(lastRef, {
									Count: {$set: lastRef.Count + ref.Count},
									Files: {$set: lastRef.Files.concat(ref.Files)},
								});
								currentRefs = update(currentRefs, {
									RepoRefs: {[lastRefIdx]: {$set: mergedRefs}},
								});
							} else {
								currentRefs = update(currentRefs, {
									RepoRefs: {$push: [ref]},
								});
							}
						}
					}
					// If no additional refs were fetched, communicate that the last page has been reached.
					if (!a.locations.RepoRefs) {
						currentRefs = update(currentRefs, {StreamTerminated: {$set: true}});
					}
					updatedContent[refLocationsKeyFor(r2)] = currentRefs;
				} else {
					updatedContent[refLocationsKeyFor(r2)] = Object.assign({}, a.locations, {PagesFetched: page});
				}

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

export default (new DefStore(Dispatcher.Stores): DefStore);
