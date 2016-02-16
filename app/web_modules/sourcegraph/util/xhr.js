import defaultXhr from "xhr";

import context from "sourcegraph/context";

export default function(options, callback) {
	let defaultOptions = {
		headers: {
			"X-Csrf-Token": context.csrfToken,
		},
	};
	if (typeof document !== "undefined" && document.head.dataset && document.head.dataset.currentUserOauth2AccessToken) {
		let auth = `x-oauth-basic:${document.head.dataset.currentUserOauth2AccessToken}`;
		defaultOptions.headers["authorization"] = `Basic ${btoa(auth)}`;
	}
	if (window.hasOwnProperty("_cacheControl")) {
		defaultOptions.headers["Cache-Control"] = window._cacheControl;
	}
	defaultXhr(Object.assign(defaultOptions, options), callback);
}
