import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { IFileStat, IResolveFileOptions } from "vs/platform/files/common/files";

import { URIUtils } from "sourcegraph/core/uri";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";

let fileStatCache: Map<string, IFileStat> = new Map();

// FileService provides files from Sourcegraph's API instead of a normal file
// system. It is used to find the files in a Workspace, but not for retrieving
// file content. File content is resolved using the modelResolver, which uses
// contentLoader.tsx.
export class FileService {
	constructor() {
		//
	}

	resolveFile(resource: URI, options?: IResolveFileOptions): TPromise<IFileStat> {
		return retrieveFilesAndDirs(resource).then(({ root }) => {
			const cachedStat = fileStatCache.get(resource.toString(true));
			if (cachedStat) {
				return cachedStat;
			}
			const files = root.repository.commit.commit.tree.files.map(file => file.name);
			const fileStat = toFileStat(resource, files);
			fileStatCache.set(resource.toString(true), fileStat);
			return fileStat;
		});
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
