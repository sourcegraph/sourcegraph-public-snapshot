import * as flatten from "lodash/flatten";

import * as glob from "vs/base/common/glob";
import * as objects from "vs/base/common/objects";
import * as scorer from "vs/base/common/scorer";
import * as strings from "vs/base/common/strings";
import URI from "vs/base/common/uri";
import { PPromise, TPromise } from "vs/base/common/winjs.base";
import { IConfigurationService } from "vs/platform/configuration/common/configuration";
import { IEnvironmentService } from "vs/platform/environment/common/environment";
import { IFileStat } from "vs/platform/files/common/files";
import { FileMatch, IFileMatch, ISearchComplete, ISearchConfiguration, ISearchProgressItem, ISearchQuery, ISearchService, QueryType } from "vs/platform/search/common/search";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";

import { RepoList } from "sourcegraph/api";
import { context, isOnPremInstance } from "sourcegraph/app/context";
import { URIUtils } from "sourcegraph/core/uri";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";
import { defaultExcludesRegExp } from "sourcegraph/workbench/ConfigurationService";
import { getFilesCached } from "sourcegraph/workbench/overrides/fileService";

const reposCache = new Map<string, RepoList>();

const logSearchTiming = Boolean(window.localStorage.getItem("logSearchTiming"));

export class SearchService implements ISearchService {
	_serviceBrand: any;

	constructor(
		@IEnvironmentService environmentService: IEnvironmentService,
		@IWorkspaceContextService private contextService: IWorkspaceContextService,
		@IConfigurationService private configurationService: IConfigurationService,
	) { }

	/**
	 * extendQuery is copied from vscode's search service implementation.
	 */
	public extendQuery(query: ISearchQuery): void {
		const configuration = this.configurationService.getConfiguration<ISearchConfiguration>();

		// Configuration: Encoding
		if (!query.fileEncoding) {
			const fileEncoding = configuration && configuration.files && configuration.files.encoding;
			query.fileEncoding = fileEncoding;
		}

		// Configuration: File Excludes
		const fileExcludes = configuration && configuration.files && configuration.files.exclude;
		if (fileExcludes) {
			if (!query.excludePattern) {
				query.excludePattern = fileExcludes;
			} else {
				objects.mixin(query.excludePattern, fileExcludes, false /* no overwrite */);
			}
		}
	}

	private fileSearch(query: ISearchQuery): PPromise<ISearchComplete, ISearchProgressItem> {
		const rawSearchQuery = new PPromise<ISearchComplete, ISearchProgressItem>((onComplete, onError, onProgress) => {
			this.getWorkspaceFiles(query).then(files => {
				onComplete({
					results: files,
					stats: {} as any,
				});
			});
		}, () => rawSearchQuery && rawSearchQuery.cancel());
		return rawSearchQuery;
	}

	private textSearch(query: ISearchQuery): PPromise<ISearchComplete, ISearchProgressItem> {
		const workspace = this.contextService.getWorkspace().resource;
		const { repo, rev } = URIUtils.repoParams(workspace);
		const search = new PPromise<ISearchComplete, ISearchProgressItem>((complete, error, progress) => {
			fetchGraphQLQuery(`query($uri: String!, $pattern: String!, $rev: String!, $isRegExp: Boolean!, $isWordMatch: Boolean!, $isCaseSensitive: Boolean!) {
				root {
					repository(uri: $uri) {
						commit(rev: $rev) {
							commit {
								textSearch(pattern: $pattern, isRegExp: $isRegExp, isWordMatch: $isWordMatch, isCaseSensitive: $isCaseSensitive) {
									path
									lineMatches {
										preview
										lineNumber
										offsetAndLengths
									}
								}
							}
						}
					}
				}
			}`, { ...query.contentPattern, rev, uri: repo, }).then(data => {
					if (!data.root.repository || !data.root.repository.commit.commit) {
						throw new Error("Repository does not exist.");
					}
					let response = data.root.repository.commit.commit.textSearch.map(file => {
						const resource = workspace.with({ fragment: file.path });
						return { ...file, resource };
					});
					complete({
						results: response,
						stats: {} as any,
					});
				});
		}, () => search && search.cancel());
		return search;
	}

	/**
	 * search returns a promise which completes with file search results matching the specified query.
	 */
	public search(query: ISearchQuery): PPromise<ISearchComplete, ISearchProgressItem> {
		this.extendQuery(query);

		if (query.type === QueryType.File) {
			return this.fileSearch(query);
		}
		return this.textSearch(query);
	}

