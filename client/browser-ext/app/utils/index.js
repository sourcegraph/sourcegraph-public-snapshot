export const supportedExtensions = [
	"go", // Golang
	"ts", "tsx", // TypeScript
	"js", "jsx" // JavaScript
];

export const upcomingExtensions = [
	"cs", // C#
	"css", // CSS
	"java", // Java
	"swift", // Swift
	"c", "h", // C
	"m", "mm", // Obj-C ("h" and "C" overlap with C/C++)
	"rb", "rbw", // Ruby
	"rs", "rlib", // Rust
	"sc", "scala", // Scala
	"htm", "html", // HTML
	"pl", "pm", "t", "pod", // Perl
	"clj", "cljs", "cljc", "edn", // Clojure
	"py", "pyc", "pyd", "pyo", "pyw", "pyz", // Python
	"cc", "cpp", "cxx", "c++", "hh", "hpp", "hxx", "h++", // C++ ("h" and "c" overlap with C)
	"php", "phtml", "php3", "php4", "php5", "php7", "phps", // PHP
];

export const readableGitHubRoute = {
	"blob": "File",
	"pull": "Pull request",
	"commit": "Commit",
}

export function getModeFromExtension(ext) {
	switch (ext) {
		case "go":
			return "go";
		case "ts":
		case "tsx":
			return "typescript";
		case "js":
		case "jsx":
			return "javascript";
		default:
			return "unknown";
	}
}

export function getGitHubRoute(loc = window.location) {
	return loc.pathname.split("/")[3];
}

export function getLinesOfCode() {
	let nTotalLines = 0;
	const view = getGitHubRoute();

	switch (view)  {
		case "blob":
			nTotalLines = parseInt(document.getElementsByClassName("file-info")[0].textContent.match(/([0-9]+) line/), 10);
			break;
		case "pull":
			if (window.location.pathname.endsWith("files")) {
				nTotalLines = parseInt(document.getElementsByClassName("diffbar-item diffstat")[0].querySelector(".tooltipped").getAttribute("aria-label").replace(/[^\d]/g, ''), 10);
			}
			break;
		case "commit":
			[].map.call(document.getElementsByClassName("toc-diff-stats")[0].getElementsByTagName("strong"), (node) => {
				nTotalLines += parseInt(node.textContent.replace(/[^\d]/g, ''), 10);
			});
			break;
	}

	return nTotalLines;
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

	if (!isGitHubURL()) return {user: null, repo: null, repoURI: null, rev: null, path: null, isDelta: null, isPullRequest: null, isCommit: null};

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

export function convertBlobStateToEventLoggerStruct(state) {
	return {
		repo: state.repo,
		repoURI: state.repoURI,
		rev: state.rev,
		path: state.path,
		isDelta: state.isDelta,
		isPullRequest: state.isPullRequest,
		isCommit: state.isCommit,
		language: state.language,
		isPrivateRepo: state.isPrivateRepo,
	};
}
