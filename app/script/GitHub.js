exports.postIssueURL = function(uri, title, body) {
	return "https://" + uri + "/issues/new?title=" + encodeURIComponent(title) + "&body=" + encodeURIComponent(body);
};
