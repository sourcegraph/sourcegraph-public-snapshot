// tslint:disable: typedef ordered-imports curly

import {rel} from "sourcegraph/app/routePatterns";
import {urlTo} from "sourcegraph/util/urlTo";
import {makeRepoRev} from "sourcegraph/repo/index";
import {lineColBoundToHash} from "sourcegraph/blob/lineColBoundToHash";
import {withLastSrclibDataVersion} from "sourcegraph/blob/withLastSrclibDataVersion";
import {BlobLoader} from "sourcegraph/blob/BlobLoader";
import {RepoNavContext} from "sourcegraph/blob/RepoNavContext";

export const routes = [
	{
		path: rel.blob,
		keepScrollPositionOnRouteChangeKey: "file",
		getComponents: (location: Location, callback: Function) => {
			callback(null, {
				main: BlobLoader,
				repoNavContext: RepoNavContext,
			});
		},
		blobLoaderHelpers: [lineColBoundToHash, withLastSrclibDataVersion],
	},
];

// urlToBlob generates the URL to a file. To get a dir's URL, use urlToTree.
export function urlToBlob(repo: string, rev: string | null, path: string | string[]): string {
	const pathStr = typeof path === "string" ? path : path.join("/");
	return urlTo("blob", {splat: [makeRepoRev(repo, rev), pathStr]} as any);
}
