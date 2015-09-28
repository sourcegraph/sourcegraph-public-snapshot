module.exports = {
	repoURL(repo, rev) {
		return "/api/repos/" + repo + (rev ? "@" + rev : "");
	},

	// defURL constructs the URL to the def API endpoint. The def spec
	// may be passed as individual arguments as listed in the function
	// prototype, or as one DefSpec object argument.
	defURL(repo, rev, unitType, unit, path) {
		if (typeof repo === "object") {
			var defSpec = repo;
			repo = defSpec.Repo;
			rev = defSpec.CommitID;
			unitType = defSpec.UnitType;
			unit = defSpec.Unit;
			path = defSpec.Path;
		}
		path = path ? "/" + path : ""; // account for package paths
		return module.exports.repoURL(repo, rev) + "/.defs/." + unitType + "/" + unit + "/.def" + path;
	},

	// defExamplesURL constructs the URL to the def examples API
	// endpoint. As with defURL, the def spec may be passed as
	// individual arguments as listed in the function prototype, or as
	// one DefSpec object argument.
	defExamplesURL(repo, rev, unitType, unit, path) {
		return module.exports.defURL(repo, rev, unitType, unit, path) + "/.examples";
	},
};
