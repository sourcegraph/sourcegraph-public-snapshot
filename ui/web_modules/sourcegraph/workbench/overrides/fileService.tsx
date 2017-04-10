import Event, { Emitter } from "vs/base/common/event";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { FileChangesEvent, FileOperationEvent, IBaseStat, IContent, IFileService, IFileStat, IImportResult, IResolveContentOptions, IResolveFileOptions, IStreamContent, IUpdateContentOptions } from "vs/platform/files/common/files";
import { IWorkspace, IWorkspaceContextService } from "vs/platform/workspace/common/workspace";

import { URIUtils } from "sourcegraph/core/uri";
import { contentCache, fetchContentAndResolveRev } from "sourcegraph/editor/contentLoader";
import { fetchGQL } from "sourcegraph/util/gqlClient";
import { getURIContext } from "sourcegraph/workbench/utils";

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

/**
 * createdFiles keeps track of which files have been created within in the workspace.
 */
const createdFiles = new Set<string>();

/**
 * deletedFiles keeps track of which files have been created within in the workspace.
 */
const deletedFiles = new Set<string>();

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
		createdFiles.add(resource.toString());
		return this.updateContent(resource, content).then(result => {
			fileStatCache.set(resource.toString(), result);
			const key = this.contextService.getWorkspace().resource.toString();
			let currentFiles = workspaceFiles.get(key) || [];
			workspaceFiles.set(key, currentFiles.concat(getURIContext(resource).path));
			return result;
		});
	}

	public touchFile(resource: URI): TPromise<IFileStat> {
		return this.createFile(resource);
	}

	public del(resource: URI): TPromise<void> {
		resource = URIUtils.tryConvertGitToFileURI(resource);

		const key = this.contextService.getWorkspace().resource.toString();
		let wsFiles = workspaceFiles.get(key);
		if (wsFiles) {
			workspaceFiles.set(key, wsFiles.filter(item => item !== getURIContext(resource).path));
			fileStatCache.delete(resource.toString());
			contentCache.delete(resource.toString());
			createdFiles.delete(resource.toString());
			deletedFiles.add(resource.toString());
		}
		return TPromise.as(void 0);
	}

	resolveFile(resource: URI, options?: IResolveFileOptions): TPromise<IFileStat> {
		const { rev } = getURIContext(resource);
		return getFilesCached({ resource, revState: { commitID: rev || undefined } }).then(files => {
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
		let originalResource = resource;
		if (resource.scheme === "zap") {
			resource = URIUtils.tryConvertGitToFileURI(resource);
			// The zap URI scheme indicates we need to intialize a file that has been created on a zap ref
			// before attempting to resolve the contents.
			this.touchFile(resource);
		}

		const isViewingZapRev = Boolean(this.contextService.getWorkspace().revState && this.contextService.getWorkspace().revState!.zapRef);

		// contentCache acts like watchFileChanges in that it is set the first time when fetching content from
		// fetchContentAndResolveRev. It is updated when updateContent is called.
		// This behavior mimicks watchFileChanges which is used by VSCode to watch for content changes at the filesystem level.
		// We will need to build on this to handle renaming and moving files so their changes are reflected in the tree.
		let contents = contentCache.get(resource.toString());
		if (isViewingZapRev && resource.scheme === "git") {
			// Check to see if this was a file that was created after the workspace was initialized. If so,
			// this needs to be converted to the canonial file URI to retrieve the original contents since
			// the created file can't be fetched from the backend.
			const asFileURI = URIUtils.tryConvertGitToFileURI(resource);
			if (createdFiles.has(asFileURI.toString())) {
				contents = contentCache.get(asFileURI.toString());
			}
			if (deletedFiles.has(asFileURI.toString())) {
				// TODO: Ideally the UI should show the contents of the file before it was deleted but diff'd
				// in red. We would need to instead return the original contents of the file here for the left side 
				// of the diff editor have the diff view opener pass in the empty string as the right side.
				contents = "";
			}
		}
		if ((contents !== undefined && isViewingZapRev) || originalResource.scheme === "zap") {
			return TPromise.wrap({
				...toBaseStat(resource),
				value: contents || "",
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

	public dispose(): void { /* noop */ }
}

export function fetchFilesAndDirs(workspace: IWorkspace): Promise<GQL.IRoot | null> {
	const { repo } = getURIContext(workspace.resource);
	if (!repo) {
		return Promise.resolve(null);
	}
	return fetchGQL(`query FileTree($repo: String!, $rev: String!) {
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
		}`, { repo, rev: workspace.revState!.commitID }).then(resp => resp.data.root);
}

export function toBaseStat(resource: URI): IBaseStat {
	return {
		resource: resource,
		name: getURIContext(resource).path,
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
		const resourcePath = getURIContext(resource).path;
		if (candidate === resourcePath) {
			isFile = true;
			// skip over self
			continue;
		}

		const prefix = resourcePath ? resourcePath + "/" : "";
		if (!candidate.startsWith(prefix)) {
			continue;
		}

		const child = candidate.substr(prefix.length);
		const index = child.indexOf("/");
		if (index === -1) {
			childFiles.push(child);
		} else {
			const dir = child.substr(0, index);
			childDirectories.add(dir);
			// candidate is in one of resource's subdirectories,
			// so we forward it as a file candidate in the recursive call.
			addRecursiveFile(dir, candidate);
		}
	}

	for (const childDir of Array.from(childDirectories)) {
		const fileStat = toFileStat(resource.with({ path: resource.path + `/${childDir}` }), recursiveFilesByDirectory.get(childDir)!);
		childStats.push(fileStat);
	}
	for (const child of childFiles) {
		const fileStat = toFileStat(resource.with({ path: resource.path + `/${child}` }), []);
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
 * TODO(john): do we need this now that we have the Apollo caching layer?
 */
export function getFilesCached(workspace: IWorkspace): TPromise<string[]> {
	const key = workspace.resource.toString() + "?" + workspace.revState!.commitID;
	if (workspaceFiles.has(key)) {
		return TPromise.wrap(workspaceFiles.get(key));
	}

	return TPromise.wrap(fetchFilesAndDirs(workspace).then(root => {
		if (!root || !root.repository) {
			return [];
		}
		const files: string[] = root.repository!.commit.commit!.tree!.files.map(file => file.name);
		workspaceFiles.set(key, files);
		return files;
	}));
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
