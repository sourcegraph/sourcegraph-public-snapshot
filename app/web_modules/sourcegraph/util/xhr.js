// @flow

import {btoa} from "abab";
import "whatwg-fetch";

import context from "sourcegraph/app/context";

// This file provides a common entrypoint to the fetch API.
//
// Use the fetch API (not XHR) because it is the future standard and because
// we can intercept calls to fetch in the reactbridge to render React
// components on the server even if they fetch external data.

function defaultOptions(): RequestOptions {
	const headers = new Headers();
	if (context.csrfToken) headers.set("X-Csrf-Token", context.csrfToken);
	if (context.authorization) {
		let auth = `x-oauth-basic:${context.authorization}`;
		headers.set("Authorization", `Basic ${btoa(auth)}`);
	}
	if (context.cacheControl) headers.set("Cache-Control", context.cacheControl);
	if (context.currentSpanID) headers.set("Parent-Span-ID", context.currentSpanID);
	return {
		headers,
		credentials: "same-origin",

		// Compress requests for browser clients but not for jsserver renderer (which
		// is colocated with the API endpoint, so network is fast).
		compress: typeof document !== "undefined",
	};
}

let _globalBaseURL: string = ""; // private

// setGlobalBaseURL sets the base URL to use for all fetches.
export function setGlobalBaseURL(baseURL: string): void {
	if (baseURL.endsWith("/")) throw new Error("base URL must not have trailing slash");
	_globalBaseURL = baseURL;
}

// defaultFetch wraps the fetch API.
//
// Note: the caller might wrap this with singleflightFetch.
export function defaultFetch(url: string | Request, options?: RequestOptions): Promise<Response> {
	if (typeof url !== "string") throw new Error("url must be a string (complex requests are not yet supported)");
	if (typeof global !== "undefined" && global.process && global.process.env.JSSERVER) {
		url = `${_globalBaseURL}${url}`;
	}

	let defaults = defaultOptions();

	// Combine headers.
	const headers = Object.assign({}, defaults.headers, options ? options.headers : null);

	return fetch(url, Object.assign(defaults, options, {headers: headers}));
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
