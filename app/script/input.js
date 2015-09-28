var $ = require("jquery");

// Shim to make sure that cursor is always at the end of an auto-focused input after page load
$(function() {
	var focusedInput = document.activeElement;
	if (document.activeElement && document.activeElement.tagName.toLowerCase() === "input") {
		focusedInput = document.activeElement;
	} else {
		var autofocus = document.querySelector("input[autofocus]");
		if (autofocus) {
			focusedInput = autofocus;
		}
	}
	if (focusedInput) {
		focusedInput.selectionStart = (focusedInput.value || "").length;
	}
});
