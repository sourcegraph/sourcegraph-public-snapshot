// build_log continuously refreshes build logs on a build page.

document.addEventListener("DOMContentLoaded", function() {
	var $log = document.getElementById("build-log");
	if ($log) {
		continuouslyRefresh($log);
	}
});

function continuouslyRefresh($el) {
	updateLogView($el.dataset.src, $el);
}

function updateLogView(href, $e) {
	var maxID = $e.dataset.maxId;
	var reqHref = href;
	if (maxID) reqHref += `?MinID=${maxID}`;

	var request = new XMLHttpRequest();
	request.open("GET", reqHref, true);

	request.onload = function() {
		if (request.status >= 200 && request.status < 400) {
			if (request.responseText) {
				$e.textContent = request.responseText;

				var maxIDHeader = request.getResponseHeader("X-Sourcegraph-Log-Max-Id");
				if (maxIDHeader) $e.dataset.maxId = maxIDHeader;
			}

			setTimeout(function() { updateLogView(href, $e); }, 2000);
		} else if (request.readyState > 1) {
			$e.innerText = "Error loading log.";
		}
	};

	request.onerror = function() {
		$e.innerText = "Failed to load log.";
	};

	request.send();
}
