module.exports = {

	repoURL(repo, rev) {
		return `/${repo}${rev ? `@${rev}` : ""}`;
	},

	changesetURL(repo, id) {
		return `${module.exports.repoURL(repo)}/.changes/${id}`;
	},

	discussionURL(defKey, id) {
		var repo;
		if (typeof defKey === "string") {
			var parts = module.exports.deconstructDefURL(defKey);
			repo = parts.repo;
		}
		if (typeof defKey === "object") {
			repo = defKey.Repo;
		}
		return `${module.exports.repoURL(repo)}/.discussion/${id}`;
	},

	discussionListURL(defKey, order) {
		var params = order ? `?order=${order}` : "";
		return module.exports._discussionsURL(defKey) + params;
	},

	discussionCreateURL(defKey) {
		return `${module.exports._discussionsURL(defKey)}/create`;
	},

	discussionCreateCommentURL(defKey, id) {
		return `${module.exports._discussionsURL(defKey)}/${id}/.comment`;
	},

	_discussionsURL(defKey) {
		return `${defKey}/.discussions`;
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
		return `${module.exports.repoURL(repo, rev)}/.tree/${path}`;
	},

	fileRangeURL(repo, rev, path, startline, endline) {
		path = (path ? path : "");
		return `${module.exports.repoURL(repo, rev)}/.tree/${path}#startline=${startline}&endline=${endline}`;
	},

	fileListURL(repo, rev) {
		return `${module.exports.repoURL(repo, rev)}/.filefinder`;
	},

	commitsURL(repo, rev) {
		return `${module.exports.repoURL(repo, rev)}/.commits`;
	},

	/**
	 * @description Constructs a compare view URL.
	 * @param {string} repo - Repository URL
	 * @param {string} base - Base compare revision
	 * @param {string} head - Head compare revision
	 * @param {string=} filter - Optional filter
	 * @returns {string} The resulting URL
	 */
	compareURL(repo, base, head, filter) {
		return `${module.exports.repoURL(repo, base)}/.compare/${head}${filter ? `?filter=${filter}` : ""}`;
	},

	signInURL(returnTo) {
		return `/login${returnTo ? `?return-to=${returnTo}` : ""}`;
	},

	personURL(login) {
		return `/${login}`;
	},

	logInURL() { return "/login"; },
	logOutURL() { return "/logout"; },

	urlToUserSubroute(route, login) {
		var subroutePaths = {
			"person.settings.profile": "/.settings/profile",
			"person.settings.integrations": "/.settings/integrations",
			"person.settings.auth": "/.settings/auth",
		};
		if (!subroutePaths[route]) throw new Error(`No such route: ${route}`);
		return module.exports.personURL(login) + subroutePaths[route];
	},

	abs(url) {
		if (/^https?:\/\//.test(url)) return url;
		return `${window.location.protocol}//${window.location.host}${url}`;
	},

	// appdashUploadPageLoadURL constructs a URL string to which a POST
	// request can be made, given start and end unix timestamps in milliseconds
	// representing the start and end of page content loading.
	appdashUploadPageLoadURL(start, end) {
		return `/.ui/.appdash/upload-page-load?S=${start}&E=${end}`;
	},
};
