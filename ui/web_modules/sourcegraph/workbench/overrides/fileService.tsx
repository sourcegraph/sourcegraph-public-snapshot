import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { IBaseStat, IContent, IFileStat, IResolveContentOptions, IResolveFileOptions, IStreamContent, IUpdateContentOptions } from "vs/platform/files/common/files";

import { IEventService } from "vs/platform/event/common/event";
import { IWorkspace, IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { LocalFileChangeEvent } from "vs/workbench/services/textfile/common/textfiles";

import { URIUtils } from "sourcegraph/core/uri";
import { contentCache, fetchContentAndResolveRev } from "sourcegraph/editor/contentLoader";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";

/**
 * Both of these caches will last until a hard navigation or refresh. We will
 * need to have a mechanism to bust the cache for Xylo, or alternatively turn
 * them into active stores.
 */

/**
 * The fileStatCache saves us from recomputing the expensive IFileStat. On
 * large dirs it can take a second. The cache will last until a hard navigation
 * or refresh.
 */
const fileStatCache: Map<string, IFileStat> = new Map();

/**
 * workspaceFiles caches the contents of a directory. This lets us prevent
 * doing multiple round trips to fetch the contents of the same directory.
 */
const workspaceFiles: Map<string, string[]> = new Map();

// FileService provides files from Sourcegraph's API instead of a normal file
// system. It is used to find the files in a Workspace, but not for retrieving
// file content. File content is resolved using the modelResolver, which uses
// contentLoader.tsx.
export class FileService {
	private workspace: IWorkspace;

	constructor(
		@IEventService private eventService: IEventService,
		@IWorkspaceContextService contextService: IWorkspaceContextService,
	) {
		this.eventService = eventService;
		this.workspace = contextService.getWorkspace();
	}

	createFile(resource: URI, content: string = ""): TPromise<IFileStat> {
		contentCache.set(resource.toString(), content);
		return this.updateContent(resource, content).then(result => {
			fileStatCache.set(resource.toString(), result);
			const { repo, rev } = URIUtils.repoParams(this.workspace.resource);
			const key = repo + rev;
			let currentFiles = workspaceFiles.get(key) || [];
			workspaceFiles.set(key, currentFiles.concat(resource.fragment));
			return result;
		});
	}

	public touchFile(resource: URI): TPromise<IFileStat> {
		return this.createFile(resource);
	}

	resolveFile(resource: URI, options?: IResolveFileOptions): TPromise<IFileStat> {
		return getFilesCached(resource).then(files => {
			const fileStat = toFileStat(resource, files);
			fileStatCache.set(resource.toString(true), fileStat);
			return fileStat;
		});
	}

	resolveContent(resource: URI, options?: IResolveContentOptions): TPromise<IContent> {
		if (resource.scheme === "zap") {
			let zapResource = resource.with({ scheme: "git" });
			this.touchFile(zapResource).then(() => this.refreshTree());
			resource = zapResource;
		}

		// contentCache acts like watchFileChanges in that it is set the first time when fetching content from 
		// fetchContentAndResolveRev. It is updated when updateContent is called.
		// This behavior mimicks watchFileChanges which is used by VSCode to watch for content changes at the filesystem level.
		// We will need to build on this to handle renaming and moving files so their changes are reflected in the tree.
		const contents = contentCache.get(resource.toString());
		if (contents) {
			return TPromise.wrap({
				...toBaseStat(resource),
				value: contents,
				encoding: "utf8",
			});
		}

		return TPromise.wrap(fetchContentAndResolveRev(resource)).then(({ content }) => {
			return {
				...toBaseStat(resource),
				value: content,
				encoding: "utf8",
			};
		});
	}

	resolveStreamContent(resource: URI, options?: IResolveContentOptions): TPromise<IStreamContent> {
		return this.resolveContent(resource, options).then(content => {
			return ({
				...content,
				value: {
					on: (event: string, callback: Function): void => {
						if (event === "data") {
							callback(content.value);
						}
						if (event === "end") {
							callback();
						}
					}
				},
			});
		});
	}

	existsFile(resource: URI): TPromise<boolean> {
		return this.resolveFile(resource).then(() => true, () => false);
	}

	private resolve(resource: URI, options: IResolveFileOptions = Object.create(null)): TPromise<IFileStat> {
		return this.toStatResolver(resource)
			.then(model => model.resolve(options));
	}

	private toStatResolver(resource: URI): TPromise<StatResolver> {
		let time = new Date().getTime();
		return TPromise.as(new StatResolver(resource, false, time, 1, false));
	}

	// Stubbed implementation to handle updating the configuration from the VSCode extension
	public updateContent(resource: URI, value: string, options: IUpdateContentOptions = Object.create(null)): TPromise<IFileStat> {
		return this.resolve(resource).then((fileStat) => {
			contentCache.set(resource.toString(), value);
			return fileStat;
		});
	}

	public refreshTree(): void {
		// Use this event to trigger the refresh of the file tree in ExplorerView.
		this.eventService.emit("files.internal:fileChanged", new LocalFileChangeEvent());
	}
}

export function fetchFilesAndDirs(resource: URI): any {
	const { repo, rev } = URIUtils.repoParams(resource);
	return fetchGraphQLQuery(`query Files($repo: String!, $rev: String!) {
			root {
				repository(uri: $repo) {
					uri
					description
					defaultBranch
					commit(rev: $rev) {
						commit {
							tree(recursive: true) {
								files {
									name
								}
							}
						}
						cloneInProgress
					}
				}
			}
		}`, { repo, rev });
}

export function toBaseStat(resource: URI): IBaseStat {
	return {
		resource: resource,
		name: resource.fragment,
		mtime: 0,
		etag: resource.toString(),
	};
}

/**
 * toFileStat returns a tree of IFileStat that represents the repository rooted at resource.
 * The files parameter is all available files in the repository.
 */
export function toFileStat(resource: URI, files: string[]): IFileStat {
	const childStats: IFileStat[] = [];
	const childDirectories = new Set();
	const childFiles: string[] = [];

	// When we recursively call toFileStat, don't forward files that aren't a transitive child of resource.
	// This is a noticible performance optimization for large repos.
	const recursiveFiles: string[] = [];

	// looking for children of resource
	for (const candidate of files) {
		if (candidate === resource.fragment) {
			// skip over self
			continue;
		}

		const prefix = resource.fragment ? resource.fragment + "/" : "";
		if (!candidate.startsWith(prefix)) {
			// candidate is not a subresource of resource
			continue;
		}

		const child = candidate.substr(prefix.length);
		const index = child.indexOf("/");
		if (index === -1) {
			childFiles.push(candidate);
		} else {
			const dir = prefix + child.substr(0, index);
			childDirectories.add(dir);

			// candidate is in one of resource's subdirectories,
			// so we forward it as a file candidate in the recursive call.
			recursiveFiles.push(candidate);
		}
	}

	const children = Array.from(childDirectories).concat(childFiles);
	for (const child of children) {
		const fileStat = toFileStat(resource.with({ fragment: child }), recursiveFiles);
		childStats.push(fileStat);
	}

	return {
		...toBaseStat(resource),
		hasChildren: childStats.length > 0,
		isDirectory: childStats.length > 0,
		children: childStats,
	};
}

/**
 * Gets and caches a list of all of the files in a workspace.
 */
export function getFilesCached(resource: URI): TPromise<string[]> {
	const { repo, rev } = URIUtils.repoParams(resource);
	const key = repo + rev;
	if (workspaceFiles.has(key)) {
		return TPromise.wrap(workspaceFiles.get(key));
	}
	return fetchFilesAndDirs(resource).then(({ root }) => {
		const files: string[] = root.repository.commit.commit.tree.files.map(file => file.name);
		workspaceFiles.set(key, files);
		return files;
	});
}

export class StatResolver {
	private resource: URI;
	private isDirectory: boolean;
	private mtime: number;
	private name: string;
	private etag: string;
	private size: number;
	private verboseLogging: boolean;

	constructor(resource: URI, isDirectory: boolean, mtime: number, size: number, verboseLogging: boolean) {

		this.resource = resource;
		this.isDirectory = isDirectory;
		this.mtime = mtime;
		this.name = resource.fsPath;
		this.etag = resource.toString();
		this.size = size;

		this.verboseLogging = verboseLogging;
	}

	public resolve(options: IResolveFileOptions): TPromise<IFileStat> {

		// General Data
		const fileStat: IFileStat = {
			resource: this.resource,
			isDirectory: this.isDirectory,
			hasChildren: undefined as any,
			name: this.name,
			etag: this.etag,
			size: this.size,
			mtime: this.mtime
		};

		// File Specific Data
		if (!this.isDirectory) {
			return TPromise.as(fileStat);
		} else {
			// Convert the paths from options.resolveTo to absolute paths
			let absoluteTargetPaths: string[] = null as any;
			if (options && options.resolveTo) {
				absoluteTargetPaths = [];
				options.resolveTo.forEach(resource => {
					absoluteTargetPaths.push(resource.fsPath);
				});
			}

			return new TPromise((c, e) => {
				c(fileStat);
			});
		}
	}
}
