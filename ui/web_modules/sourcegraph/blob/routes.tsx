import { makeRepoRev } from "sourcegraph/repo";
import { urlTo } from "sourcegraph/util/urlTo";

// urlToBlob generates the URL to a file. To get a dir's URL, use urlToTree.
export function urlToBlob(repo: string, rev: string | null, path: string | string[]): string {
	const pathStr = typeof path === "string" ? path : path.join("/");
	return urlTo("blob", { splat: [makeRepoRev(repo, rev), pathStr] });
}

export function urlToBlobLine(repo: string, rev: string | null, path: string, line: number): string {
	return `${urlToBlob(repo, rev, path)}#L${line}`;
}

export function urlToBlobLineCol(repo: string, rev: string | null, path: string, line: number, col: number): string {
	return `${urlToBlob(repo, rev, path)}#L${line}:${col}`;
}
