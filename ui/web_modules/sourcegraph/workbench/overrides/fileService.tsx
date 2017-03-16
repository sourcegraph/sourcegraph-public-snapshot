import Event, { Emitter } from "vs/base/common/event";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { FileChangesEvent, FileOperationEvent, IBaseStat, IContent, IFileService, IFileStat, IImportResult, IResolveContentOptions, IResolveFileOptions, IStreamContent, IUpdateContentOptions } from "vs/platform/files/common/files";

import { IWorkspace, IWorkspaceContextService } from "vs/platform/workspace/common/workspace";

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
export class FileService implements IFileService {
	_serviceBrand: any;
	private workspace: IWorkspace;

	private _onFileChanges: Emitter<FileChangesEvent> = new Emitter<FileChangesEvent>();
	private _onAfterOperation: Emitter<FileOperationEvent> = new Emitter<FileOperationEvent>();

	constructor(
		@IWorkspaceContextService private contextService: IWorkspaceContextService,
	) {
		this.workspace = contextService.getWorkspace();
	}

	public get onFileChanges(): Event<FileChangesEvent> {
		return this._onFileChanges.event;
	}

	public get onAfterOperation(): Event<FileOperationEvent> {
		return this._onAfterOperation.event;
	}

	public updateOptions(options: any): void {
		throw new Error("not implemented");
	}

	createFile(resource: URI, content: string = ""): TPromise<IFileStat> {
		contentCache.set(resource.toString(), content);
		return this.updateContent(resource, content).then(result => {
			fileStatCache.set(resource.toString(), result);
			const key = this.contextService.getWorkspace().resource.toString();
			let currentFiles = workspaceFiles.get(key) || [];
			workspaceFiles.set(key, currentFiles.concat(resource.fragment));
			return result;
		});
	}

	public touchFile(resource: URI): TPromise<IFileStat> {
		return this.createFile(resource);
	}

	public del(resource: URI): TPromise<void> {
		const key = this.contextService.getWorkspace().resource.toString();
		let wsFiles = workspaceFiles.get(key);
		if (wsFiles) {
			workspaceFiles.set(key, wsFiles.filter(item => item !== resource.fragment));
			fileStatCache.delete(resource.toString());
			contentCache.delete(resource.toString());
		}
		this.refreshTree();
		return TPromise.as(void 0);
	}

	resolveFile(resource: URI, options?: IResolveFileOptions): TPromise<IFileStat> {
		return getFilesCached(resource).then(files => {
			const statCacheHit = fileStatCache.get(resource.toString());
			if (statCacheHit) {
				// TODO(john): this may adversely impact zap -- we may want a flag to force re-computation
				// of the IFileStat object.
				return statCacheHit;
			}
			const fileStat = toFileStat(resource, files);
			fileStatCache.set(resource.toString(true), fileStat);
			return fileStat;
		});
	}

	public resolveContents(resources: URI[]): TPromise<IContent[]> {
		return TPromise.join(resources.map(resource => this.resolveContent(resource)));
	}

	resolveContent(resource: URI, options?: IResolveContentOptions): TPromise<IContent> {
		if (resource.scheme === "zap") {
			let zapResource = resource.with({ scheme: "git" });
			this.touchFile(zapResource).then(() => this.refreshTree());
			resource = zapResource;
		}

		const isViewingZapRef = Boolean(this.contextService.getWorkspace().revState && this.contextService.getWorkspace().revState!.zapRef);

		// contentCache acts like watchFileChanges in that it is set the first time when fetching content from
		// fetchContentAndResolveRev. It is updated when updateContent is called.
		// This behavior mimicks watchFileChanges which is used by VSCode to watch for content changes at the filesystem level.
		// We will need to build on this to handle renaming and moving files so their changes are reflected in the tree.
		const contents = contentCache.get(resource.toString());
		if (contents && isViewingZapRef) {
			return TPromise.wrap({
				...toBaseStat(resource),
				value: contents,
				encoding: "utf8",
			});
		}

		return TPromise.wrap(fetchContentAndResolveRev(resource, isViewingZapRef)).then(({ content }) => {
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
		return this.resolveFile(resource).then(stat => stat["exists"]);
	}

	public moveFile(source: URI, target: URI, overwrite?: boolean): TPromise<IFileStat> {
		throw new Error("not implemented");
	}

	public copyFile(source: URI, target: URI, overwrite?: boolean): TPromise<IFileStat> {
		return TPromise.wrap({ isDirectory: true, hasChildren: true } as IFileStat);
	}

	public createFolder(resource: URI): TPromise<IFileStat> {
		throw new Error("not implemented");
	}

	public rename(resource: URI, newName: string): TPromise<IFileStat> {
		throw new Error("not implemented");
	}

	public importFile(source: URI, targetFolder: URI): TPromise<IImportResult> {
		throw new Error("not implemented");
	}

	public watchFileChanges(resource: URI): void {
		throw new Error("not implemented");
	}

	public getEncoding(resource: URI): string {
		throw new Error("not implemented");
	}

	public unwatchFileChanges(resource: URI): void;
	public unwatchFileChanges(path: string): void;
	public unwatchFileChanges(arg1: any): void {
		throw new Error("not implemented");
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
		this._onFileChanges.fire(new FileChangesEvent([]));
	}

	public dispose(): void { /* noop */ }
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
	const childDirectories = new Set<string>();
	const childFiles: string[] = [];
	let isFile = false;

	// When we recursively call toFileStat, don't forward files that aren't a transitive child of resource.
	// This is a noticible performance optimization for large repos.
	const recursiveFilesByDirectory = new Map<string, string[]>();
	const addRecursiveFile = (dir: string, file: string) => {
		if (!recursiveFilesByDirectory.has(dir)) {
			recursiveFilesByDirectory.set(dir, []);
		}
		recursiveFilesByDirectory.get(dir)!.push(file);
	};

	// looking for children of resource
	for (const candidate of files) {
		if (candidate === resource.fragment) {
			isFile = true;
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
			addRecursiveFile(dir, candidate);
		}
	}

	for (const childDir of Array.from(childDirectories)) {
		const fileStat = toFileStat(resource.with({ fragment: childDir }), recursiveFilesByDirectory.get(childDir)!);
		childStats.push(fileStat);
	}
	for (const child of childFiles) {
		const fileStat = toFileStat(resource.with({ fragment: child }), []);
		childStats.push(fileStat);
	}

	const isDir = childStats.length > 0;

	return {
		...toBaseStat(resource),
		hasChildren: isDir,
		isDirectory: isDir,
		children: childStats,
		exists: isDir || isFile,
	};
}

/**
 * Gets and caches a list of all of the files in a workspace.
 */
export function getFilesCached(resource: URI): TPromise<string[]> {
	const key = resource.toString();
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
