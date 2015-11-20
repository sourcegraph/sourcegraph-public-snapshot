/**
 * @description Set the cache header on all subsequent AJAX requests to match
 * the data-cache-control attribute in the head tag of the current page.
 * @returns {void}
 */
 document.addEventListener("DOMContentLoaded", function() {
	if (document.head.dataset.cacheControl) {
		$.ajaxSetup({
			headers: {"Cache-Control": document.head.dataset.cacheControl},
		});
	}
});