document.addEventListener("DOMContentLoaded", function() {
	var btn = document.getElementById("check-all-repos");
	if (btn !== null) {
		btn.addEventListener("change", function(e) {
			var checkboxes = document.getElementsByName("RepoURI[]");
			for (var i=0, n=checkboxes.length; i<n; i++) {
				checkboxes[i].checked = e.target.checked;
			}
		});
	}
	return;
});
