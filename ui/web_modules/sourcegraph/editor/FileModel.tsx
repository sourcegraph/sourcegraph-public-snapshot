// tslint:disable typedef ordered-imports
import * as invariant from "invariant";
import {urlToBlob, parseBlobURL} from "sourcegraph/blob/routes";
import {RangeOrPosition} from "sourcegraph/core/rangeOrPosition";

export function uriForTreeEntry(repo: string, rev: string | null, path: string, range?: monaco.IRange): monaco.Uri {
	return monaco.Uri.from({
		scheme: "sourcegraph",
		path: urlToBlob(repo, rev, path),
		fragment: range ? `L${RangeOrPosition.fromMonacoRange(range).toString()}` : undefined,
	});
}

export function treeEntryFromUri(uri: monaco.Uri): {repo: string, rev: string | null, path: string, range?: monaco.IRange} {
	invariant(uri.scheme === "sourcegraph", `unexpected uri scheme: ${uri}`);
	const u = parseBlobURL(`${uri.path}${uri.fragment ? `#${uri.fragment}` : ""}`);

	let range: monaco.IRange | undefined;
	if (u.hash) {
		const r = RangeOrPosition.parse(u.hash.replace(/^L/, ""));
		if (r) {
			range = r.toMonacoRangeAllowEmpty();
		}
	}

	return {
		repo: u.repo,
		rev: u.rev,
		path: u.path,
		range,
	};
}
