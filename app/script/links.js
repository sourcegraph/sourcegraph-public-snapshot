require("../style/_links.scss");

document.addEventListener("DOMContentLoaded", function() {
	function triggerLink(elem, ev) {
		var href = elem.dataset.href;
		if (!href) { // not a data-href element
			return;
		}

		var target = "_self";
		if (ev.button === 1 || ev.ctrlKey || ev.metaKey || ev.shiftKey) {
			target = "_blank";
		}
		window.open(href, target);
		ev.preventDefault();
	}

	document.addEventListener("click", function(ev) {
		var el = ev.target;
		while (el && el.dataset && !el.dataset.href) el = el.parentNode;
		if (el && el.dataset && el.dataset.href) triggerLink(el, ev);
	});

	document.addEventListener("keydown", function(ev) {
		if (ev.keyCode === 13) { // <RETURN> or <ENTER> key
			triggerLink(ev.target, ev);
		}
	});
});
