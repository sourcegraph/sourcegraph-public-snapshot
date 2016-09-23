export const supportedExtensions = [
	"go" //, "java", "js", "jsx", "ts", "tsx"
];

export function getPathExtension(path) {
	const pathSplit = path.split(".");
	if (pathSplit.length === 1) return null;
	if (pathSplit.length === 2 && pathSplit[0] === "") return null; // e.g. .gitignore
	return pathSplit[pathSplit.length - 1].toLowerCase();
}

export function parseURL(loc = window.location) {
	// TODO: this method has problems handling branch revisions with "/" character.
	let user, repo, repoURI, rev, path, isDelta;

	const urlsplit = loc.pathname.slice(1).split("/");
	user = urlsplit[0];
	repo = urlsplit[1]

	let revParts = 1; // a revision may have "/" chars, in which case we consume multiple parts;
	if (urlsplit[3] && (urlsplit[2] === "tree" || urlsplit[2] === "blob") || urlsplit[2] === "commit") {
		const currBranch = getCurrentBranch();
		if (currBranch) {
			revParts = currBranch.split("/").length;
		}
		rev = urlsplit.slice(3, 3 + revParts).join("/");
	}
	if (urlsplit[2] === "blob") {
		path = urlsplit.slice(3 + revParts).join("/");
	}
	return {user, repo, repoURI: user && repo ? `github.com/${user}/${repo}` : null, rev, path, isDelta: urlsplit[2] === "pull" || urlsplit[2] === "commit", isPullRequest: urlsplit[2] === "pull", isCommit: urlsplit[2] === "commit"};
}

export function isGitHubURL(loc = window.location) {
	return Boolean(loc.href.match(/https:\/\/(www.)?github.com/));
}

export function isSourcegraphURL(loc = window.location) {
	return Boolean(loc.href.match(/https:\/\/(www.)?sourcegraph.com/));
}

export function getCurrentBranch() {
	let branchDropdownEl = document.getElementsByClassName("btn btn-sm select-menu-button js-menu-target css-truncate");
	if (!branchDropdownEl || branchDropdownEl.length !== 1) return null;

	return branchDropdownEl[0].title;
}
