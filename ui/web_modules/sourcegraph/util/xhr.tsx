import "whatwg-fetch";

import {context} from "sourcegraph/app/context";

// This file provides a common entrypoint to the fetch API.

export function combineHeaders(a: any, b: any): any {
	if (!b) { return a; }
	if (!a) { return b; }

	if (!(a instanceof Headers)) { throw new Error("must be Headers type"); }
	if (!(b instanceof Headers)) { throw new Error("must be Headers type"); }

	if (b.forEach) {
		// node-fetch's Headers is not a full implementation and doesn't support iterable,
		// but it does expose forEach.
		b.forEach((val: string, name: string) => a.append(name, val));
	} else {
		for (let [name, val] of b) {
			a.append(name, val);
		}
	}
	return a;
}

// defaultFetch wraps the fetch API.
//
// Note: the caller might wrap this with singleflightFetch.
export function defaultFetch(url: string | Request, init?: RequestInit): Promise<Response> {
	if (typeof url !== "string") { throw new Error("url must be a string (complex requests are not yet supported)"); }

	let defaultHeaders = new Headers();
	Object.keys(context.xhrHeaders).forEach((key) => {
		defaultHeaders.set(key, context.xhrHeaders[key]);
	});

	return fetch(url, {
		method: (init && init.method) || "GET",
		headers: combineHeaders(defaultHeaders, init ? init.headers : null),
		body: init && init.body,
		mode: init && init.mode,
		redirect: init && init.redirect,
		credentials: (init && init.credentials) || "same-origin",
		cache: init && init.cache,
	});
}

// checkStatus is intended to be chained in a fetch call. For example:
//   fetch(...).then(checkStatus) ...
export function checkStatus(resp: Response): Promise<Response> | Response {
	if (resp.status >= 200 && resp.status <= 299) { return resp; }
	return resp.text().then((body) => {
		if (typeof document === "undefined") {
			// Don't log in the browser because the devtools network inspector
			// makes it easy enough to see failed HTTP requests.
			console.error(`HTTP fetch failed with status ${resp.status} ${resp.statusText}: ${resp.url}: ${body}`);
		}
		let err: Error;
		try {
			err = new Error(resp.status.toString());
			(err as any).body = JSON.parse(body);
		} catch (error) {
			err = new Error(resp.statusText);
			(err as any).body = body;
			(err as any).response = {status: resp.status, statusText: resp.statusText, url: resp.url};
		}
		throw err;
	});
}
