// Sourcegraph API client
//
// ad-hoc for now

var router = require("./routing/router");
var $ = require("jquery");

function repoFiles(repo, rev) {
	return $.ajax({
		url: `/.ui/${router.fileListURL(repo, rev)}`,
	});
}

exports.repoFiles = repoFiles;

function builds(repoURI, rev, noCache) {
	return $.ajax({
		url: `/.api/builds?Sort=updated_at&Direction=desc&PerPage=1&Repo=${repoURI || ""}&CommitID=${rev || ""}`,
		cache: !noCache,
	});
}

exports.builds = builds;

function createInvite(email, perms, cb) {
	return $.ajax({
		url: "/.ui/.invite",
		method: "post",
		headers: {
			"X-Csrf-Token": window._csrfToken,
		},
		data: JSON.stringify({
			Email: email,
			Permission: perms,
		}),
		success: cb.success,
		error: cb.error,
	});
}

exports.createInvite = createInvite;

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

