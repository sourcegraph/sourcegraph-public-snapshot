var $ = require("jquery");
var router = require("./routing/router");

// TODO(slimsag): for finer-grained access consider sending all of the info in
// performance.timing to Appdash for display (when available). This would narrow
// down DNS lookup time, DOM load time, redirection time, etc (right now we just
// have page load time, inclusive of everything).

// Grab the start time now so that we begin measuring at page load, not at DOM
// load (since setupMeasurement is called after DOMContentLoaded has fired).
var startTime = null;
if (performance) {
	// Not all browsers (e.g. mobile) support this but most do, and it is much
	// more accurate in representing the entire page load time.
	startTime = performance.timing.fetchStart;
} else {
	// We don't have performance.timing, so we can fallback to getTime which is
	// off by quite a bit (does not include network time, DOM load time, or
	// anything else that occured before this code runs). Still better than
	// nothing.
	startTime = new Date().getTime();
}

// setupMeasurement sets up event handlers to identify the "end" of AJAX
// requests in the page (i.e. when the page is _fully_ loaded, not just the DOM)
// and sends the relevant start and end times to the server to associate with
// the Appdash trace for the page.
function setupMeasurement() {
	var delay = 1000;
	var measured = false;
	var endTime = null;

	var measure = function() {
		// At this point the page is considered loaded fully, so we send a POST
		// request to the server in order to trace the time it took to load the
		// page.
		measured = true;
		var loadTimeSeconds = (endTime-startTime) / 1000;
		$.ajax({
			url: router.appdashUploadPageLoadURL(startTime, endTime),
			method: "post",
			headers: {"X-CSRF-Token": document.head.dataset.csrfToken},
		});

		// Update the debug display on the page with the time.
		$("body>#debug>a").html(loadTimeSeconds + "s");
	};

	// When all AJAX requests stop, start the timer.
	var timeout = null;
	$(document).ajaxStop(function() {
		// Store the time at which the last AJAX request ended.
		endTime = new Date().getTime();

		// Clear any previous timeout just to be safe.
		clearTimeout(timeout);

		// Only set a new timeout if we haven't yet measured the page load time.
		if (!measured) {
			timeout = setTimeout(measure, delay);
		}
	});

	// When an AJAX request begins, stop the timer.
	$(document).ajaxStart(function() {
		clearTimeout(timeout);
	});
}

/**
 * @description Invokes the AppDash setup by configuring headers on consequent AJAX
 * requests.
 * @returns {void}
 */
 document.addEventListener("DOMContentLoaded", function() {
	if (document.head.dataset.appdashCurrentSpanId) {
		$.ajaxSetup({
			headers: {"Parent-Span-ID": document.head.dataset.appdashCurrentSpanId},
		});

		setupMeasurement();
	}
});
