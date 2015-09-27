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
