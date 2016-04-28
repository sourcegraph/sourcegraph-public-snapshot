// @flow

import {btoa} from "abab";
import "whatwg-fetch";

import context from "sourcegraph/app/context";
import UserStore from "sourcegraph/user/UserStore";

// This file provides a common entrypoint to the fetch API.

function defaultOptions(): RequestOptions {
	const headers = new Headers();
	if (context.csrfToken) headers.set("X-Csrf-Token", context.csrfToken);
	if (UserStore.activeAccessToken) {
		let auth = `x-oauth-basic:${UserStore.activeAccessToken}`;
		headers.set("Authorization", `Basic ${btoa(auth)}`);
	}
	if (context.cacheControl) headers.set("Cache-Control", context.cacheControl);
	if (context.currentSpanID) headers.set("Parent-Span-ID", context.currentSpanID);
	return {
		headers,
		credentials: "same-origin",
	};
}

export function combineHeaders(a: any, b: any): any {
	// NOTE(sqs): Flow gave a lot of weird "inconsistent use of library definitions" errors
	// when I tried to use the Headers and HeadersInit types here. This has a unit test,
	// so leave these as "any" types for now.
	if (!b) return a;
	if (!a) return b;

	if (!(a instanceof Headers)) throw new Error("must be Headers type");
	if (!(b instanceof Headers)) throw new Error("must be Headers type");

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
export function defaultFetch(url: string | Request, init?: RequestOptions): Promise<Response> {
	if (typeof url !== "string") throw new Error("url must be a string (complex requests are not yet supported)");

	const defaults = defaultOptions();

	return fetch(url, {
		...defaults,
		...init,
		headers: combineHeaders(defaults.headers, init ? init.headers : null),
	});
}

// checkStatus is intended to be chained in a fetch call. For example:
//   fetch(...).then(checkStatus) ...
export function checkStatus(resp: Response): Promise<Response> | Response {
	if (resp.status >= 200 && resp.status <= 299) return resp;
	return resp.text().then((body) => {
		if (typeof document === "undefined") {
			// Don't log in the browser because the devtools network inspector
			// makes it easy enough to see failed HTTP requests.
			console.error(`HTTP fetch failed with status ${resp.status} ${resp.statusText}: ${resp.url}: ${body}`);
		}
		let err: any;
		try {
			err = {...(new Error(resp.status)), body: JSON.parse(body)};
		} catch (error) {
			err = {...(new Error(resp.statusText)),
				body: body,
				response: {status: resp.status, statusText: resp.statusText, url: resp.url},
			};
		}
		throw err;
	});
}
