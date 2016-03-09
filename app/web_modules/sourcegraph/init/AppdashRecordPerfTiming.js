import xhr from "sourcegraph/util/xhr";

// TODO(slimsag): for finer-grained access consider sending all of the info in
// performance.timing to Appdash for display (when available). This would narrow
// down DNS lookup time, DOM load time, redirection time, etc (right now we just
// have page load time, inclusive of everything).

// record sends the relevant start and end times to the server to associate with
// the Appdash trace for the page.
function record() {
	// Not all browsers (e.g., mobile) support this, but most do.
	if (typeof performance === "undefined") return;
	const startTime = performance.timing.fetchStart;
	const endTime = performance.timing.domComplete;

	// At this point the page is considered loaded fully, so we send a POST
	// request to the server in order to trace the time it took to load the
	// page.
	const loadTimeSeconds = (endTime-startTime) / 1000;
	const currentRoute = document.head.dataset.currentRoute;
	const templateName = document.head.dataset.templateName;
	xhr({
		uri: `/.ui/.appdash/upload-page-load?S=${startTime}&E=${endTime}&Route=${currentRoute}&Template=${templateName}`,
		method: "post",
	}, (err, resp) => {
		if (!err && resp.statusCode !== 200) err = `HTTP ${resp.statusCode}`;
		if (err) console.error("appdash: recording page load:", err);
	});

	// Update the debug display on the page with the time.
	document.querySelector("body>#debug>a").text = `${loadTimeSeconds}s`;
}

if (typeof document !== "undefined" && document.head.dataset.appdashCurrentSpanId) {
	document.addEventListener("readystatechange", () => {
		if (document.readyState === "complete") record();
	});
}
