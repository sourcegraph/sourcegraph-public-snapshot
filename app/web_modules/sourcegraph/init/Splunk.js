// Splunk error monitoring code

import "whatwg-fetch";
import context from "sourcegraph/app/context";

let appIsHandlingError = false;

// globalErrorHandler is called when global JS errors occur, once the error
// has been fully handled, appIsHandlingError must be set to false for any
// future errors to be handled.
//
// ev is a user-defined object (Splunk makes no assumptions about the data),
// and should include all relevant informatino to debug the error.
function globalErrorHandler(ev) {
	let opts = {
		method: "POST",
		headers: new Headers({
			"Authorization": `Splunk D70E82E5-34CC-4DFA-A08A-E7FA115FB45B`,
			"Accept": "application/json",
			"Content-Type": "application/json",
		}),
		body: JSON.stringify({
			source: context.currentUser ? context.currentUser.Login : "anonymous",
			sourcetype: "sourcegraph-frontend",
			event: ev,
		}),
	};
	fetch("https://splunk-ext.sourcegraph.com:8088/services/collector/event/1.0", opts)
		.then((response) => {
			if (response.status >= 200 && response.status < 300) {
				return response;
			}
			let error = new Error(response.statusText);
			error.response = response;
			throw error;
		}).catch((err) => {
			console.log("Splunk: error", err);
		}).then(function(data) {
			// Request succeeded.
			appIsHandlingError = false;
		});
}

if (typeof window !== "undefined") {
	// Register the global error handler, being careful to not accidently
	// recurse infinitely should the globalErrorHandler itself throw an error
	// due to e.g. a bug in splunk-logging, etc.
	//
	// Do not use a try-catch because globalErrorHandler could use asynchronous
	// code and still end up back in window.onerror.
	window.onerror = function(message, source, line, column, jserr) {
		if (!appIsHandlingError) {
			appIsHandlingError = true;

			// Define the event.
			let ev = {
				message: message,
				error: jserr,

				// Include a plaintext stacktrace.
				stackTrace: new Error().stack,

				// Add in general runtime/browser info.
				browser: {
					location: window.location.href,
					userAgent: navigator.userAgent,
				},
			};

			// Add in various tags from template data (deployed commit, user info, etc).
			for (let k in window._splunkTags) {
				if (window._splunkTags.hasOwnProperty(k)) {
					ev[k] = window._splunkTags[k];
				}
			}

			globalErrorHandler(ev);
		}
	};
}
