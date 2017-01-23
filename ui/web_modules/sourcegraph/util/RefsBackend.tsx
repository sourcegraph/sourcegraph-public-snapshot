import * as hash from "object-hash";
import { Observable } from "rxjs";
import { Definition, Hover, Position } from "vscode-languageserver-types";

import { IRange, IReadOnlyModel } from "vs/editor/common/editorCommon";
import { Location } from "vs/editor/common/modes";

import { Repo } from "sourcegraph/api/index";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { URIUtils } from "sourcegraph/core/uri";
import * as lsp from "sourcegraph/editor/lsp";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { timeFromNowUntil } from "sourcegraph/util/dateFormatterUtil";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";
import { OneReference, ReferencesModel } from "sourcegraph/workbench/info/referencesModel";

import * as _ from "lodash";

export interface RefData {
	language: string;
	repo: string;
	version: string;
	file: string;
	line: number;
	column: number;
}

export interface DefinitionData {
	definition: {
		uri: string;
		range: IRange;
	};
	docString: string;
	funcName: string;
}

export interface DepRefsData {
	Data: {
		Location: {
			location: lsp.Location,
			symbol: any,
		};
		References: [DepReference];
	};
	RepoData: { [id: number]: Repo };
}

interface DepReference {
	DepData: any;
	Hints: {
		dirs: [string];
	};
	RepoID: number;
}

export interface ReferenceCommitInfo {
	hunk: GQL.IHunk;
}

function cacheKey(model: IReadOnlyModel, pos: Position): string {
	const word = model.getWordAtPosition(RangeOrPosition.fromLSPPosition(pos).toMonacoPosition());
	return `${model.uri.toString(true)}:${pos.line}:${word.startColumn}:${word.endColumn}`;
}

const hoverCache = new Map<string, Promise<Hover>>();
const defCache = new Map<string, Promise<Definition>>();
const referencesCache = new Map<string, Promise<Location[]>>();

export async function provideDefinition(model: IReadOnlyModel, pos: Position): Promise<DefinitionData | null> {
	const key = cacheKey(model, pos);

	const hoverCacheHit = hoverCache.get(key);
	const hoverPromise = hoverCacheHit ? hoverCacheHit : lsp.send(model, "textDocument/hover", {
		textDocument: { uri: model.uri.toString(true) },
		position: pos,
		context: { includeDeclaration: false },
	}).then(resp => {
		if (resp.result) {
			hoverCache.set(key, hoverPromise);
			return resp.result;
		}
	});

	const defCacheHit = defCache.get(key);
	const defPromise = defCacheHit ? defCacheHit : lsp.send(model, "textDocument/definition", {
		textDocument: { uri: model.uri.toString(true) },
		position: pos,
		context: { includeDeclaration: false },
	}).then(resp => {
		if (resp && resp.result) {
			defCache.set(key, defPromise);
			return resp.result as Definition;
		}
	});

	const hover = await hoverPromise;
	const def = await defPromise;

	if (!hover || !hover.contents || !def || !def[0]) { // TODO(john): throw the error in `lsp.send`, then do try/catch around await.
		return null;
	}

	let docString: string;
	let funcName: string;
	if (hover.contents instanceof Array) {
		const [first, second] = hover.contents;
		// TODO(nico): this shouldn't be detrmined by position, but language of the content (e.g. 'text/markdown' for doc string)
		funcName = first instanceof String ? first : first.value;
		docString = second ? (second instanceof String ? second : second.value) : "";
	} else {
		funcName = hover.contents instanceof String ? hover.contents : hover.contents.value;
		docString = "";
	}

	const firstDef = def[0]; // TODO: handle disambiguating multiple declarations
	let definition = {
		uri: firstDef.uri,
		range: {
			startLineNumber: firstDef.range.start.line,
			startColumn: firstDef.range.start.character,
			endLineNumber: firstDef.range.end.line,
			endColumn: firstDef.range.end.character,
		}
	};
	return { funcName, docString, definition };
}

