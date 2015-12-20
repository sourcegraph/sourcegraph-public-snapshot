import defaultXhr from "xhr";

import context from "sourcegraph/context";

export default function(options, callback) {
	let defaultOptions = {
		headers: {
			"X-Csrf-Token": context.csrfToken,
		},
	};
	if (window.hasOwnProperty("_cacheControl")) {
		defaultOptions.headers["Cache-Control"] = window._cacheControl;
	}
	defaultXhr(Object.assign(defaultOptions, options), callback);
}
