// sourcegraph_activateDefnPopovers activates definition popovers for all
// "a.defn-popover" child elements of el. It requires manual activation and
// doesn't just run on DOMContentLoaded because this same function is used in
// our Chrome extension (this file is copied over), where we want to be able to
// manually trigger it.
//
// TODO(x): make sure this works in the chrome ext and in sourceboxes
// after switching to webpack.
function sourcegraph_activateDefnPopovers(el, baseURL) {
	var activeA;
	el.addEventListener("mouseover", function(ev) {
		var t = getTarget(ev.target);
		if (!t) return;
		if (activeA !== t) {
			activeA = t;
			var url = activeA.classList.contains("ann") ? activeA.dataset.popover : (activeA.href + "/.popover");
			ajaxGet(url, function(html) {
				if (activeA) showPopover(html);
			});
		}
	});
	el.addEventListener("mouseout", function(ev) {
		var t = getTarget(ev.toElement);
		if (!t) {
			setTimeout(hidePopover);
			activeA = null;
		}
	});

	function getTarget(t) {
		while (t && (t.tagName === "SPAN" || t.tagName === "TT")) { t = t.parentNode; }
		if (t && t.tagName === "A" && t.classList.contains("defn-popover")) return t;
	}

	var popover;
	var showingPopover;
	var preShowingX, preShowingY;
	function showPopover(html) {
		showingPopover = true;
		if (!popover) {
			popover = document.createElement("div");
			popover.classList.add("sourcegraph-popover");
			popover.style.position = "absolute";
			document.body.appendChild(popover);
		}
		positionPopover(preShowingX, preShowingY);
		popover.innerHTML = html;
		popover.classList.add("visible");
		popover.style.display = "block";
	}

	function hidePopover() {
		if (!popover) return;
		showingPopover = false;
		setTimeout(function() {
			if (!showingPopover) popover.style.display = "none";
		}, 200);
		popover.classList.remove("visible");
	}

	document.addEventListener("mousemove", function(ev) {
		preShowingX = ev.pageX;
		preShowingY = ev.pageY;
		positionPopover(ev.pageX, ev.pageY);
	});

	function positionPopover(x, y) {
		if (!popover || !showingPopover) return;
		popover.style.top = (y + 15) + "px";
		popover.style.left = (x + 15) + "px";
	}

	var ajaxCache = {};
	function ajaxGet(url, cb) {
		if (ajaxCache[url]) {
			cb(ajaxCache[url]);
			return;
		}
		var request = new XMLHttpRequest();
		request.open("GET", url, true);
		request.onload = function() {
			if (request.status >= 200 && request.status < 400) {
				ajaxCache[url] = request.responseText;
				cb(request.responseText);
			} else if (request.readyState > 1) {
				console.error("Sourcegraph error getting definition info.", JSON.stringify(request));
			}
		};
		request.onerror = function() { console.error("Sourcegraph error getting definition info."); };
		request.send();
	}
}

module.exports = sourcegraph_activateDefnPopovers;
window.sourcegraph_activateDefnPopovers = sourcegraph_activateDefnPopovers;
