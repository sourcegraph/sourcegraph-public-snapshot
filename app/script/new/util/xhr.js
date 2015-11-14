import defaultXhr from "xhr";

let defaultOptions = {
	headers: {
		"X-Csrf-Token": window._csrfToken,
	},
};

export default function(options, callback) {
	defaultXhr(Object.assign(defaultOptions, options), callback);
}
