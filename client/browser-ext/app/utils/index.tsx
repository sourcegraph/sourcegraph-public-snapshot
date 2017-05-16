import { isBrowserExtension } from "./context";
import { Domain, GitHubURLData } from "./types";

/**
 * supportedExtensions are the file extensions
 * the extension will apply annotations to
 */
export const supportedExtensions = new Set<string>([
	"go", // Golang
	"ts", "tsx", // TypeScript
	"js", "jsx", // JavaScript
	"java", // Java
	"py", "pyc", "pyd", "pyo", "pyw", "pyz", // Python
]);

/**
 * upcomingExtensions are the file types which
 * will be supported "soon"
 */
export const upcomingExtensions = new Set<string>([
	// "cs", // C#
	// "css", // CSS
	// "swift", // Swift
	// "c", "h", // C
	// "m", "mm", // Obj-C ("h" and "C" overlap with C/C++)
	// "rb", "rbw", // Ruby
	// "rs", "rlib", // Rust
	// "sc", "scala", // Scala
	// "htm", "html", // HTML
	// "pl", "pm", "t", "pod", // Perl
	// "clj", "cljs", "cljc", "edn", // Clojure
	// "cc", "cpp", "cxx", "c++", "hh", "hpp", "hxx", "h++", // C++ ("h" and "c" overlap with C)
	"php", "phtml", "php3", "php4", "php5", "php7", "phps", // PHP
]);

/**
 * getModeFromExtension returns the LSP mode for the
 * provided file extension (e.g. "jsx")
 */
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
		case "java":
			return "java";
		case "py":
		case "pyc":
		case "pyd":
		case "pyo":
		case "pyw":
		case "pyz":
			return "python";
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

export function parseURL(loc: Location): GitHubURLData {
	// TODO(john): this method has problems handling branch revisions with "/" character.
	// TODO(john): this all needs unit testing!

	let user: string | undefined;
	let repo: string | undefined;
	let repoURI: string | undefined;
	let rev: string | undefined;
	let path: string | undefined;

	const domain = getDomain(loc);
	if (domain !== Domain.GITHUB) {
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

	return { user, repo, rev, path, repoURI, isDelta, isPullRequest, isCommit };
}

export function getCurrentBranch(): string | null {
	let branchDropdownEl = document.getElementsByClassName("btn btn-sm select-menu-button js-menu-target css-truncate");
	if (branchDropdownEl.length !== 1) {
		return null;
	}

	return (branchDropdownEl[0] as HTMLElement).title;
}

export function getPlatformName(): string {
	if (!isBrowserExtension()) {
		return "phabricator-integration";
	}
	return window.navigator.userAgent.indexOf("Firefox") !== -1 ? "firefox-extension" : "chrome-extension";
}

export function isE2ETest(): boolean {
	return process.env.NODE_ENV === "test";
}

export function getDomain(loc: Location): Domain {
	if (/^https?:\/\/phabricator.aws.sgdev.org/.test(loc.href)) {
		return Domain.SGDEV_PHABRICATOR;
	}
	if (/^https?:\/\/(www.)?github.com/.test(loc.href)) {
		return Domain.GITHUB;
	}
	if (/^https?:\/\/(www.)?sourcegraph.com/.test(loc.href)) {
		return Domain.SOURCEGRAPH;
	}
	if (/^https?:\/\/(www.)?localhost:7990/.test(loc.href)) {
		return Domain.SGDEV_BITBUCKET;
	}
	throw new Error(`Unable to determine the domain, ${loc.href}`);
}

/**
 * This method created a unique username based on the platform and domain the user is visiting.
 * Examples: sg_dev_phabricator:matt , or uber_phabricator:matt
 */
export function getDomainUsername(domain: string, username: string): string {
	return `${domain}:${username}`;
}

export function getSourcegraphBlobUrl(sourcegraphUrl: string, repoUri: string, path: string, commitId?: string): string {
	const commitString = commitId ? `@${commitId}` : "";
	return `${sourcegraphUrl}/${repoUri}${commitString}/-/blob/${path}?utm_source=${getPlatformName()}`;
}
