// Sourcegraph API client
//
// ad-hoc for now

var $ = require("jquery");

function builds(repoURI, rev, noCache) {
	return $.ajax({
		url: `/.api/builds?Sort=updated_at&Direction=desc&PerPage=1&Repo=${repoURI || ""}&CommitID=${rev || ""}`,
		cache: !noCache,
	});
}

exports.builds = builds;

function createRepoBuild(repoURI, commitID, branch) {
	return $.ajax({
		url: `/.api/repos/${repoURI}/.builds`,
		method: "post",
		headers: {
			"X-Csrf-Token": window._csrfToken,
		},
		data: JSON.stringify({
			CommitID: commitID,
			Branch: branch,
			Config: {
				Import: true,
				Queue: true,
			},
		}),
	});
}

exports.createRepoBuild = createRepoBuild;

