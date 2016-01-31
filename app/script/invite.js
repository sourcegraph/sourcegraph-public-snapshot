var client = require("./client");

document.addEventListener("DOMContentLoaded", function() {
	var btn = document.getElementById("user-invite-btn");
	if (btn !== null) {
		addInviteClickListener(btn);
	}
	var btn2 = document.getElementById("teammate-invite-btn");
	if (btn2 !== null) {
		addTeammateInviteClickListener(btn2);
	}
	return;
});

function addInviteClickListener(element) {
	element.addEventListener("click", function(e) {
		if (!document.getElementById("user-invite-email").checkValidity()) return;
		var email = document.getElementById("user-invite-email").value;

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

function addTeammateInviteClickListener(element) {
	element.addEventListener("click", function(e) {
		if (!document.getElementById("teammate-invite-email").checkValidity()) return;
		var email = document.getElementById("teammate-invite-email").value;

		var cb = {
			success(resp) {
				var node = document.createElement("LI");
				setInviteHTML(node, email, resp);
				document.getElementById("teammate-invites-list").appendChild(node);
			},
			error(err) {
				console.error(err);
				alert("".concat("Error creating invite: ", err.responseText));
			},
		};
		client.createInvite(email, "write", cb);
	});
}

function setInviteHTML(node, email, pendingInvite) {
	node.className += " list-group-item";
	if (pendingInvite["EmailSent"]) {
		node.innerHTML = "".concat(
			"Email sent to <strong>", email, "</strong>: ",
			`<input class="input-large form-inline form-control" type="text" value="`, pendingInvite["Link"], `">`
		);
	} else {
		node.innerHTML = "".concat(
			"Share this link with <strong>", email, "</strong>: ",
			`<input class="input-large form-inline form-control" type="text" value="`, pendingInvite["Link"], `">`
		);
	}
}
