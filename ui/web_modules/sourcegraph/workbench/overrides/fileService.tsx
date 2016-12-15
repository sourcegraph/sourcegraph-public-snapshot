import * as uniq from "lodash/uniq";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { IFileStat, IResolveFileOptions } from "vs/platform/files/common/files";
import * as electronService from "vs/workbench/services/files/electron-browser/fileService";
import * as nodeService from "vs/workbench/services/files/node/fileService";

import { URIUtils } from "sourcegraph/core/uri";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";

// File service provides files from Sourcegraph's API instead of a normal file
// system. It is used for Workspace discovery, but not for retrieving file
// content. File content is resolved using the modelResolver, which uses
// contentLoader.tsx.

class FileService {
	constructor() {
		//
	}

	resolveFile(resource: URI, options?: IResolveFileOptions): TPromise<IFileStat> {
		return retrieveFilesAndDirs(resource).then(({root}) => {
			const files = root.repository.commit.commit.tree.files.map(file => file.name);
			return toFileStat(resource, files);
		});
	}
}

(nodeService as any).FileService = FileService;
(electronService as any).FileService = FileService;


function retrieveFilesAndDirs(resource: URI): any {
	const {repo, rev} = URIUtils.repoParams(resource);
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
		}`, {repo, rev});
}

// Convert a list of files into a hierarchical file stat structure.
function toFileStat(resource: URI, files: string[]): IFileStat {
	const {repo, rev, path} = URIUtils.repoParams(resource);
	const childrenOfResource = files.filter(x => x.startsWith(path) && x !== path);
	const dirComponents = childrenOfResource.map(x => {
		x = x.substr(path.length);
		return x.split("/")[0] || x;
	});
	const uniqDirs = uniq(dirComponents);
	const childStats = uniqDirs.map(dir => toFileStat(
		URIUtils.pathInRepo(repo, rev, path + dir),
		files,
	));
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
