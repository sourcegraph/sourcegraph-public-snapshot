import * as uniq from "lodash/uniq";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { IFileService, IFilesConfiguration, IResolveFileOptions, IFileStat, IContent, IStreamContent, IImportResult, IResolveContentOptions, IUpdateContentOptions } from "vs/platform/files/common/files";
import * as electronService from "vs/workbench/services/files/electron-browser/fileService";
import * as nodeService from "vs/workbench/services/files/node/fileService";

import { URIUtils } from "sourcegraph/core/uri";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";

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

nodeService.FileService = FileService;
electronService.FileService = FileService;


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
	let children = files.filter(x => x.startsWith(path) && x !== path);
	children = children.map(x => {
		x = x.substr(path.length);
		return x.split("/")[0] || x;
	});
	children = uniq(children);
	const childStats = children.map(x => toFileStat(
		URIUtils.pathInRepo(repo, rev, path + x),
		files,
	));
	return {
		hasChildren: childStats.length > 0,
		isDirectory: childStats.length > 0, // TODO
		resource: resource,
		name: path,
		mtime: 0,
		etag: resource.toString(),
		children: childStats,
	};
}
