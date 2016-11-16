import * as utils from ".";

function invariant(cond: any): void {
	if (!cond) {
		throw new Error("invariant exception");
	}
}

// getFileContainers returns the elements on the page which should be marked
// up with tooltips & links:
// - blob view: a single file
// - commit view: one or more file diffs
// - PR conversation view: snippets with inline comments
// - PR unified/split view: one or more file diffs
export function getFileContainers(): HTMLCollectionOf<HTMLElement> {
	return document.getElementsByClassName("file") as HTMLCollectionOf<HTMLElement>;
}

export function isInlineCommentContainer(file: HTMLElement): boolean {
	return file.classList.contains("inline-review-comment");
}

export function isPrivateRepo(): boolean {
	return document.getElementsByClassName("label label-private v-align-middle").length > 0;
}

export function registerExpandDiffClickHandler(cb: (ev: any) => void): void {
	const diffExpanders = document.getElementsByClassName("diff-expander");
	for (let i = 0; i < diffExpanders.length; ++i) {
		const expander = diffExpanders[i];
		if (expander.className.indexOf("sg-diff-expander") !== -1) {
			// Don't register more than one handler.
			continue;
		}
		expander.className = `${expander.className} sg-diff-expander`;
		expander.addEventListener("click", cb);
	}
}

// getDeltaFileName returns the path of the file container
export function getDeltaFileName(container: HTMLElement): string {
	const info = container.querySelector(".file-info") as HTMLElement;
	invariant(info);

	if (info.title) {
		// for PR conversation snippets
		return info.title;
	} else {
		const link = info.querySelector("a") as HTMLElement;
		invariant(link);
		invariant(link.title);
		return link.title;
	}
}

export function isSplitDiff(): boolean {
	const {isDelta, isPullRequest} = utils.parseURL(window.location);
	if (!isDelta) {
		return false;
	}

	if (isPullRequest) {
		const headerBar = document.getElementsByClassName("float-right pr-review-tools");
		if (!headerBar || headerBar.length !== 1) {
			return false;
		}

		const diffToggles = headerBar[0].getElementsByClassName("BtnGroup");
		invariant(diffToggles && diffToggles.length === 1);

		const disabledToggle = diffToggles[0].getElementsByTagName("A")[0] as HTMLAnchorElement;
		return disabledToggle && !disabledToggle.href.includes("diff=split");
	} else { // delta for a commit view
		const headerBar = document.getElementsByClassName("details-collapse table-of-contents js-details-container");
		if (!headerBar || headerBar.length !== 1) {
			return false;
		}

		const diffToggles = headerBar[0].getElementsByClassName("BtnGroup float-right");
		invariant(diffToggles && diffToggles.length === 1);

		const selectedToggle = diffToggles[0].querySelector(".selected") as HTMLAnchorElement;
		return selectedToggle && selectedToggle.href.includes("diff=split");
	}
}

export interface DeltaRevs {
	base: string;
	head: string;
}

export function getDeltaRevs(): DeltaRevs | null {
	const {isDelta, isCommit} = utils.parseURL(window.location);
	if (!isDelta) {
		return null;
	}

	let base = "";
	let head = "";
	// const fetchContainer = document.getElementsByClassName("js-socket-channel js-updatable-content js-pull-refresh-on-pjax");
	let fetchContainers = document.getElementsByClassName("js-socket-channel js-updatable-content js-pull-refresh-on-pjax");
	if (fetchContainers && fetchContainers.length === 1) {
		for (let i = 0; i < fetchContainers.length; ++i) {
		// for conversation view of pull request
		const el = fetchContainers[i] as HTMLElement;
		const url = el.dataset ? el.dataset["url"] : null;
		if (!url) {
			continue;
		}

		const urlSplit = url.split("?");
		invariant(urlSplit.length === 2);
		const query = urlSplit[1];
		const querySplit = query.split("&");
		for (let kv of querySplit) {
			const kvSplit = kv.split("=");
			const k = kvSplit[0];
			const v = kvSplit[1];
			if (k === "base_commit_oid") {
				base = v;
			}
			if (k === "end_commit_oid") {
				head = v;
			}
		}
		}
	} else if (isCommit) {
		const shaContainer = document.querySelectorAll(".sha-block");
		if (shaContainer && shaContainer.length === 2) {
			const baseShaEl = shaContainer[0].querySelector("a");
			if (baseShaEl) {
				// e.g "https://github.com/gorilla/mux/commit/0b13a922203ebdbfd236c818efcd5ed46097d690"
				base = baseShaEl.href.split("/").slice(-1)[0];
			}
			const headShaEl = shaContainer[1].querySelector("span.sha") as HTMLElement;
			if (headShaEl) {
				head = headShaEl.innerHTML;
			}
		}
	}

	if (base === "" || head === "") {
		return null;
	}
	return {base, head};
}

export interface DeltaInfo {
	baseBranch: string;
	baseURI: string;
	headBranch: string;
	headURI: string;
}
export function getDeltaInfo(): DeltaInfo | null {
	const {repo, repoURI, isDelta, isPullRequest, isCommit} = utils.parseURL(window.location);
	if (!isDelta) {
		return null;
	}

	invariant(repoURI);

	let baseBranch = "";
	let headBranch = "";
	let baseURI = "";
	let headURI = "";
	if (isPullRequest) {
		const branches = document.querySelectorAll(".commit-ref,.current-branch") as HTMLCollectionOf<HTMLElement>;
		baseBranch = branches[0].innerText;
		headBranch = branches[1].innerText;

		if (baseBranch.includes(":")) {
			const baseSplit = baseBranch.split(":");
			baseBranch = baseSplit[1];
			baseURI = `github.com/${baseSplit[0]}/${repo}`;
		} else {
			baseBranch = repoURI as string;
		}
		if (headBranch.includes(":")) {
			const headSplit = headBranch.split(":");
			headBranch = headSplit[1];
			headURI = `github.com/${headSplit[0]}/${repo}`;
		} else {
			headURI = repoURI as string;
		}

	} else if (isCommit) {
		let branchEl = document.querySelector("li.branch") as HTMLElement;
		if (branchEl) {
			branchEl = branchEl.querySelector("a") as HTMLElement;
		}
		if (branchEl) {
			baseBranch = branchEl.innerText;
			headBranch = branchEl.innerText;
		}
		baseURI = repoURI as string;
		headURI = repoURI as string;
	}

	if (baseBranch === "" || headBranch === "" || baseURI === "" || headURI === "") {
		return null;
	}
	return {baseBranch, headBranch, baseURI, headURI};
}
