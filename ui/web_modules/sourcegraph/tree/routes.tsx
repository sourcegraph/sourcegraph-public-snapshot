import { makeRepoRev } from "sourcegraph/repo";
import { urlToRepoRev } from "sourcegraph/repo/routes";
import { urlTo } from "sourcegraph/util/urlTo";

// urlToTree generates the URL to a dir. To get a file's URL, use urlToBlob.
export function urlToTree(repo: string, rev: string | null, path: string | string[]): string {
	rev = rev || "";

	// Fast-path: we redirect the tree root to the repo route anyway, so just construct
	// the repo route URL directly.
	if (!path || path === "/" || path.length === 0) {
		return urlToRepoRev(repo, rev);
	}

	const pathStr = typeof path === "string" ? path : path.join("/");
	return urlTo("tree", { splat: [makeRepoRev(repo, rev), pathStr] } as any);
}
