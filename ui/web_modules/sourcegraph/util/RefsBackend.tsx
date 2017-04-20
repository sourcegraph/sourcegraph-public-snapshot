import * as zipObject from "lodash/zipObject";
import * as hash from "object-hash";
import { Observable, Subject, Subscriber } from "rxjs";

import URI from "vs/base/common/uri";
import { Position } from "vs/editor/common/core/position";
import { IPosition, IRange, IReadOnlyModel } from "vs/editor/common/editorCommon";
import { IReferenceInformation, LanguageIdentifier, Location, WorkspaceReferenceProviderRegistry } from "vs/editor/common/modes";
import { getDefinitionsAtPosition } from "vs/editor/contrib/goToDeclaration/common/goToDeclaration";
import { getHover } from "vs/editor/contrib/hover/common/hover";
import { provideReferences as getReferences, provideWorkspaceReferences } from "vs/editor/contrib/referenceSearch/common/referenceSearch";

import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { URIUtils } from "sourcegraph/core/uri";
import { setupWorker } from "sourcegraph/ext/main";
import { timeFromNowUntil } from "sourcegraph/util/dateFormatterUtil";
import { Features } from "sourcegraph/util/features";
import { fetchGQL } from "sourcegraph/util/gqlClient";
import { getURIContext } from "sourcegraph/workbench/utils";

// TODO(john): consolidate / standardize types.

export interface DefinitionData {
	definition?: {
		uri: URI;
		range: IRange;
	};
	docString: string;
	funcName: string;
}

export interface ReferenceCommitInfo {
	hunk: GQL.IHunk;
}

