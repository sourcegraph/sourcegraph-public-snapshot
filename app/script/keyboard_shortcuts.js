var $ = require("jquery");
require("jquery.hotkeys/jquery.hotkeys");

document.addEventListener("DOMContentLoaded", function() {
	// "/" focuses search
	$(document).on("keydown", null, "/", function(e) {
		if ($(e.target).is("input")) return;
		setTimeout(function() {
			$("#nav .tt-input").focus();
		});
	});

	// "?" opens the keyboard shortcuts help screen
	$(document).bind("keyup", function(e) {
		if (e.keyCode === 191 && !$(e.target).is("input") && !$(e.target).is("textarea")) {
			$("#keyboard-shortcuts-help").modal({show: true}).modal("show");
		}
	});

	// "o" opens the thing that is focused
	var openAction = function(e) {
		var href = $(":focus").data("href");
		if (href) {
			window.location.href = href;
		}
	};
	$(document).on("keydown", null, "o", openAction);
	$(document).on("keydown", function(e) {  // <ENTER>
		if (e.keyCode === 13) {
			openAction(e);
		}
	});

	// Up/down arrow keys should navigate among search results, and allow
	// navigation into and out of the search text field.
	var nextPrevAction = function(e) {
		var allResults = $(".result");
		var cur = allResults.index(e.target); // index of current result (before keypress)
		if (e.keyCode === 38) { // <UPARROW>
			var prev;
			if (cur === 0) {
				// Up-arrow when focused on the *first* result should go to the search
				// field.
				prev = $(".search-form input[name=q]");
			} else {
				// Up-arrow should go to the previous result, even in a different result
				// type group (repos/people/defs/etc.).
				prev = allResults[cur - 1];
			}
			prev.focus();
			e.preventDefault();
		} else if (e.keyCode === 40) { // <DOWNARROW>
			var next;
			if (cur === allResults.length - 1) {
				// Down-arrow when focused on the *last* should go to the search
				// field.
				next = $(".search-form input[name=q]");
			} else {
				// Up-arrow should go to the next result, even in a different result
				// type group (repos/people/defs/etc.).
				next = allResults[cur + 1];
			}
			next.focus();
			e.preventDefault();
		}
	};
	$(document).on("keydown", ".result", null, nextPrevAction);
	// $(document).on("keydown", ".search-form input[name=q]", null, function(e) {
	//   if (e.keyCode == 40) { // <DOWNARROW>
	//     // Down-arrow when focused on a search field should go to the first result.
	//     $(".result").first().focus();
	//     e.preventDefault();
	//   } else if (e.keyCode == 38) { // <UPARROW>
	//     // Up-arrow when focused on a search field should go to the last result.
	//     $(".result").last().focus();
	//     e.preventDefault();
	//   }
	// });

	// "g + r" goes to repository
	$(document).on("keydown", null, "g", function(e) {
		if ($(e.target).is("input")) return;

		var deactivate = [];

		if (typeof $currentRepoURL !== "undefined") {
			$(document).on("keydown", null, "r", function() { window.location.href = document.body.dataset.currentRepoUrl; });
			deactivate.push("r");

			$(document).on("keydown", null, "b", function() { window.location.href = document.body.dataset.currentRepoBuildsUrl; });
			deactivate.push("b");
		}

		$(document).on("keydown", null, "h", function() { window.location.href = "/"; });
		deactivate.push("h");

		setTimeout(function() {
			deactivate.forEach(function(k) {
				$(document).off("keydown", null, k);
			});
		}, 1500);
	});
});
