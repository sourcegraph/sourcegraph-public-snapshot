import * as without from "lodash/without";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { IBaseStat, IContent, IFileStat, IResolveContentOptions, IResolveFileOptions, IStreamContent, IUpdateContentOptions } from "vs/platform/files/common/files";

import { IEventService } from "vs/platform/event/common/event";
import { IWorkspace, IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { LocalFileChangeEvent } from "vs/workbench/services/textfile/common/textfiles";

import { WorkspaceOp, compose, isFilePath, stripFileOrBufferPathPrefix } from "libzap/lib/ot/workspace";

import { URIUtils } from "sourcegraph/core/uri";
import { contentCache, fetchContentAndResolveRev } from "sourcegraph/editor/contentLoader";
import { Features } from "sourcegraph/util/features";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";
import { OutputListener } from "sourcegraph/workbench/outputListener";

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
	private zapFileService: ZapFileService;
	private workspace: IWorkspace;

	constructor(
		@IEventService private eventService: IEventService,
		@IWorkspaceContextService contextService: IWorkspaceContextService,
	) {
		this.eventService = eventService;
		this.workspace = contextService.getWorkspace();

		if (Features.zap.isEnabled()) {
			this.zapFileService = new ZapFileService();
			OutputListener.subscribe("zapFileTree", (msg: string) => {
				const data = JSON.parse(msg);
				if (data.reset) {
					this.zapFileService.reset(data);
					this.refreshTree();
				} else {
					this.onReceiveOp(JSON.parse(msg));
				}
			});
		}
	}

	resolveFile(resource: URI, options?: IResolveFileOptions): TPromise<IFileStat> {
		return this.getFilesCached(resource).then(files => {
			if (Features.zap.isEnabled()) {
				// TODO(renfred): figure out how to take advantage of the stat cache with zap files.
				files = this.zapFileService.updateTree(this.workspace.resource, files.slice());
			} else {
				const cachedStat = fileStatCache.get(resource.toString(true));
				if (cachedStat) {
					return cachedStat;
				}
			}
			const fileStat = toFileStat(resource, files);
			fileStatCache.set(resource.toString(true), fileStat);
			return fileStat;
		});
	}

	resolveContent(resource: URI, options?: IResolveContentOptions): TPromise<IContent> {
		return TPromise.wrap(fetchContentAndResolveRev(resource)).then(({ content }) => {
			return {
				...toBaseStat(resource),
				value: content,
				encoding: "utf8",
			};
		});
	}

	resolveStreamContent(resource: URI, options?: IResolveContentOptions): TPromise<IStreamContent> {
		return this.resolveContent(resource, options).then(content => ({
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
		}));
	}

	existsFile(resource: URI): TPromise<boolean> {
		return this.getFilesCached(resource).then(files => {
			if (Features.zap.isEnabled()) {
				files = this.zapFileService.updateTree(this.workspace.resource, files.slice());
			}
			const path = resource.fragment;
			return Boolean(files.find(file => file === path));
		});
	}

	/**
	 * Gets and caches a list of all of the files in a workspace.
	 */
	private getFilesCached(resource: URI): TPromise<string[]> {
		const { repo, rev } = URIUtils.repoParams(resource);
		const key = repo + rev;
		if (workspaceFiles.has(key)) {
			return TPromise.wrap(workspaceFiles.get(key));
		}
		return retrieveFilesAndDirs(resource).then(({ root }) => {
			const files: string[] = root.repository.commit.commit.tree.files.map(file => file.name);
			workspaceFiles.set(key, files);
			return files;
		});
	}

	// Stubbed implementation to handle updating the configuration from the VSCode extension
	public updateContent(resource: URI, value: string, options: IUpdateContentOptions = Object.create(null)): TPromise<IFileStat> {
		return TPromise.as({ isDirectory: true, hasChildren: false });
	}

	public onReceiveOp(op: WorkspaceOp): void {
		this.zapFileService.apply(op);
		this.refreshTree();
	}

	public refreshTree(): void {
		// Use this event to trigger the refresh of the file tree in ExplorerView.
		this.eventService.emit("files.internal:fileChanged", new LocalFileChangeEvent());
	}
}

class ZapFileService {
	private state: WorkspaceOp;

	public apply(op: WorkspaceOp): void {
		if (!this.state) {
			this.state = op;
		} else {
			this.state = compose(this.state, op);
		}
	}

	public reset(history?: WorkspaceOp): void {
		this.state = history || {};
	}

	public updateTree(resource: URI, files: string[]): string[] {
		if (!this.state) { return files; }

		if (this.state.create) {
			for (const f of this.state.create) {
				if (isFilePath(f)) {
					const resourceKey = URIUtils.withFilePath(resource, stripFileOrBufferPathPrefix(f)).toString();
					let content = "";
					if (this.state.edit && this.state.edit[f]) {
						if (this.state.edit[f].length > 1 || typeof this.state.edit[f][0] !== "string") {
							throw new Error(`updateTree: initial edit op for ${f} should only contain one insert`);
						}
						content = this.state.edit[f][0] as string;
					}
					// Use the content cache to store the inital content of the file.
					// TODO use TextDocumentService instead to enable editing capabilities.
					contentCache.set(resourceKey, content);

					files.push(stripFileOrBufferPathPrefix(f));
				}
			}
		}
		if (this.state.delete) {
			for (const f of this.state.delete) {
				if (isFilePath(f)) {
					const resourceKey = URIUtils.withFilePath(resource, stripFileOrBufferPathPrefix(f)).toString();
					contentCache.delete(resourceKey);

					files = without(files, f);
				}
			}
		}
		return files;
	}
}



function retrieveFilesAndDirs(resource: URI): TPromise<any> {
	const { repo, rev } = URIUtils.repoParams(resource);
	return TPromise.wrap(fetchGraphQLQuery(`query Files($repo: String!, $rev: String!) {
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
		}`, { repo, rev }));
}

function toBaseStat(resource: URI): IBaseStat {
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
function toFileStat(resource: URI, files: string[]): IFileStat {
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
