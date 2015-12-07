var client = require("./client");

document.addEventListener("DOMContentLoaded", function() {
	var btn = document.getElementById("user-invite-btn");
	if (btn !== null) {
		addInviteClickListener(btn);
	}
	return;
});

var emailRegEx = /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,4}$/i;

function addInviteClickListener(element) {
	element.addEventListener("click", function(e) {
		var email = document.getElementById("user-invite-email").value;

		if (email.search(emailRegEx) === -1) {
			// Bad email address
			alert("Please enter a valid email address");
			return;
		}

		var permsSelect = document.getElementById("user-invite-perms");
		var perms = permsSelect.options[permsSelect.selectedIndex].value;
		var cb = {
			success(resp) {
				var node = document.createElement("LI");
				setInviteHTML(node, email, resp);
				document.getElementById("user-invites-list").appendChild(node);
			},
			error(err) {
				console.error(err);
				alert("".concat("Error creating invite: ", err.responseText));
			},
		};
		client.createInvite(email, perms, cb);
	});
}

function setInviteHTML(node, email, pendingInvite) {
	node.className += " list-group-item";
	node.innerHTML = "".concat(
		"Share this link with <strong>", email, "</strong>: ",
		`<input class="input-large form-inline form-control" type="text" value="`, pendingInvite["Link"], `">`
	);
}