export async function provideReferences(model: IReadOnlyModel, pos: Position): Promise<Location[]> {
	const key = cacheKey(model, pos);

	const referencesCacheHit = referencesCache.get(key);
	const referencesPromise = referencesCacheHit ? referencesCacheHit : lsp.send(model, "textDocument/references", {
		textDocument: { uri: model.uri.toString(true) },
		position: pos,
		context: { includeDeclaration: false },
	})
		.then(resp => resp ? resp.result : null)
		.then((resp: lsp.Location | lsp.Location[] | null) => {
			if (!resp || Object.keys(resp).length === 0) {
				return [];
			}
			referencesCache.set(key, referencesPromise);

			const { repo, rev, path } = URIUtils.repoParams(model.uri);
			AnalyticsConstants.Events.CodeReferences_Viewed.logEvent({ repo, rev: rev || "", path });

			const locs: lsp.Location[] = resp instanceof Array ? resp : [resp];
			return locs.map(lsp.toMonacoLocation);
		});

	return referencesPromise;
}

export async function provideReferencesCommitInfo(referencesModel: ReferencesModel): Promise<ReferencesModel> {
	// Blame is slow, so only blame the first N references in the first N repos.
	//
	// These parameters were chosen arbitrarily.
	const maxReposToBlame = 6;
	const maxReferencesToBlamePerRepo = 4;
	const blameQuota = new Map<string, number>();
	const shouldBlame = (reference: OneReference): boolean => {
		const repo = `${reference.uri.authority}${reference.uri.path}`;
		let quotaRemaining = blameQuota.get(repo);
		if (quotaRemaining === undefined) {
			if (blameQuota.size === maxReposToBlame) { return false; }
			quotaRemaining = maxReferencesToBlamePerRepo;
		}
		if (quotaRemaining === 0) { return false; }
		blameQuota.set(repo, quotaRemaining - 1);
		return true;
	};

	let refModelQuery: string = "";
	referencesModel.references.forEach(reference => {
		if (!shouldBlame(reference)) { return; }
		refModelQuery = refModelQuery +
			`${reference.id.replace("#", "")}:repository(uri: "${reference.uri.authority}${reference.uri.path}") {
				commit(rev: "${reference.uri.query}") {
					commit {
						file(path: "${reference.uri.fragment}") {
							blame(startLine: ${reference.range.startLineNumber}, endLine: ${reference.range.endLineNumber}) {
								rev
								startLine
								endLine
								startByte
								endByte
								message
								author {
									person {
										gravatarHash
										name
										email
									}
									date
								}
							}
		  				}
					}
				}
			}`;
	});

	const query =
		`query {
			root {
				${refModelQuery}
			}
		}`;

	let data = await fetchGraphQLQuery(query);
	if (!data.root) {
		return referencesModel;
	}

	referencesModel.references.forEach(reference => {
		let dataByRefID = data.root[reference.id.replace("#", "")];
		if (!dataByRefID) {
			return; // likely means the blame was skipped by shouldBlame; continue without it
		}
		let hunk: GQL.IHunk = dataByRefID.commit.commit.file.blame[0];
		if (!hunk || !hunk.author || !hunk.author.person) {
			return;
		}
		hunk.author.date = timeFromNowUntil(hunk.author.date, 14);
		reference.commitInfo = { hunk };
	});

	return referencesModel;
}

const MAX_GLOBAL_REFS_REPOS = 5;
const globalRefsObservables = new Map<string, Observable<Location[]>>();

