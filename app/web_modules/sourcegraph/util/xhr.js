import {btoa} from "abab";
import "whatwg-fetch";

import context from "sourcegraph/app/context";

// This file provides a common entrypoint to the fetch API.
//
// Use the fetch API (not XHR) because it is the future standard and because
// we can intercept calls to fetch in the reactbridge to render React
// components on the server even if they fetch external data.

function defaultOptions() {
	let options = {
		headers: {
			"X-Csrf-Token": context.csrfToken,
			"X-Device-Id": context.deviceID,
		},
		credentials: "same-origin",
	};
	if (context.authorization) {
		let auth = `x-oauth-basic:${context.authorization}`;
		options.headers["authorization"] = `Basic ${btoa(auth)}`;
	}
	if (context.cacheControl) {
		options.headers["Cache-Control"] = context.cacheControl;
	}
	if (context.currentSpanID) options.headers["Parent-Span-ID"] = context.currentSpanID;
	return options;
}

const allFetches = [];
export function defaultFetch(url, options) {
	if (typeof global !== "undefined" && global.process && global.process.env.JSSERVER) {
		url = `${context.appURL}${url}`;
	}

	let defaults = defaultOptions();

	// Combine headers.
	const headers = Object.assign({}, defaults.headers, options ? options.headers : null);

	const f = fetch(url, Object.assign(defaults, options, {headers: headers}));
	allFetches.push(f);
	return f;
}

// allFetchesCount returns the total number of fetches initiated.
//
// Only this count, not the allFetches list itself, is exported, for
// better encapsulation.
export function allFetchesCount(): number { return allFetches.length; }

// checkStatus is intended to be chained in a fetch call. For example:
//   fetch(...).then(checkStatus) ...
export function checkStatus(resp) {
	if (resp.status === 200 || resp.status === 201) return resp;
	return resp.text().then((body) => {
		let err = new Error(body || resp.statusText);
		err.body = body;
		err.response = resp;
		if (typeof document === "undefined") {
			// Don't log in the browser because the devtools network inspector
			// makes it easy enough to see failed HTTP requests.
			console.error(`HTTP fetch failed with status ${resp.status} ${resp.statusText}: ${resp.url}: ${body}`);
		}
		throw err;
	});
}

// allFetchesResolved returns a promise that is resolved when all fetch calls
// so far are resolved. It lets server.js determine when the initial data
// loading is complete.
export function allFetchesResolved() { return Promise.all(allFetches); }
