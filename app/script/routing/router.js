module.exports = {
	fileURL(repo, rev, path) {
		path = (path ? path : "");
		path = path.replace(/^\//, "");
		return `/${repo}@${rev}/.tree/${path}`;
	},
	commitsURL(repo, rev) {
		return `/${repo}@${rev}/.commits`;
	},
};
