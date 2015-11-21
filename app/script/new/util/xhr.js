import defaultXhr from "xhr";

export default function(options, callback) {
	let defaultOptions = {
		headers: {
			"X-Csrf-Token": window._csrfToken,
		},
	};
	if (window.hasOwnProperty("_cacheControl")) {
		defaultOptions.headers["Cache-Control"] = window._cacheControl;
	}
	defaultXhr(Object.assign(defaultOptions, options), callback);
}
