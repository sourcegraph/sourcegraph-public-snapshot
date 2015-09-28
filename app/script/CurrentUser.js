if (document.head.dataset.currentUserLogin) {
	module.exports = {
		Login: document.head.dataset.currentUserLogin,
	};
} else {
	module.exports = null;
}
