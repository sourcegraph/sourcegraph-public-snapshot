var $ = require("jquery");
require("../bower_components/jquery-pjax/jquery.pjax");

document.addEventListener("DOMContentLoaded", function() {
	if ($.support.pjax) {
		$.pjax.defaults.timeout = 0;

		$(document).on("submit", "form[data-pjax]", function(event) {
			var $e = $(this);
			var $container = $($e.attr("data-pjax"));
			$.pjax.submit(event, {container: $container});
			event.preventDefault();
		});

		$(document).on("pjax:send", function(ev, _, opt) {
			if (opt.noLoadingIndicator) return;
			$("#loading").show();
			$(ev.target).addClass("pjax-loading");
		});

		$(document).on("pjax:complete", function(ev, _, opt) {
			if (opt.noLoadingIndicator) return;
			$("#loading").hide();
			$(ev.target).removeClass("pjax-loading");
		});
	}
});
