export const supportedExtensions = [
	"go", // Golang
	"ts", "tsx" // TypeScript
];

export const upcomingExtensions = [
	"cs", // C#
	"css", // CSS
	"java", // Java
	"swift", // Swift
	"c", "h", // C
	"m", "mm", // Obj-C ("h" and "C" overlap with C/C++)
	"rb", "rbw", // Ruby
	"js", "jsx", // JavaScript
	"rs", "rlib", // Rust
	"sc", "scala", // Scala
	"htm", "html", // HTML
	"pl", "pm", "t", "pod", // Perl
	"clj", "cljs", "cljc", "edn", // Clojure
	"py", "pyc", "pyd", "pyo", "pyw", "pyz", // Python
	"cc", "cpp", "cxx", "c++", "hh", "hpp", "hxx", "h++", // C++ ("h" and "c" overlap with C)
	"php", "phtml", "php3", "php4", "php5", "php7", "phps", // PHP
];

export function getModeFromExtension(ext) {
	switch (ext) {
		case "go":
			return "go";
		case "ts":
		case "tsx":
			return "typescript";
		default:
			return "unknown";
	}
}

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
