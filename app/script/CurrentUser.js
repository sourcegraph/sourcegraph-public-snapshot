if (document.head.dataset.currentUserLogin) {
	module.exports = {
		Login: document.head.dataset.currentUserLogin,
		Admin: document.head.dataset.currentUserAdmin === "true",
		Write: document.head.dataset.currentUserWrite === "true",
	};
} else {
	module.exports = null;
}
