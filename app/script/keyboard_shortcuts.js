var $ = require("jquery");
require("jquery.hotkeys/jquery.hotkeys");

document.addEventListener("DOMContentLoaded", function() {
	// "/" focuses search
	$(document).on("keydown", null, "/", function(e) {
		if ($(e.target).is("input")) return;
		setTimeout(function() {
			$("#nav input.search-input").focus();
		});
	});

	// "?" opens the keyboard shortcuts help screen
	$(document).bind("keyup", function(e) {
		if (e.keyCode === 191 && !$(e.target).is("input") && !$(e.target).is("textarea")) {
			$("#keyboard-shortcuts-help").modal({show: true}).modal("show");
		}
	});

	$(document).on("keydown", null, "g", function(e) {
		if ($(e.target).is("input")) return;

		var deactivate = [];

		// "g + r" goes to repository
		if (document.body.dataset.currentRepoUrl) {
			$(document).on("keydown", null, "r", function() { window.location.href = document.body.dataset.currentRepoUrl; });
			deactivate.push("r");
		}

		// "g + h" goes to homepage
		$(document).on("keydown", null, "h", function() { window.location.href = "/"; });
		deactivate.push("h");

		setTimeout(function() {
			deactivate.forEach(function(k) {
				$(document).off("keydown", null, k);
			});
		}, 1500);
	});
});
