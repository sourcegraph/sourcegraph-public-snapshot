// tslint:disable: typedef ordered-imports

import { abs, rel, getRouteParams } from "sourcegraph/app/routePatterns";
import { urlTo } from "sourcegraph/util/urlTo";
import { makeRepoRev, repoPath, repoRev } from "sourcegraph/repo";
import { withLineColBoundToHash } from "sourcegraph/blob/withLineColBoundToHash";
import { withLastSrclibDataVersion } from "sourcegraph/blob/withLastSrclibDataVersion";
import { withResolvedRepoRev } from "sourcegraph/repo/withResolvedRepoRev";
import { withFileBlob } from "sourcegraph/blob/withFileBlob";
import { BlobMain } from "sourcegraph/blob/BlobMain";
import { RepoNavContext } from "sourcegraph/blob/RepoNavContext";

let _blobMainComponent: any;

export const routes = [
	{
		path: rel.blob,
		keepScrollPositionOnRouteChangeKey: "file",
		getComponents: (location: Location, callback: Function) => {
			if (!_blobMainComponent) {
				// Create only once to avoid unnecessary remounting after each route change.
				_blobMainComponent = withLineColBoundToHash(
					withResolvedRepoRev(
						withLastSrclibDataVersion(
							withFileBlob(
								BlobMain
							)
						)
					)
				);
			}
			callback(null, {
				main: _blobMainComponent,
				repoNavContext: RepoNavContext,
			});
		},
		blobLoaderHelpers: [],
	},
];

// urlToBlob generates the URL to a file. To get a dir's URL, use urlToTree.
export function urlToBlob(repo: string, rev: string | null, path: string | string[]): string {
	const pathStr = typeof path === "string" ? path : path.join("/");
	return urlTo("blob", { splat: [makeRepoRev(repo, rev), pathStr] } as any);
}

export function urlToBlobLine(repo: string, rev: string | null, path: string, line: number): string {
	return `${urlToBlob(repo, rev, path)}#L${line}`;
}

export function parseBlobURL(urlPathname: string): {repo: string, rev: string | null, path: string} {
	if (urlPathname.startsWith("/")) {
		urlPathname = urlPathname.slice(1);
	}
	const params = getRouteParams(abs.blob, urlPathname);
	if (!params.splat || params.splat.length !== 2) {
		throw new Error(`invalid blob URL: ${urlPathname} (params are: ${JSON.stringify(params)})`);
	}
	return {
		repo: repoPath(params.splat[0]),
		rev: repoRev(params.splat[0]),
		path: params.splat[1],
	};
}