	/**
	 * searchRepo returns a promise which completes with repo search results matching the specified query.
	 */
	public searchRepo(query: ISearchQuery): PPromise<ISearchComplete, ISearchProgressItem> {
		this.extendQuery(query);

		const convertResults = repoList => {
			const results: any[] = [];
			for (const repo of repoList) {
				// TODO(john): support jumping to repos that haven't been cloned yet
				if (repo.lastIndexedRevOrLatest && repo.lastIndexedRevOrLatest.commit && repo.lastIndexedRevOrLatest.commit.sha1) {
					results.push(new FileMatch(URI.parse(`git://${repo.uri}?${repo.lastIndexedRevOrLatest.commit.sha1}`)));
				}
			}
			return results.filter(file => this.matches(file.resource, query.filePattern!, query.includePattern!, query.excludePattern!));
		};

		const rawSearchQuery = new PPromise<ISearchComplete, ISearchProgressItem>((onComplete, onError, onProgress) => {
			const cacheHit = reposCache.get(query.filePattern!);
			if (cacheHit) {
				return onComplete({ results: convertResults(cacheHit), stats: {} as any });
			}
			// repo search doesn't work for on-premises. the quick and dirty hack is to return all repos,
			// then allow VSCode to do the filtering on the front-end
			// TODO(neelance): fix repo search
			fetchGraphQLQuery(`query {
				root {
					repositories(query: $query) {
						uri
						lastIndexedRevOrLatest {
							commit {
								sha1
							}
						}
					}
				}
			}`, { query: isOnPremInstance(context.authEnabled) ? "" : query.filePattern })
				.then(data => {
					reposCache.set(query.filePattern!, data.root.repositories);
					onComplete({ results: convertResults(data.root.repositories), stats: {} as any });
				});
		}, () => rawSearchQuery && rawSearchQuery.cancel());
		return rawSearchQuery;
	}

	/**
	 * getWorkspaceFiles returns a promise which completes with the complete set of files
	 * in a workspace which match the specified query.
	 */
	private getWorkspaceFiles(query: ISearchQuery): TPromise<IFileMatch[]> {
		function getURIs(stat: IFileStat): URI[] {
			if (!stat.isDirectory) {
				return [stat.resource];
			}
			if (stat.children) {
				return flatten(stat.children.map(getURIs));
			}
			return [];
		}

		if (query.type === QueryType.File) {
			if (logSearchTiming) { console.time("search " + query.filePattern); } // tslint:disable-line no-console
			return getFilesCached(this.contextService.getWorkspace().resource)
				.then(fileNames => {
					let matches: FileMatch[] = [];
					for (const fileName of fileNames) {
						// Fast path to eliminate vendored files, which slow down search considerably.
						if (defaultExcludesRegExp.test(fileName)) {
							continue;
						}

						const resource = this.contextService.getWorkspace().resource.with({ fragment: fileName });
						if (this.matches(resource, query.filePattern!, query.includePattern!, query.excludePattern!)) {
							matches.push(new FileMatch(resource));
						}

						// maxResults is 0 when quickopen is initially
						// opened because we choose to not show any
						// files there.
						if (matches.length >= (query.maxResults || 0)) {
							break;
						}
					}
					if (logSearchTiming) { console.timeEnd("search " + query.filePattern); } // tslint:disable-line no-console
					return matches;
				});
		}

		return TPromise.wrap([]);
	}

	/**
	 * matches is used to filter candidate search results. It is mostly copied from vscode's search service implementation.
	 */
	private matches(resource: URI, filePattern: string, includePattern: glob.IExpression, excludePattern: glob.IExpression): boolean {
		// NOTE: This assumes the workspace is always at the root of
		// the repository. If this no longer holds, you must use
		// this.contextService.toWorkspaceRelativePath instead of just
		// the fsPath below.

		// If fragment is empty, it is a repo URI (not a file path).
		const fsPath = resource.fragment ? URIUtils.repoParams(resource).path : resource.path;

		// file pattern
		if (filePattern) {
			if (resource.scheme !== "git") {
				return false; // if we match on file pattern, we have to ignore non file resources
			}

			if (!scorer.matches(fsPath, strings.stripWildcards(filePattern).toLowerCase())) {
				return false;
			}
		}

		// includes
		if (includePattern) {
			if (resource.scheme !== "git") {
				return false; // if we match on file patterns, we have to ignore non file resources
			}

			if (!glob.match(includePattern, fsPath)) {
				return false;
			}
		}

		// excludes
		if (excludePattern) {
			if (resource.scheme !== "git") {
				return true; // e.g. untitled files can never be excluded with file patterns
			}

			if (glob.match(excludePattern, fsPath)) {
				return false;
			}
		}

		return true;
	}

	public clearCache(cacheKey: string): TPromise<void> {
		return TPromise.as(void 0); // noop
	}
}
