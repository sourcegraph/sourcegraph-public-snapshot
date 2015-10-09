module.exports = function(url) {
	window.history.pushState(null, "", url);
	window.dispatchEvent(new Event("popstate"));
};
