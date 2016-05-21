// @flow

import {rel} from "sourcegraph/app/routePatterns";
import urlTo from "sourcegraph/util/urlTo";
import {makeRepoRev} from "sourcegraph/repo";

export const routes = [
	{
		path: rel.blob,
		getComponents: (location: Location, callback: Function) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/blob/BlobLoader").default,
					repoNavContext: require("sourcegraph/blob/RepoNavContext").default,
				}, [
					require("sourcegraph/blob/lineColBoundToHash").default,
				]);
			});
		},
	},
];

// urlToBlob generates the URL to a file. To get a dir's URL, use urlToTree.
export function urlToBlob(repo: string, rev: ?string, path: string | string[]): string {
	const pathStr = typeof path === "string" ? path : path.join("/");
	return urlTo("blob", {splat: [makeRepoRev(repo, rev), pathStr]});
}
