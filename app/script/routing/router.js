module.exports = {

	repoURL(repo, rev) {
		return `/${repo}${rev ? `@${rev}` : ""}`;
	},

	// defURL constructs the application (not API) URL to a def. The def
	// spec may be passed as individual arguments as listed in the
	// function prototype, or as one DefSpec object argument.
	defURL(repo, rev, unitType, unit, path) {
		if (typeof repo === "object") {
			var defSpec = repo;
			repo = defSpec.Repo;
			rev = defSpec.CommitID;
			unitType = defSpec.UnitType;
			unit = defSpec.Unit;
			path = defSpec.Path;
		}
		return `${module.exports.repoURL(repo, rev)}/.${unitType}/${unit}/.def${path !== "." ? `/${path}` : ""}`;
	},

	// Returns an object having keys repo, rev, unitType, unit and path extracted from
	// a (valid) URL.
	//
	// Assumes that format is {repo}[@rev]/.{unitType}[unit]/.def/{path} where square brackets
	// are optionallly present. No "unit" means "." and will be returned as undefined.
	deconstructDefURL(url) {
		var parts = url.split("/."),
			repoAndRev = parts[0],
			unitTypeAndUnit = parts[1],
			defPath = parts[2],
			result = {},
			bits;

		bits = repoAndRev.split("@");
		result.repo = bits[0][0] === "/" ? bits[0].slice(1) : bits[0];
		result.rev = bits[1];

		bits = unitTypeAndUnit.split("/");
		result.unitType = bits[0];
		result.unit = bits.length > 1 ? unitTypeAndUnit.slice(unitTypeAndUnit.indexOf("/")+1) : undefined;
		result.path = defPath.split("def/")[1];

		return result;
	},

	defExamplesURL(repo, rev, unitType, unit, path) {
		return `${module.exports.defURL(repo, rev, unitType, unit, path)}/.examples`;
	},

	fileURL(repo, rev, path) {
		path = (path ? path : "");
		path = path.replace(/^\//, "");
		return `${module.exports.repoURL(repo, rev)}/.tree/${path}`;
	},

	fileRangeURL(repo, rev, path, startline, endline) {
		path = (path ? path : "");
		return `${module.exports.repoURL(repo, rev)}/.tree/${path}#L${startline}-${endline}`;
	},

	commitsURL(repo, rev) {
		return `${module.exports.repoURL(repo, rev)}/.commits`;
	},

	signInURL(returnTo) {
		return `/login${returnTo ? `?return-to=${returnTo}` : ""}`;
	},

	logInURL() { return "/login"; },
	logOutURL() { return "/logout"; },

	urlToUserSubroute(route, login) {
		var subroutePaths = {
			"person.settings.profile": "/.settings/profile",
			"person.settings.integrations": "/.settings/integrations",
		};
		if (!subroutePaths[route]) throw new Error(`No such route: ${route}`);
		return `/~${login}${subroutePaths[route]}`;
	},

	abs(url) {
		if (/^https?:\/\//.test(url)) return url;
		return `${window.location.protocol}//${window.location.host}${url}`;
	},
};
