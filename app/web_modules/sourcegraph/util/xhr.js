import defaultXhr from "xhr";

import context from "sourcegraph/context";

export default function(options, callback) {
	let defaultOptions = {
		headers: {
			"X-Csrf-Token": context.csrfToken,
			"X-Device-Id": context.deviceID,
		},
	};
	if (typeof document !== "undefined" && document.head.dataset && document.head.dataset.currentUserOauth2AccessToken) {
		let auth = `x-oauth-basic:${document.head.dataset.currentUserOauth2AccessToken}`;
		defaultOptions.headers["authorization"] = `Basic ${btoa(auth)}`;
	}
	if (context.cacheControl) {
		defaultOptions.headers["Cache-Control"] = context.cacheControl;
	}
	if (context.parentSpanID) defaultOptions.headers["Parent-Span-ID"] = context.parentSpanID;
	defaultXhr(Object.assign(defaultOptions, options), callback);
}
