export function supportsAnnotatingFile(path) {
	if (!path) return false;

	const pathParts = path.split("/");
	let lang = pathParts[pathParts.length - 1].split(".")[1] || null;
	lang = lang ? lang.toLowerCase() : null;
	return lang === "go" || lang === "java" || lang === "sh" || lang === "bash";
}

export function parseURL(loc = window.location) {
	// TODO: this method has problems handling branch revisions with "/" character.
	let user, repo, repoURI, rev, path, isDelta;

	const urlsplit = loc.pathname.slice(1).split("/");
	user = urlsplit[0];
	repo = urlsplit[1]
	if (urlsplit[3] && (urlsplit[2] === "tree" || urlsplit[2] === "blob") || urlsplit[2] === "commit") {
		rev = urlsplit[3];
	}
	if (urlsplit[2] === "blob") {
		path = urlsplit.slice(4).join("/");
	}
	return {user, repo, repoURI: user && repo ? `github.com/${user}/${repo}` : null, rev, path, isDelta: urlsplit[2] === "pull" || urlsplit[2] === "commit"};
}

export function parseURLWithSourcegraphDef(loc = window.location) {
	let info = parseURL(loc);

	// We scrape the current branch and set rev to it so we stay on the same branch when doing jump-to-def.
	// Need to use the branch selector button because _clickRef passes a pathname as the location which,
	// only includes ${user}/${repo}, and no rev.
	let currBranch = getCurrentBranch();
	info.rev = currBranch;

	// Check for URL hashes like "#sourcegraph&def=...".
	if (loc.hash.startsWith("#sourcegraph&")) {
		loc.hash.slice(1).split("&").slice(1).forEach((p) => { // omit "sourcegraph" sentinel
			const kv = p.split("=", 2);
			if (kv.length != 2) return;
			let k = kv[0];
			const v = kv[1];
			if (k === "def") k = "defPath"; // disambiguate with def obj
			if (!info[k]) info[k] = v; // don't clobber
		});
	}
	return info;
}

export function isGitHubURL(loc = window.location) {
	return Boolean(loc.href.match(/https:\/\/(www.)?github.com/));
}

export function isSourcegraphURL(loc = window.location) {
	return Boolean(loc.href.match(/https:\/\/(www.)?sourcegraph.com/));
}

export function getCurrentBranch() {
	if (document.getElementsByClassName("select-menu-button js-menu-target css-truncate")[0]) {
		if (document.getElementsByClassName("select-menu-button js-menu-target css-truncate")[0].title !== "") {
			return document.getElementsByClassName("select-menu-button js-menu-target css-truncate")[0].title
		} else {
			return document.getElementsByClassName("js-select-button css-truncate-target")[0].innerText;
		}
	} else {
		return "master";
	}
}
