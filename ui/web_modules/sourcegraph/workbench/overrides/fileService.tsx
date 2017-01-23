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

// Convert a list of files into a hierarchical file stat structure.
function toFileStat(resource: URI, files: string[]): IFileStat {
	let path = resource.fragment;
	const directories = new Map();
	const childFiles: string[] = [];
	const childStats: IFileStat[] = [];
	for (const candidate of files) {
		const index = candidate.indexOf("/");
		if (index === -1) {
			childFiles.push(candidate);
			continue;
		}
		const dir = candidate.substr(0, index);
		if (!directories.has(dir)) {
			directories.set(dir, []);
		}
		directories.get(dir).push(candidate.substr(index + 1));
	}
	path += path ? "/" : "";
	directories.forEach((children, dir) => {
		childStats.push(toFileStat(
			resource.with({ fragment: path + dir }),
			children,
		));
	});
	childFiles.forEach((file) => {
		childStats.push(toFileStat(
			resource.with({ fragment: path + file }),
			[],
		));
	});
	return {
		hasChildren: childStats.length > 0,
		isDirectory: childStats.length > 0,
		resource: resource,
		name: path,
		mtime: 0,
		etag: resource.toString(),
		children: childStats,
	};
}
