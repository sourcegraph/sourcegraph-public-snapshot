import { SourcegraphURL, BlobURL } from "util/types";

export function parse(loc: Location = window.location): SourcegraphURL {

	const urlsplit = loc.pathname.slice(1).split("/");
	if (urlsplit.length < 3 && urlsplit[0] !== "github.com") {
		return {};
	}

	let uri = urlsplit.slice(0, 3).join("/");
	let rev: string | undefined;
	let path: string | undefined;
	const uriSplit = uri.split("@");
	if (uriSplit.length > 0) {
		uri = uriSplit[0];
		rev = uriSplit[1];
	}

	if (loc.pathname.indexOf("/-/blob/") !== -1) {
		path = urlsplit.slice(5).join("/");
	}

	return { uri, rev, path };
}

export function toBlob(loc: BlobURL): string {
	let url = `/${loc.uri}${loc.rev ? "@" + loc.rev : ""}/-/blob/${loc.path}`;
	if (loc.line) { // construct hash w/ format #L[line][:char][$modal[:mode]]
		url += "#L" + loc.line;
		if (loc.char) {
			url += ":" + loc.char;
		}
		if (loc.modal) {
			url += `$${loc.modal}`;
			if (loc.modalMode) {
				url += `:${loc.modalMode}`;
			}
		}
	}
	return url;
}