export function provideGlobalReferences(model: IReadOnlyModel, depRefs: DepRefsData): Observable<Location[]> {
	const dependents = depRefs.Data.References;
	const repoData = depRefs.RepoData;
	const symbol = depRefs.Data.Location.symbol;
	const modeID = model.getModeId();

	const key = hash(symbol); // key is the encoded data about the symbol, which will be the same across different locations of the symbol
	const cacheHit = globalRefsObservables.get(key);
	if (cacheHit) {
		return cacheHit;
	}

	const observable = new Observable<Location[]>(observer => {
		let interval: number | null = null;
		let unsubscribed = false;
		let completed = false;

		let countNonempty = 0;
		function incrementCountIfNonempty(refs: Location[]): void {
			if (refs && refs.length > 0) {
				++countNonempty;
			}
		}

		function complete(): void {
			if (!completed) {
				completed = true;
			}
			observer.complete();
		}

		function pollNext(count: number): Promise<void>[] {
			if (dependents.length > 0 && countNonempty < MAX_GLOBAL_REFS_REPOS) {
				const isLastDependent = dependents.length === 1;
				const nextDependentsToPoll = dependents.splice(0, Math.min(count, dependents.length));
				return nextDependentsToPoll.map(dependent => {
					const repo = repoData[dependent.RepoID];
					return fetchGlobalReferencesForDependentRepo(dependent, repo, modeID, symbol).then(refs => {
						incrementCountIfNonempty(refs);
						if (!completed && refs.length > 0) {
							observer.next(refs);
						}
						if (isLastDependent || countNonempty === MAX_GLOBAL_REFS_REPOS) {
							complete();
						}
					});
				});
			}

			return [];
		}

		// Fetch references from the first two dependent repos right away, in parallel.
		Promise.all(pollNext(2)).then(() => {
			// Unsubscription may happen before we complete fetching the first 2 dependent repo refs.
			// If so, bail before starting the interval poll.
			if (unsubscribed) {
				return;
			}
			if (countNonempty >= MAX_GLOBAL_REFS_REPOS || dependents.length === 0) {
				complete();
			}
			// Every 2.5 seconds following, fetch refs for another dependent repo until
			// there are no more dependent repos or we've fetched non-empty results from MAX_GLOBAL_REFS_REPOS.
			interval = setInterval(() => pollNext(1), 2500);
		});

		return function unsubscribe(): void {
			unsubscribed = true;
			if (interval) {
				clearInterval(interval);
			}
		};
	}).publishReplay(MAX_GLOBAL_REFS_REPOS).refCount();

	globalRefsObservables.set(key, observable);
	return observable;
}

function fetchGlobalReferencesForDependentRepo(reference: DepReference, repo: Repo, modeID: string, symbol: any): Promise<Location[]> {
	if (!repo.URI || !repo.DefaultBranch) {
		return Promise.resolve([]);
	}

	let repoURI = URIUtils.pathInRepo(repo.URI, repo.DefaultBranch, "");
	return lsp.sendExt(repoURI.toString(), modeID, "workspace/xreferences", {
		query: symbol,
		hints: reference.Hints,
	}).then(resp => !resp.result ? [] : resp.result.map(ref => lsp.toMonacoLocation(ref.reference)));
}

export async function fetchDependencyReferences(model: IReadOnlyModel, pos: Position): Promise<DepRefsData | null> {
	let refModelQuery =
		`repository(uri: "${model.uri.authority}${model.uri.path}") {
			commit(rev: "${model.uri.query}") {
				commit {
					file(path: "${model.uri.fragment}") {
						dependencyReferences(Language: "${model.getModeId()}", Line: ${pos.line}, Character: ${pos.character}) {
							data
						}
					}
				}
			}
		}`;

	const query =
		`query {
			root {
				${refModelQuery}
			}
		}`;

	let data = await fetchGraphQLQuery(query);
	if (!data.root.repository || !data.root.repository.commit.commit || !data.root.repository.commit.commit.file ||
		!data.root.repository.commit.commit.file.dependencyReferences || !data.root.repository.commit.commit.file.dependencyReferences.data.length) {
		return null;
	}
	let object = JSON.parse(data.root.repository.commit.commit.file.dependencyReferences.data);
	if (!object.RepoData || !object.Data || !object.Data.References) {
		return null;
	}
	let repos = _.values(object.RepoData);

	let currentRepo = _.remove(repos, function (repo: any): boolean {
		return (repo as any).URI === `${model.uri.authority}${model.uri.path}`;
	});
	if (!currentRepo || !currentRepo.length) {
		return object;
	}

	_.remove(object.Data.References, function (reference: any): boolean {
		return reference.RepoID === currentRepo[0].ID;
	});
	delete object.RepoData[currentRepo[0].ID];

	return object;
}
