import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { IFileStat, IResolveFileOptions, IUpdateContentOptions } from "vs/platform/files/common/files";

import { URIUtils } from "sourcegraph/core/uri";
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

	resolveFile(resource: URI, options?: IResolveFileOptions): TPromise<IFileStat> {
		return this.getFilesCached(resource).then(files => {
			const cachedStat = fileStatCache.get(resource.toString(true));
			if (cachedStat) {
				return cachedStat;
			}
			const fileStat = toFileStat(resource, files);
			fileStatCache.set(resource.toString(true), fileStat);
			return fileStat;
		});
	}

	existsFile(resource: URI): TPromise<boolean> {
		return this.getFilesCached(resource).then(files => {
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
}

function retrieveFilesAndDirs(resource: URI): any {
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
		hasChildren: childStats.length > 0,
		isDirectory: childStats.length > 0,
		resource: resource,
		name: resource.fragment,
		mtime: 0,
		etag: resource.toString(),
		children: childStats,
	};
}
