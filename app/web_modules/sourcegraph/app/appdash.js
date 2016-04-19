// @flow

import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import context from "sourcegraph/app/context";

// recordSpan records a single span (operation) as starting and ending at the
// given unix timestamps (time in milliseconds since the unix epoch) to
// Appdash.
//
// It is important that the name uniquely identify the type of operation, but
// not include the exact contents/details of the operation. For example, use
// the name "/search" but not "/search?q=<some user query>". This is so that
// spans can be aggregated based on their name.
//
// Any potential error that would occur is sent to console.error instead of
// being thrown.
export function recordSpan(name: string, start: number, end: number) {
	defaultFetch(`/.api/internal/appdash/record-span?S=${start}&E=${end}&Name=${name}`, {
		method: "POST",
	})
	.then(checkStatus)
	.catch((err) => {
		console.error("appdash:", err);
	});
}

// recordInitialPageLoad record the initial load time of the page.
//
// TODO(slimsag): for finer-grained access consider sending all of the info in
// performance.timing to Appdash for display (when available). This would
// narrow down DNS lookup time, DOM load time, redirection time, etc. (right
// now we just have page load time, inclusive of everything).
function recordInitialPageLoad() {
	// Not all browsers (e.g., mobile) support this, but most do.
	if (typeof window.performance === "undefined") return;

	// Record the time between when the browser was ready to fetch the
	// document, and when the document.readyState was changed to "complete".
	// i.e., the time it took to load the page.
	const startTime = window.performance.timing.fetchStart;
	const endTime = window.performance.timing.domComplete;
	recordSpan(`load ${window.location.pathname}`, startTime, endTime);

	// Update the debug display on the page with the time.
	let debug = document.querySelector("body>#debug>a");
	const loadTimeSeconds = (endTime-startTime) / 1000;

	// $FlowHack
	if (debug) debug.text = `${loadTimeSeconds}s`;
}


if (typeof document !== "undefined" && context.currentSpanID) { // eslint-disable-line no-undefined
	document.addEventListener("readystatechange", () => {
		if (document.readyState === "complete") recordInitialPageLoad();
	});
}
