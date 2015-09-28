document.addEventListener("DOMContentLoaded", function() {
	var btns = document.querySelectorAll("button.loading-indicator");
	for (var i = 0; i < btns.length; i++) {
		addLoadingIndicatorClickListener(btns[i]);
	}
});

function addLoadingIndicatorClickListener(element) {
	element.addEventListener("click", function(e) {
		var target = e.target;
		while (target && target.tagName !== "BUTTON") target = target.parentNode;
		var origHTML = target.innerHTML;
		target.innerHTML = "<i class=\"fa fa-refresh fa-spin\"></i> " + target.dataset.loadingText;
		target.classList.add("disabled");

		if (target.dataset.loadingTime) {
			setTimeout(function() {
				target.innerHTML = origHTML;
				target.classList.remove("disabled");
			}, parseInt(target.dataset.loadingTime, 10));
		}
	});
}
