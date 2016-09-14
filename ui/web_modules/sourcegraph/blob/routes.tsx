// tslint:disable: typedef ordered-imports

import { rel } from "sourcegraph/app/routePatterns";
import { urlTo } from "sourcegraph/util/urlTo";
import { makeRepoRev } from "sourcegraph/repo";
import { withLineColBoundToHash } from "sourcegraph/blob/withLineColBoundToHash";
import { withLastSrclibDataVersion } from "sourcegraph/blob/withLastSrclibDataVersion";
import { withResolvedRepoRev } from "sourcegraph/repo/withResolvedRepoRev";
import { withFileBlob } from "sourcegraph/blob/withFileBlob";
import { withAnnotations } from "sourcegraph/blob/withAnnotations";
import { LegacyBlobMain } from "sourcegraph/blob/LegacyBlobMain";
import { BlobMain } from "sourcegraph/blob/BlobMain";
import { RepoNavContext } from "sourcegraph/blob/RepoNavContext";

let componentForBlobMain = LegacyBlobMain;
if (global.window && localStorage.getItem("monaco") === "true") {
	componentForBlobMain = BlobMain;
}

export const routes = [
	{
		path: rel.blob,
		keepScrollPositionOnRouteChangeKey: "file",
		getComponents: (location: Location, callback: Function) => {
			callback(null, {
				main: withLineColBoundToHash(
					withResolvedRepoRev(
						withLastSrclibDataVersion(
							withFileBlob(
								withAnnotations(
									componentForBlobMain
								)
							)
						)
					)
				),
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
