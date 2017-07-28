import { SourcegraphURL, BlobURL } from "util/types";
import * as urlparse from "url-parse";

// parse parses a generic Sourcegraph URL, where most components are shared
// across all routes, e.g. repo URI and rev.
export function parse(_loc: String = window.location.href): SourcegraphURL {
	const loc = urlparse(_loc);
	const urlsplit = loc.pathname.slice(1).split("/");
	if (urlsplit.length < 3 && urlsplit[0] !== "github.com") {
		return {};
	}

	let uri = urlsplit.slice(0, 3).join("/");
	let rev: string | undefined;
	const uriSplit = uri.split("@");
	if (uriSplit.length > 0) {
		uri = uriSplit[0];
		rev = uriSplit[1];
	}
	return { uri, rev };
}

// parseBlob parses a blob page URL.
export function parseBlob(_loc: String = window.location.href): BlobURL {
	const loc = urlparse(_loc);
	// Parse the generic Sourcegraph URL
	const u = parse(loc);

	// Parse blob-specific URL components.
	const urlsplit = loc.pathname.slice(1).split("/");
	if (urlsplit.length < 3 && urlsplit[0] !== "github.com") {
		return {};
	}
	let path: string | undefined;
	if (loc.pathname.indexOf("/-/blob/") !== -1) {
		path = urlsplit.slice(5).join("/");
	}
	let v: BlobURL = { ...u, path };

	const lineCharModalInfo = loc.hash.split("$"); // e.g. "#L17:19$references:external"
	if (lineCharModalInfo[0]) {
		const coords = lineCharModalInfo[0].split("#L")[1].split(":");
		v.line = parseInt(coords[0], 10); // 17
		v.char = parseInt(coords[1], 10); // 19
	}
	if (lineCharModalInfo[1]) {
		const modalInfo = lineCharModalInfo[1].split(":");
		v.modal = modalInfo[0]; // "references"
		v.modalMode = modalInfo[1]; // "external"
	}
	return v;
}

export function toBlob(loc: BlobURL): string {
	return `/${loc.uri}${loc.rev ? "@" + loc.rev : ""}/-/blob/${loc.path}$(toBlobHash(loc))`;
}

export function toBlobHash(loc: BlobURL): string {
	let hash = "";
	if (loc.line) { // construct hash w/ format #L[line][:char][$modal[:mode]]
		hash += "#L" + loc.line;
		if (loc.char) {
			hash += ":" + loc.char;
		}
		if (loc.modal) {
			hash += `$${loc.modal}`;
			if (loc.modalMode) {
				hash += `:${loc.modalMode}`;
			}
		}
	}
	return hash;
}
