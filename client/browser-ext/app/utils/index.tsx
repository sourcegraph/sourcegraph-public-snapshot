export const supportedExtensions = [
	"go", // Golang
	"ts", "tsx", // TypeScript
	"js", "jsx", // JavaScript
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
};

export function getModeFromExtension(ext: string): string {
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

export function getGitHubRoute(loc: Location): string {
	return loc.pathname.split("/")[3];
}

export function getPathExtension(path: string): string {
	const pathSplit = path.split(".");
	if (pathSplit.length === 1) {
		return "";
	}
	if (pathSplit.length === 2 && pathSplit[0] === "") {
		return ""; // e.g. .gitignore
	}
	return pathSplit[pathSplit.length - 1].toLowerCase();
}

export interface GitHubURLData {
	user?: string;
	repo?: string;
	repoURI?: string;
	rev?: string;
	path?: string;
	isDelta?: boolean;
	isPullRequest?: boolean;
	isCommit?: boolean;
}
export function parseURL(loc: Location): GitHubURLData {
	// TODO(john): this method has problems handling branch revisions with "/" character.
	// TODO(john): this all needs unit testing!

	let user: string | undefined;
	let repo: string | undefined;
	let repoURI: string | undefined;
	let rev: string | undefined;
	let path: string | undefined;

	if (!isGitHubURL(loc)) {
		return {};
	}

	const urlsplit = loc.pathname.slice(1).split("/");
	user = urlsplit[0];
	repo = urlsplit[1];

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
	if (user && repo) {
		repoURI = `github.com/${user}/${repo}`;
	}

	const isPullRequest = urlsplit[2] === "pull";
	const isCommit = urlsplit[2] === "commit";
	const isDelta = isPullRequest || isCommit;

	return {user, repo, rev, path, repoURI, isDelta, isPullRequest, isCommit};
}

export function isGitHubURL(loc: Location): boolean {
	return Boolean(loc.href.match(/https:\/\/(www.)?github.com/));
}

export function isSourcegraphURL(loc: Location): boolean {
	return Boolean(loc.href.match(/https:\/\/(www.)?sourcegraph.com/));
}

export function getCurrentBranch(): string | null {
	let branchDropdownEl = document.getElementsByClassName("btn btn-sm select-menu-button js-menu-target css-truncate");
	if (!branchDropdownEl || branchDropdownEl.length !== 1) {
		return null;
	}

	return (branchDropdownEl[0] as any).title; // TODO(john): provide proper types
}

// TODO(john): improve the types here...urlData is actually object state.
export function convertBlobStateToEventLoggerStruct(urlData: GitHubURLData & {language: string, isPrivateRepo: boolean}): GitHubURLData & {language: string, isPrivateRepo: boolean} {
	return {
		repo: urlData.repo,
		repoURI: urlData.repoURI,
		rev: urlData.rev,
		path: urlData.path,
		isDelta: urlData.isDelta,
		isPullRequest: urlData.isPullRequest,
		isCommit: urlData.isCommit,
		language: urlData.language,
		isPrivateRepo: urlData.isPrivateRepo,
	};
}
