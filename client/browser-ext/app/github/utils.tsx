import { getDomain } from "../utils";
import { Domain, GitHubURL } from "../utils/types";

export function parseURL(loc: Location = window.location): GitHubURL {
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

	return { user, repo, rev, path, repoURI, uri: repoURI, isDelta, isPullRequest, isCommit };
}

export function getCurrentBranch(): string | null {
	const branchDropdownEl = document.getElementsByClassName("btn btn-sm select-menu-button js-menu-target css-truncate");
	if (branchDropdownEl.length !== 1) {
		return null;
	}

	return (branchDropdownEl[0] as HTMLElement).title;
}
