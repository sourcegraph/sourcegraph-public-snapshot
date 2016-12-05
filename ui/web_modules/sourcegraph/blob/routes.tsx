import { abs, getRouteParams } from "sourcegraph/app/routePatterns";
import { makeRepoRev, repoPath, repoRev } from "sourcegraph/repo";
import { urlTo } from "sourcegraph/util/urlTo";

// urlToBlob generates the URL to a file. To get a dir's URL, use urlToTree.
export function urlToBlob(repo: string, rev: string | null, path: string | string[]): string {
	const pathStr = typeof path === "string" ? path : path.join("/");
	return urlTo("blob", { splat: [makeRepoRev(repo, rev), pathStr] } as any);
}

export function urlToBlobLine(repo: string, rev: string | null, path: string, line: number): string {
	return `${urlToBlob(repo, rev, path)}#L${line}`;
}

export function urlToBlobLineCol(repo: string, rev: string | null, path: string, line: number, col: number): string {
	return `${urlToBlob(repo, rev, path)}#L${line}:${col}`;
}

export function parseBlobURL(urlPathname: string): { repo: string, rev: string | null, path: string, hash?: string } {
	let hash: string | undefined;
	if (urlPathname.indexOf("#") !== -1) {
		const idx = urlPathname.indexOf("#");
		hash = urlPathname.slice(idx + 1);
		urlPathname = urlPathname.slice(0, idx);
	}
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
		hash,
	};
}
