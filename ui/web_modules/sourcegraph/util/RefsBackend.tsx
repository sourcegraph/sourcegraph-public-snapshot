import * as remove from "lodash/remove";
import * as values from "lodash/values";
import * as hash from "object-hash";
import { Observable } from "rxjs";

import URI from "vs/base/common/uri";
import { Position } from "vs/editor/common/core/position";
import { IPosition, IRange, IReadOnlyModel } from "vs/editor/common/editorCommon";
import { Location } from "vs/editor/common/modes";
import { getDeclarationsAtPosition } from "vs/editor/contrib/goToDeclaration/common/goToDeclaration";
import { getHover } from "vs/editor/contrib/hover/common/hover";
import { provideReferences as getReferences } from "vs/editor/contrib/referenceSearch/common/referenceSearch";

import { Repo } from "sourcegraph/api/index";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { URIUtils } from "sourcegraph/core/uri";
import * as lsp from "sourcegraph/editor/lsp";
import { setupWorker } from "sourcegraph/ext/main";
import { timeFromNowUntil } from "sourcegraph/util/dateFormatterUtil";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";

export interface DefinitionData {
	definition?: {
		uri: URI;
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

export async function provideDefinition(model: IReadOnlyModel, pos: IPosition): Promise<DefinitionData | null> {
	const position = new Position(pos.lineNumber, pos.column);
	const hoversPromise = getHover(model, position);
	const defPromise = getDeclarationsAtPosition(model, position);

	const [hovers, def] = await Promise.all([hoversPromise, defPromise]);

	if (!hovers || hovers.length === 0 || !hovers[0].contents || !def || !def[0]) { // TODO(john): throw the error in `lsp.send`, then do try/catch around await.
		return null;
	}

	const hover = hovers[0]; // TODO(john): support multiple hover tooltips

	let docString = "";
	let funcName = "";
	if (hover.contents.length > 1) {
		const [first, second] = hover.contents;
		if (first) {
			if (typeof first === "string") {
				funcName = first;
			} else if (first.value) {
				funcName = first.value;
			} else {
				return null;
			}
		}
		if (second) {
			if (typeof second === "string") {
				docString = second;
			} else if (second.value) {
				docString = second.value;
			}
		}
	} else {
		const content = hover.contents[0];
		funcName = typeof content === "string" ? content : content.value;
		docString = "";
	}

	const firstDef = def[0]; // TODO: handle disambiguating multiple declarations
	const definition = {
		uri: firstDef.uri,
		range: firstDef.range,
	};
	return { funcName, docString, definition };
}

export async function provideReferences(model: IReadOnlyModel, pos: IPosition, progress: (locations: Location[]) => void): Promise<Location[]> {
	return getReferences(model, Position.lift(pos), progress);
}

export interface LocationWithCommitInfo extends Location {
	commitInfo?: { hunk: GQL.IHunk };
}

export async function provideReferencesCommitInfo(references: Location[]): Promise<LocationWithCommitInfo[]> {
	// Blame is slow, so only blame the first N references in the first N repos.
	//
	// These parameters were chosen arbitrarily.
	const maxReposToBlame = 6;
	const maxReferencesToBlamePerRepo = 4;
	const blameQuota = new Map<string, number>();
	const shouldBlame = (reference: Location): boolean => {
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

	function refKey(reference: Location): string {
		return `${reference.uri.toString()}:${RangeOrPosition.fromMonacoRange(reference.range).toString()}`.replace(/\W/g, "_"); // graphql keys must be alphanumeric
	}

	let refModelQuery = "";
	references.forEach(reference => {
		if (!shouldBlame(reference)) { return; }
		refModelQuery = refModelQuery +
			`${refKey(reference)}:repository(uri: "${reference.uri.authority}${reference.uri.path}") {
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
		return references;
	}

	return references.map(reference => {
		let dataByRefID = data.root[refKey(reference)];
		if (!dataByRefID || !dataByRefID.commit) {
			return reference; // likely means the blame was skipped by shouldBlame; continue without it
		}
		let blame = dataByRefID.commit.commit.file.blame;
		if (!blame || !blame[0]) {
			return reference;
		}
		let hunk: GQL.IHunk = blame[0];
		if (!hunk.author || !hunk.author.person) {
			return reference;
		}
		hunk.author.date = timeFromNowUntil(hunk.author.date, 14);
		return { ...reference, commitInfo: { hunk } };
	});
}

const MAX_GLOBAL_REFS_REPOS = 5;
const globalRefsObservables = new Map<string, Observable<LocationWithCommitInfo[]>>();

export function provideGlobalReferences(model: IReadOnlyModel, depRefs: DepRefsData): Observable<LocationWithCommitInfo[]> {
	const dependents = depRefs.Data.References;
	const repoData = depRefs.RepoData;
	const symbol = depRefs.Data.Location.symbol;
	const modeID = model.getModeId();

	const key = hash(symbol); // key is the encoded data about the symbol, which will be the same across different locations of the symbol
	const cacheHit = globalRefsObservables.get(key);
	if (cacheHit) {
		return cacheHit;
	}

	const observable = new Observable<LocationWithCommitInfo[]>(observer => {
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
					return fetchGlobalReferencesForDependentRepo(dependent, repo, modeID, symbol).then(provideReferencesCommitInfo).then(refs => {
						incrementCountIfNonempty(refs);
						if (!completed && refs.length > 0) {
							// HACK(john): this should be written into vscode internals (https://github.com/sourcegraph/sourcegraph/issues/3998)
							// The first URI in the list is assumed to share the same workspace with the rest.
							setupWorker(refs[0].uri.with({ fragment: "" }));
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
	if (!repo.URI || !repo.IndexedRevision) {
		return Promise.resolve([]);
	}

	let repoURI = URIUtils.pathInRepo(repo.URI, repo.IndexedRevision, "");
	return lsp.sendExt(repoURI.toString(), modeID, "workspace/xreferences", {
		query: symbol,
		hints: reference.Hints,
	}).then(resp => (!resp || !resp.result) ? [] : resp.result.map(ref => lsp.toMonacoLocation(ref.reference)));
}

const globalRefLangs = new Set(["go", "java"]);
export async function fetchDependencyReferences(model: IReadOnlyModel, pos: IPosition): Promise<DepRefsData | null> {
	// Only fetch global refs for certain languages.
	if (!globalRefLangs.has(model.getModeId())) {
		return null;
	}

	let refModelQuery =
		`repository(uri: "${model.uri.authority}${model.uri.path}") {
			commit(rev: "${model.uri.query}") {
				commit {
					file(path: "${model.uri.fragment}") {
						dependencyReferences(Language: "${model.getModeId()}", Line: ${pos.lineNumber - 1}, Character: ${pos.column - 1}) {
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
	if (!object.RepoData || !object.Data || !object.Data.References || object.Data.References.length === 0) {
		return null;
	}
	let repos = values(object.RepoData);

	let currentRepo = remove(repos, function (repo: any): boolean {
		return (repo as any).URI === `${model.uri.authority}${model.uri.path}`;
	});
	if (!currentRepo || !currentRepo.length) {
		return object;
	}

	remove(object.Data.References, function (reference: any): boolean {
		return reference.RepoID === currentRepo[0].ID;
	});
	delete object.RepoData[currentRepo[0].ID];

	return object;
}
