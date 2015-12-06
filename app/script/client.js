// Sourcegraph API client
//
// ad-hoc for now

var apirouter = require("./routing/apirouter");
var router = require("./routing/router");
var $ = require("jquery");

function userOrgs(uid) {
	return $.ajax({
		url: `/.api/users/$${uid}/orgs`,
	});
}

exports.userOrgs = userOrgs;

function repos(opts) {
	return $.ajax({
		url: "/.api/repos",
		data: opts,
	});
}

exports.repos = repos;

function repoFiles(repo, rev) {
	return $.ajax({
		url: `/.ui/${router.fileListURL(repo, rev)}`,
	});
}

exports.repoFiles = repoFiles;

function searchSuggestions(rawQuery) {
	return $.ajax({
		url: "/.api/search/suggestions",
		data: rawQuery,
	});
}

exports.searchSuggestions = searchSuggestions;

function builds(repoURI, rev, noCache) {
	return $.ajax({
		url: `/.api/builds?Sort=updated_at&Direction=desc&PerPage=1&Repo=${repoURI || ""}&CommitID=${rev || ""}`,
		cache: !noCache,
	});
}

exports.builds = builds;

function createInvite(email, perms) {
	return $.ajax({
		url: `/.ui/.invite?Email=${encodeURIComponent(email)}&Permission=${perms}`,
	});
}

exports.createInvite = createInvite;

function createRepoBuild(repoURI, rev) {
	return $.ajax({
		url: `/.api/repos/${repoURI}@${rev}/.builds`,
		method: "post",
		data: JSON.stringify({
			Import: true,
			Queue: true,
		}),
	});
}

exports.createRepoBuild = createRepoBuild;

function listExamples(defKey, query) {
	query = query ? query : "";
	var d = defKey;
	return $.ajax({
		url: `${apirouter.defExamplesURL(d.Repo, d.CommitID, d.UnitType, d.Unit, d.Path)}?${query}`,
		type: "GET",
		dataType: "json",
	});
}
exports.listExamples = listExamples;

function createDeltaRoute(routeVars) {
	return `/.api/repos/${routeVars["Repo"]}/.deltas/${routeVars["Rev"]}..${routeVars["DeltaHeadRev"]}`;
}

function deltaListUnits(routeVars) {
	return $.ajax({
		url: `${createDeltaRoute(routeVars)}/.units`,
		method: "get",
	});
}
exports.deltaListUnits = deltaListUnits;

function deltaListDefs(routeVars, opt) {
	return $.ajax({
		url: `${createDeltaRoute(routeVars)}/.defs`,
		data: opt,
		method: "get",
	});
}
exports.deltaListDefs = deltaListDefs;

function listFiles(routeVars, opt, cb) {
	var optPieces = [];
	if (opt.Filter) {
		optPieces.push(`Filter=${encodeURIComponent(opt.Filter)}`);
	}
	var url = `${createDeltaRoute(routeVars)}/.files`;
	if (optPieces.length > 0) {
		url = `${url}?${optPieces.join("&")}`;
	}
	return $.ajax({
		url: url,
		method: "get",
		success: cb.success,
		error: cb.error,
	});
}
exports.listFiles = listFiles;

function listAffectedDependents(routeVars, opt) {
	return $.ajax({
		url: `${createDeltaRoute(routeVars)}/.affected-dependents${opt.NotFormatted ? "?NotFormatted=true" : ""}`,
		method: "get",
	});
}
exports.listAffectedDependents = listAffectedDependents;

function listReviewers(routeVars) {
	return $.ajax({
		url: `${createDeltaRoute(routeVars)}/.reviewers`,
		method: "get",
	});
}
exports.listReviewers = listReviewers;

function listAffectedAuthors(routeVars, opt) {
	return $.ajax({
		url: `${createDeltaRoute(routeVars)}/.affected-authors`,
		data: opt,
		method: "get",
	});
}
exports.listAffectedAuthors = listAffectedAuthors;

function listAffectedClients(routeVars, opt) {
	return $.ajax({
		url: `${createDeltaRoute(routeVars)}/.affected-clients`,
		data: opt,
		method: "get",
	});
}
exports.listAffectedClients = listAffectedClients;

function renderMarkdown(markdown, checkboxes, cb) {
	return $.ajax({
		url: "/.api/markdown",
		method: "post",
		data: JSON.stringify({
			Markdown: markdown,
			RenderCheckboxes: checkboxes,
		}),
		success: cb.success,
		error: cb.error,
	});
}
exports.renderMarkdown = renderMarkdown;
