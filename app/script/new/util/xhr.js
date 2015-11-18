import defaultXhr from "xhr";

export default function(options, callback) {
	let defaultOptions = {
		headers: {
			"X-Csrf-Token": window._csrfToken,
		},
	};
	defaultXhr(Object.assign(defaultOptions, options), callback);
}