export async function provideDefinition(model: IReadOnlyModel, pos: IPosition): Promise<DefinitionData | null> {
	const position = new Position(pos.lineNumber, pos.column);
	const hoversPromise = getHover(model, position);
	const defPromise = getDefinitionsAtPosition(model, position);

	const [hovers, def] = await Promise.all([hoversPromise, defPromise]);

	if (!hovers || hovers.length === 0 || !hovers[0].contents || !def || !def[0]) {
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

export function provideReferencesStreaming(model: IReadOnlyModel, pos: IPosition): Observable<Location> {
	return new Observable<Location>(observer => {
		const handler = createProgressHandler(observer);
		getReferences(model, Position.lift(pos), handler, { includeDeclaration: false }).then(locations => {
			handler(locations.map(loc => {
				loc.uri = URIUtils.tryConvertGitToFileURI(loc.uri);
				return loc;
			}));
			observer.complete();
		}, err => observer.error(err));
	})
		// TODO(nick): This is a workaround for https://github.com/sourcegraph/sourcegraph/issues/4594
		// Once that issue is fixed, we should not need to dedupe.
		// Instead, we should discard the final result if any progress results were sent.
		.distinct(getLocationId)
		.take(MAX_REFS_PER_REPO);
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
	const shouldBlame = (reference: Location, repo: string): boolean => {
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
		const refUri = URIUtils.tryConvertGitToFileURI(reference.uri);
		const { repo, path } = getURIContext(refUri);
		if (!shouldBlame(reference, repo)) { return; }
		refModelQuery = refModelQuery + // TODO(john): name this query
			`${refKey(reference)}:repository(uri: "${repo}") {
				commit(rev: "${reference.uri.query}") {
					commit {
						file(path: "${path}") {
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

	const resp = await fetchGQL(query);
	const root = resp.data.root;
	if (!root) {
		return references;
	}

	return references.map(reference => {
		let dataByRefID = root[refKey(reference)];
		if (!dataByRefID || !dataByRefID.commit) {
			return reference; // likely means the blame was skipped by shouldBlame; continue without it
		}
		let blame = dataByRefID.commit.commit.file.blame;
		if (!blame || !blame[0]) {
			return reference;
		}
		let hunk: GQL.IHunk = Object.assign({}, blame[0]); // allow mutation
		if (!hunk.author || !hunk.author.person) {
			return reference;
		}
		hunk.author = Object.assign({}, hunk.author); // allow mutation
		hunk.author.date = timeFromNowUntil(hunk.author.date, 14);
		return { ...reference, commitInfo: { hunk } };
	});
}

const MAX_GLOBAL_REFS_REPOS = 5;

/**
 * Limits the number of references displayed for any given repo.
 * This reduces the number of graphql queries to resolve the previews and
 * get the user to a "done" state with no loading spinner quicker.
 */
const MAX_REFS_PER_REPO = 100;
const globalRefsObservablesStreaming = new Map<string, Observable<LocationWithCommitInfo>>();

function log(message?: any, ...optionalParams: any[]): void {
	if (Features.refLogs.isEnabled()) {
		// tslint:disable: no-console
		console.log(message, ...optionalParams);
	}
}

export function provideGlobalReferencesStreaming(model: IReadOnlyModel, pos: IPosition): Observable<Location> {
	return Observable.from(fetchDependencyReferences(model, pos)).flatMap(depRefs => {
		if (!depRefs) {
			return Observable.empty();
		}

		const repos = depRefs.repoData!.repos;
		const repoIDs = depRefs.repoData!.repoIds;
		if (!repos || !repoIDs) {
			return Observable.empty();
		}

		const repoMap = zipObject(repoIDs, repos);
		const symbol = depRefs.dependencyReferenceData!.location!.symbol!;
		const key = hash(symbol); // key is the encoded data about the symbol, which will be the same across different locations of the symbol
		const cacheHit = globalRefsObservablesStreaming.get(key);
		if (cacheHit) {
			log(`cache hit for provideGlobalReferencesStreaming`, depRefs);
			// We may have cached global refs from inside a different workspace context.
			// Always filter "global" refs from the current repo context (since they aren't global)
			// in this context.
			//
			// TODO(john): there's another issue here, if you are in repo A and ask for global refs
			// from A, then the cached observable won't include results from repo A. If you navigate
			// to repo B and ask for global refs for the same symbol as from repo A, you will no longer
			// see A in the global refs list. It might be better to just avoid cacheing altogether, or instead
			// cache an observable per repo and construct a new aggregate observable from these possibly
			// cached streams.
			return cacheHit.filter(loc => getURIContext(loc.uri).repo !== getURIContext(model.uri).repo);
		}

		// triggerRequest.next() causes a workspace/xreferences request
		// to be made for the next available repo.
		const triggerRequest = new Subject<void>();

		// Start with a number of requests (in parallel) that is equal to
		// the number of repos that we want results from.
		const initialRequests = triggerRequest.merge(Observable.range(1, MAX_GLOBAL_REFS_REPOS));

		let foundRepos = 0;

		// When a repo returns results, we will process those results.
		// When a repo does not have results, we will trigger another request (until we run out).
		const result = Observable.from(depRefs.dependencyReferenceData!.references)
			// Throttle the emission of references by calls to .next() on triggerRequest.
			.zip(initialRequests, ref => ref)

			// Merge the observables from each request into a single observable.
			.flatMap(depRef => {
				const depRepo = repoMap[depRef.repoId];
				let hasResults = false;

				// Make the actual request to get global references.
				log(`starting request ${depRepo.uri}`);
				return fetchGlobalReferencesForDependentRepoStreaming(depRepo, model.getLanguageIdentifier(), JSON.parse(symbol), depRef)
					.take(MAX_REFS_PER_REPO)
					.do(location => {
						log(`progress ${depRepo.uri}`);
						// If this repo has results, we don't want to trigger another request after it finishes.
						hasResults = true;
					})
					.finally(() => {
						log(`finished ${depRepo.uri} hasResults=${hasResults}`);
						if (!hasResults) {
							// If this repo didn't have any global references, we want to trigger another request.
							triggerRequest.next();
						} else {
							foundRepos++;
							if (foundRepos === MAX_GLOBAL_REFS_REPOS) {
								triggerRequest.complete();
							}
						}
					});
			})

			// Allow future subscribers to observe a relay of all events.
			// This is necessary for caching to work.
			.publishReplay()

			// Automatically start the observable when someone subscribes.
			.refCount();

		globalRefsObservablesStreaming.set(key, result);
		return result;
	});
}

function fetchGlobalReferencesForDependentRepoStreaming(repo: GQL.IRepository, language: LanguageIdentifier, symbol: any, reference: GQL.IDependencyReference): Observable<Location> {
	if (!repo.uri || !repo.lastIndexedRevOrLatest) {
		return Observable.empty();
	}

	let repoURI = URIUtils.createResourceURI(repo.uri);

	// Setting up a workspace is async and there is currently no way for
	// the main thread to know when the extension is finished registering.
	// Since we are spinning up this workspace specifically to make a workspace
	// references request, we just poll the workspace until it has registered a handler.
	const workspaceIsReady = () => {
		return WorkspaceReferenceProviderRegistry.has({
			isTooLargeForHavingARichMode(): boolean {
				return false;
			},
			getLanguageIdentifier(): LanguageIdentifier {
				return language;
			},
			getModeId(): string {
				return language.language;
			},
			uri: repoURI,
		} as IReadOnlyModel);
	};

	return new Observable<IReferenceInformation>(observer => {
		setupWorkspace(repoURI, repo.lastIndexedRevOrLatest.commit!.sha1, workspaceIsReady).then(() => { // TODO(john): rev argument is incorrect
			const handler = createProgressHandler(observer);
			provideWorkspaceReferences(language, repoURI, symbol, JSON.parse(reference.hints), handler).then(references => {
				handler(references.map(ref => {
					ref.reference.uri = URIUtils.tryConvertGitToFileURI(ref.reference.uri);
					return ref;
				}));
				observer.complete();
			}, err => observer.error(err));
		});
	})
		.map(ref => ref.reference)
		// TODO(nick): This is a workaround for https://github.com/sourcegraph/sourcegraph/issues/4594
		// Once that issue is fixed, we should not need to dedupe.
		// Instead, we should discard the final result if any progress results were sent.
		.distinct(getLocationId);
}

/**
 * getLocationId returns a unique identifier for the given location.
 */
function getLocationId(location: Location): string {
	return [
		location.uri.toString(),
		location.range.startColumn,
		location.range.startLineNumber,
		location.range.endColumn,
		location.range.endLineNumber
	].join(":");
}

/**
 * createProgressHandler returns a function that forwards value to
 * the provided observer when it is called.
 */
function createProgressHandler<T>(observer: Subscriber<T>): (values: T[]) => void {
	return values => {
		for (const v of values) {
			observer.next(v);
		}
	};
}

function setupWorkspace(uri: URI, rev: string | undefined, isReady: () => boolean): Promise<void> {
	log(`setup workspace ${uri}`);
	setupWorker({ resource: uri, revState: { commitID: rev } });
	return new Promise<void>(resolve => {
		const interval = setInterval(() => {
			if (isReady()) {
				log(`workspace ready ${uri}`);
				clearInterval(interval);
				resolve();
			}
		}, 100);
	});
}

const globalRefLangs = new Set(["go", "java", "typescript", "javascript"]);
async function fetchDependencyReferences(model: IReadOnlyModel, pos: IPosition): Promise<GQL.IDependencyReferences | null> {
	// Only fetch global refs for certain languages.
	if (!globalRefLangs.has(model.getModeId())) {
		return null;
	}

	const { repo, rev, path } = getURIContext(model.uri);
	let refModelQuery =
		`repository(uri: "${repo}") {
			commit(rev: "${rev}") {
				commit {
					file(path: "${path}") {
						dependencyReferences(Language: "${model.getModeId()}", Line: ${pos.lineNumber - 1}, Character: ${pos.column - 1}) {
							dependencyReferenceData {
								references {
									dependencyData
									repoId
									hints
								}
								location {
									location
									symbol
								}
							}
							repoData {
								repos {
									id
									uri
									lastIndexedRevOrLatest {
										commit {
											sha1
										}
									}
								}
								repoIds
							}
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

	const resp = await fetchGQL(query);
	const root = resp.data.root;
	if (!root.repository || !root.repository.commit.commit || !root.repository.commit.commit.file ||
		!root.repository.commit.commit.file.dependencyReferences.repoData || !root.repository.commit.commit.file.dependencyReferences.dependencyReferenceData || !root.repository.commit.commit.file.dependencyReferences.dependencyReferenceData.references.length) {
		return null;
	}
	return root.repository.commit.commit.file.dependencyReferences;
}
