var $ = require("jquery");

document.addEventListener("DOMContentLoaded", function() {
	$(document.body).tooltip({
		selector: "[data-tooltip]",
		placement(_, node) {
			var pos = node.getAttribute("data-tooltip");
			return (!pos || pos === "true") ? "top" : pos;
		},
	});
});
