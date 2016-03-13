import "whatwg-fetch";

import context from "sourcegraph/context";

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
	if (typeof document !== "undefined" && document.head.dataset && document.head.dataset.currentUserOauth2AccessToken) {
		let auth = `x-oauth-basic:${document.head.dataset.currentUserOauth2AccessToken}`;
		options.headers["authorization"] = `Basic ${btoa(auth)}`;
	}
	if (context.cacheControl) {
		options.headers["Cache-Control"] = context.cacheControl;
	}
	if (context.parentSpanID) options.headers["Parent-Span-ID"] = context.parentSpanID;
	return options;
}

export function defaultFetch(url, options) {
	let defaults = defaultOptions();

	// Combine headers.
	const headers = Object.assign({}, defaults.headers, options ? options.headers : null);

	return fetch(url, Object.assign(defaults, options, {headers: headers}));
}

// checkStatus is intended to be chained in a fetch call. For example:
//   fetch(...).then(checkStatus) ...
export function checkStatus(resp) {
	if (resp.status === 200 || resp.status === 201) return resp;
	let err = new Error(resp.statusText || `HTTP ${resp.status}`);
	err.response = resp;
	throw err;
}
