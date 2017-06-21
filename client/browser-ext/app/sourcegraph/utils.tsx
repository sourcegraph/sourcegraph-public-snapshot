import { getDomain } from "../utils";
import { Domain, SourcegraphURL } from "../utils/types";

export function parseURL(loc: Location = window.location): SourcegraphURL {
	const domain = getDomain(loc);
	if (domain !== Domain.SOURCEGRAPH) {
		return {};
	}

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
