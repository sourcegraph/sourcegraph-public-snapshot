var $ = require("jquery");

// Authenticate AJAX HTTP requests as the logged-in user (if any),
// using their OAuth2 token.
document.addEventListener("DOMContentLoaded", function() {
	if (document.head.dataset.currentUserOauth2AccessToken) {
		var cred = `x-oauth-basic:${document.head.dataset.currentUserOauth2AccessToken}`;
		$.ajaxSetup({
			headers: {Authorization: `Basic ${btoa(cred)}`},
		});
	}
});

// Set up listeners for oauth2 client actions.
$(document).ready(function() {
	$("#continue-oauth").click(function() {
		popupWindow($(this).attr("data-url"), "Sourcegraph Authentication", 600, 500);
		return false;
	});

	$("#return-oauth").click(function() {
		if (window.opener) {
			window.opener.location.href = $(this).attr("data-url");
			window.close();
		} else {
			window.location.href = $(this).attr("data-url");
		}
		return false;
	});

	// Trigger a redirect to the local page after 1 second.
	if ($("#return-oauth").length) {
		setInterval(function() {
			$("#return-oauth").trigger("click");
		}, 1000);
	}
});

// Open a new window with given url, centered on the parent window.
function popupWindow(url, title, w, h) {
	var wLeft = window.screenLeft ? window.screenLeft : window.screenX;
	var wTop = window.screenTop ? window.screenTop : window.screenY;

	// Center the popup on the parent window
	var left = wLeft + (window.innerWidth / 2) - (w / 2);
	var top = wTop + (window.innerHeight / 2) - (h / 2);
	return window.open(url, title, `modal=yes, alwaysRaised=yes, copyhistory=no, width=${w}, height=${h}, top=${top}, left=${left}`);
}
