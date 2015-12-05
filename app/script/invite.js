var client = require("./client");

document.addEventListener("DOMContentLoaded", function() {
	var btn = document.getElementById("user-invite-btn");
	if (btn !== null) {
		addInviteClickListener(btn);
	}
	return;
});

function addInviteClickListener(element) {
	element.addEventListener("click", function(e) {
		var email = document.getElementById("user-invite-email").value;
		var permsSelect = document.getElementById("user-invite-perms");
		var perms = permsSelect.options[permsSelect.selectedIndex].value;
		client.createInvite(email, perms).success(function(resp) {
			var node = document.createElement("LI");
			setInviteHTML(node, email, resp);
			document.getElementById("user-invites-list").appendChild(node);
		}).error(function(err) {
			console.error(err);
		});
	});
}

function setInviteHTML(node, email, pendingInvite) {
	node.className += " list-group-item";
	node.innerHTML = "".concat(
		"Share this link with <strong>", email, "</strong>: ",
		`<input class="input-large form-inline form-control" type="text" value="`, pendingInvite["Link"], `">`
	);
}
