// build_log makes logs open up inline when the user clicks on "Logs" on a build
// page. The log is continuously refreshed.

var updateLogViews = {};

document.addEventListener("DOMContentLoaded", function() {
	var $$a = document.querySelectorAll("a.build-task-logs");
	for (var i = 0; i < $$a.length; i++) {
		addBuildTaskLogClickListener($$a[i]);
	}
});

function addBuildTaskLogClickListener(target) {
	target.addEventListener("click", function(ev) {
		ev.preventDefault();

		var $a = ev.target;
		var $tr = $a.parentNode.parentNode;
		var nextSib = $tr.nextSibling;

		if (nextSib.tagName === "TR") {
			// currently showing; hide it.
			Reflect.deleteProperty(updateLogViews, $a.href);
			nextSib.parentNode.removeChild(nextSib);
			$a.innerText = $a.dataset.origText;
		} else {
			// not currently showing logs; show them.
			$a.dataset.origText = $a.innerText;
			$a.innerText = "Hide";

			var $logTR = document.createElement("tr");
			$logTR.classList.add("log");
			$logTR.innerHTML = "<td colspan=7><div class=log-lines><span class=status>Loading log...</span></div><div class=pull-right><label><input type=checkbox checked> Scroll to bottom</label></div></td>";
			nextSib.parentNode.insertBefore($logTR, nextSib);
			var $e = $logTR.querySelector("div.log-lines");
			updateLogViews[$a.href] = true;
			updateLogView($a.href, $e);
		}
	});
}

var NO_LOGS_FOUND_MESSAGE = "(No logs found. Logs are typically removed after 3 days. Contact support@sourcegraph.com for help.)";

function updateLogView(href, $e) {
	// removal check
	if (!updateLogViews[href]) {
		return;
	}

	var $status = $e.querySelector(".status");

	var maxID = $e.dataset.maxId;
	var reqHref = href;
	if (maxID) reqHref += "?MinID=" + maxID;

	var request = new XMLHttpRequest();
	request.open("GET", reqHref, true);

	request.onload = function() {
		if (request.status >= 200 && request.status < 400) {
			if ($status) $status.innerText = "";

			if (request.responseText) {
				$e.appendChild(document.createTextNode(request.responseText));

				var maxIDHeader = request.getResponseHeader("X-Sourcegraph-Log-Max-Id");
				if (maxIDHeader) $e.dataset.maxId = maxIDHeader;

				var scrollToBottom = $e.parentNode.querySelector("input[type=checkbox]").checked;
				if (scrollToBottom) $e.scrollTop = $e.scrollHeight;
			} else {
				$status.innerText = NO_LOGS_FOUND_MESSAGE;
			}

			setTimeout(function() { updateLogView(href, $e); }, 2000);
		} else if (request.readyState > 1) {
			$status.innerText = "Error loading log.";
		}
	};

	request.onerror = function() {
		$status.innerText = "Failed to load log.";
	};

	request.send();
}
