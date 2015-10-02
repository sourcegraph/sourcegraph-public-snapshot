var $ = require("jquery");

// Feedback form scripts may depend on using information from the current
// window's href. During a push state event this function will reload them
// so they can access the new href.
$(window).on("sg:pushState popstate", () => {
	var $feedbackForm = $("#custom-feedback-form");
	if ($feedbackForm.length > 0) {
		var feedbackFormScripts = $("#custom-feedback-form").html();
		$("#custom-feedback-form").html(feedbackFormScripts);
	}
});
