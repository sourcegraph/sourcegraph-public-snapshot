import defaultXhr from "xhr";

import context from "sourcegraph/context";

export default function(options, callback) {
	if (typeof document === "undefined") {
		// On the server (in the Duktape JS VM), there is no XHR. This HTTP request
		// will not be issued nor satisfied, but that's OK, since the client will
		// pick up where the server left off by issuing the same request.
		//
		// Calling XHR on the server indicates that data was not preloaded. This is
		// OK, but you can improve performance by preloading the necessary data. NOTE:
		// manual preloading of data will no longer be necessary in the upcoming
		// pure-react branch (this was written on 2016 Mar 28).
		return;
	}

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
