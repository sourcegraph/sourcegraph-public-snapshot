import { PhabricatorInstance } from "../../app/utils/classes";
import { CodeCell, PhabChangeUrl, PhabDifferentialUrl, PhabDiffusionUrl, PhabRevisionUrl, PhabricatorCodeCell, PhabricatorMode, PhabUrl } from "../../app/utils/types";
import { phabricatorInstance } from "./context";

const REV_SHA_PATTERN = /r([0-9A-z]+)([0-9a-f]{40})/;

function getRevFromPage(): string | null {
	const revElement = document.getElementsByClassName("phui-tag-core").item(0);
	if (!revElement) {
		return null;
	}
	const shaMatch = REV_SHA_PATTERN.exec(revElement.children[0].getAttribute("href") as string);
	if (!shaMatch) {
		return null;
	}
	return shaMatch[2];
}

function isDifferentialLanded(): boolean {
	const closedElement = document.getElementsByClassName("visual-only phui-icon-view phui-font-fa fa-check-square-o");
	if (closedElement.length === 0) {
		return false;
	}
	return true;
}

/**
 * DEPRECATED: there can be more than one commit here, to different branches
 * 	prefer to get the commit ID to master from the revision contents table,
 *  in the description of the last diff.
 */
function getDifferentialCommitFromPage(): string | null {
	const possibleRevElements = document.getElementsByClassName("phui-property-list-value");
	for (const revElement of Array.from(possibleRevElements)) {
		if (!(revElement && revElement.children && revElement.children[0])) {
			continue;
		}
		const linkHref =  revElement.children[0].getAttribute("href");;
		if (!linkHref) {
			continue;
		}
		const shaMatch = REV_SHA_PATTERN.exec(linkHref);
		if (!shaMatch) {
			continue;
		}
		return shaMatch[2];
	}
	return null;
}

const DIFF_LINK = /D[0-9]+\?id=([0-9]+)/i;
function getMaxDiffFromTabView(): {diffId: number, revDescription: string} | null {
	// first, find Revision contents table box
	const headerShells = document.getElementsByClassName("phui-header-header");
	let revisionContents: Element | null = null;
	for (const headerShell of Array.from(headerShells)) {
		if (headerShell.textContent === "Revision Contents") {
			revisionContents = headerShell;
		}
	}
	if (!revisionContents) {
		return null;
	}
	const parentContainer = revisionContents.parentElement!.parentElement!.parentElement!.parentElement!.parentElement!;
	const tables = parentContainer.getElementsByClassName("aphront-table-view");
	for (const table of Array.from(tables)) {
		const tableRows = (table as HTMLTableElement).rows;
		const row = tableRows[0];
		// looking for the history tab of the revision contents table
		if (row.children[0].textContent !== "Diff") {
			continue;
		}
		const links = table.getElementsByTagName("a");
		let max: {diffId: number, revDescription: string} | null = null;
		for (const link of Array.from(links)) {
			const linkHref = link.getAttribute("href");
			if (!linkHref) {
				continue;
			}
			const matches = DIFF_LINK.exec(linkHref);
			if (!matches) {
				continue;
			}
			let revDescription = link.parentNode!.parentNode!.childNodes[3].textContent;
			const shaMatch = REV_SHA_PATTERN.exec(revDescription!);
			if (!shaMatch) {
				continue;
			}
			max = max && max.diffId > parseInt(matches[1], 10) ? max : { diffId: parseInt(matches[1], 10), revDescription: shaMatch[2]};
		}
		return max;
	}
	return null;
}

function getRepoFromDifferentialPage(): string | null {
	const mainColumn = document.getElementsByClassName("phui-main-column").item(0);
	if (!mainColumn) {
		console.warn("no 'phui-main-column'[0] class found on differential page.");
		return null;
	}
	const diffDetailBox = mainColumn.children[1];
	const repositoryTag = diffDetailBox.getElementsByClassName("phui-property-list-value").item(0);
	if (!repositoryTag) {
		console.warn("no 'phui-property-list-value'[0] class found on differential page.");
		return null;
	}
	// format is /source/phabricator/
	let repoUrl = repositoryTag.children[0].getAttribute("href");
	if (!repoUrl) {
		console.warn("no href found on repository link on differential page.");
		return null;
	}
	if (repoUrl.startsWith("/source/")) {
		repoUrl = repoUrl.substr("/source/".length);
	} else if (repoUrl.startsWith("/diffusion/")) {
		// this second one exists @ umami
		repoUrl = repoUrl.substr("/diffusion/".length);
	} else {
		console.error(`Unrecongized prefix on repo url ${repoUrl}`);
		return null;
	}
	repoUrl = repoUrl.substr(0, repoUrl.length - 1);
	return repoUrl;
}

const DIFF_PATTERN = /Diff ([0-9]+)/;
function getDiffIdFromDifferentialPage(): string | null {
	const diffsContainer = document.getElementById("differential-review-stage");
	if (!diffsContainer) {
		console.error(`no element with id differential-review-stage found on page.`);
		return null;
	}
	const wrappingDiffBox = diffsContainer.parentElement;
	if (!wrappingDiffBox) {
		console.error(`parent container of diff container not found.`);
		return null;
	}
	const diffTitle = wrappingDiffBox.children[0].getElementsByClassName("phui-header-header").item(0);
	if (!diffTitle || !diffTitle.textContent) {
		return null;
	}
	const matches = DIFF_PATTERN.exec(diffTitle.textContent);
	if (!matches) {
		return null;
	}
	return matches[1];
}

function getParentFromRevisionPage(): string | null {
	const keyElements = document.getElementsByClassName("phui-property-list-key");
	for (let keyElement of Array.from(keyElements)) {
		if (keyElement.textContent === "Parents ") {
			const parentUrl = ((keyElement.nextSibling as HTMLElement).children[0].children[0] as HTMLLinkElement).href;
			const revisionMatch = PHAB_REVISION_REGEX.exec(parentUrl);
			if (revisionMatch) {
				return revisionMatch[6];
			}
		}
	}
	return null;
}

const PHAB_DIFFUSION_REGEX = /^(https?):\/\/([A-Z\d\.-]{2,})\.([A-Z]{2,})(:\d{2,4})?\/(source|diffusion)\/([A-Za-z0-9]+)\/browse\/([\w-]+)\/([^;$]+)(;[0-9a-f]{40})?(?:\$[0-9]+)?/i;
const PHAB_DIFFERENTIAL_REGEX = /^(https?):\/\/([A-Z\d\.-]{2,})\.([A-Z]{2,})(:\d{2,4})?\/(D[0-9]+)(?:\?(?:(?:id=([0-9]+))|(vs=(?:[0-9]+|on)&id=[0-9]+)))?/i;
const PHAB_REVISION_REGEX = /^(https?):\/\/([A-Z\d\.-]{2,})\.([A-Z]{2,})(:\d{2,4})?\/r([0-9A-z]+)([0-9a-f]{40})/i;
// http://phabricator.aws.sgdev.org/source/nmux/change/master/mux.go
const PHAB_CHANGE_REGEX = /^(https?):\/\/([A-Z\d\.-]{2,})\.([A-Z]{2,})(:\d{2,4})?\/(source|diffusion)\/([A-Za-z0-9]+)\/change\/([\w-]+)\/([^;]+)(;[0-9a-f]{40})?/i;

const COMPARISON_REGEX = /^vs=((?:[0-9]+|on))&id=([0-9]+)/i;

export function getPhabricatorState(loc: Location): PhabUrl | null {
	const diffusionMatches = PHAB_DIFFUSION_REGEX.exec(loc.href);
	if (diffusionMatches) {
		const match = {
			protocol: diffusionMatches[1],
			hostname: diffusionMatches[2],
			extension: diffusionMatches[3],
			port: diffusionMatches[4],
			viewType: diffusionMatches[5],
			repoUri: diffusionMatches[6],
			branch: diffusionMatches[7],
			path: diffusionMatches[8],
			revInUrl: diffusionMatches[9], // only on previous versions
		};
		const sourcegraphUri = phabricatorInstance.getPhabricatorRepoFromMap(match.repoUri);
		if (!sourcegraphUri) {
			console.error(`could not map ${match.repoUri} to a valid git repository location.`);
			return null;
		}
		let phabricatorMode: PhabricatorMode | null = null;
		if (match.viewType === "source" || match.viewType === "diffusion") {
			phabricatorMode = PhabricatorMode.Diffusion;
		}
		if (!phabricatorMode) {
			console.error(`diffusion window.location not recognized.`);
			return null;
		}
		const rev = getRevFromPage();
		if (!rev) {
			console.error("cannot determine revision from page.");
			return null;
		}
		return {
			repoURI: sourcegraphUri,
			branch: match.branch,
			path: match.path,
			mode: phabricatorMode,
			rev: rev,
		} as PhabDiffusionUrl;
	}
	const differentialMatches = PHAB_DIFFERENTIAL_REGEX.exec(loc.href);
	if (differentialMatches) {
		const match = {
			protocol: differentialMatches[1],
			hostname: differentialMatches[2],
			extension: differentialMatches[3],
			port: differentialMatches[4],
			differentialId: differentialMatches[5],
			diffBeingViewed: differentialMatches[6],
			comparison: differentialMatches[7],
		};
		const differentialId = match.differentialId;
		const phabURI = getRepoFromDifferentialPage();
		if (!phabURI) {
			console.error(`repository name not found on page.`);
			return null;
		}
		const diffId = getDiffIdFromDifferentialPage();
		if (!diffId) {
			console.error(`diff id not found on page.`);
		}
		const repoUrl = phabricatorInstance.getPhabricatorRepoFromMap(phabURI);
		if (!repoUrl) {
			console.error(`repository name ${phabURI} could not be mapped to a URL.`);
			return null;
		}
		const stagingUrl = phabricatorInstance.getStagingRepoUriFromRepoUrl(repoUrl);
		if (!stagingUrl) {
			console.error(`repository url ${repoUrl} could not be mapped to a Phabricator staging URL, required for differential views.`);
			return null;
		}
		let baseBranch = `phabricator/base/${diffId}`;
		let headBranch = `phabricator/diff/${diffId}`;

		const maxDiff = getMaxDiffFromTabView();
		const diffLanded = isDifferentialLanded();
		if (diffLanded && !maxDiff) {
			console.error("looking for the final diff id in the revision contents table failed. expected final row to have the commit in the description field.");
			return null;
		}
		if (match.comparison) {
			// urls that looks like this: http://phabricator.aws.sgdev.org/D3?vs=on&id=8&whitespace=ignore-most#toc
			// if the first parameter (vs=) is not 'on', not sure how to handle
			const comparisonMatch = COMPARISON_REGEX.exec(match.comparison)!;
			const leftId = comparisonMatch[1];
			if (leftId !== "on") {
				// haven't figured out how to handle this case yet.
				return null;
			}
			headBranch = `phabricator/diff/${comparisonMatch[2]}`;
			baseBranch = `phabricator/base/${comparisonMatch[2]}`;
			if (diffLanded && maxDiff && comparisonMatch[2] === `${maxDiff.diffId}`) {
				headBranch = maxDiff.revDescription;
				baseBranch = headBranch.concat("~1");
			}
		} else {
			// check if the diff we are viewing is the max diff. if so,
			// right is the merged rev into master, and left is master~1
			if (diffLanded && maxDiff && diffId === `${maxDiff.diffId}`) {
				headBranch = maxDiff.revDescription;
				baseBranch = maxDiff.revDescription.concat("~1");
			}
		}
		return {
			baseRepoURI: stagingUrl,
			baseBranch: baseBranch,
			headRepoURI: stagingUrl, // not clear if this should be repoURI or something else
			headBranch: headBranch, // This will be blank on GitHub, but on a manually staged instance should exist
			differentialId: differentialId,
			mode: PhabricatorMode.Differential,
		} as PhabDifferentialUrl;
	}
	const revisionMatch = PHAB_REVISION_REGEX.exec(loc.href);
	if (revisionMatch) {
		const match = {
			protocol: revisionMatch[1],
			hostname: revisionMatch[2],
			extension: revisionMatch[3],
			port: revisionMatch[4],
			repoUri: revisionMatch[5],
			rev: revisionMatch[6],
		};
		const repoURI = phabricatorInstance.getPhabricatorRepoFromMap(match.repoUri);
		if (!repoURI) {
			console.error(`did not successfully map ${match.repoUri} to repository uri.`);
			return null;
		}
		const childRev = match.rev;
		const parentRev = getParentFromRevisionPage();
		if (!parentRev) {
			console.error(`did not successfully determine parent revision.`);
			return null;
		}
		return {
			repoUri: repoURI,
			parentRev: parentRev,
			childRev: childRev,
			mode: PhabricatorMode.Revision,
		} as PhabRevisionUrl;
	}
	const changeMatch = PHAB_CHANGE_REGEX.exec(loc.href);
	if (changeMatch) {
		const match = {
			protocol: changeMatch[1],
			hostname: changeMatch[2],
			extension: changeMatch[3],
			port: changeMatch[4],
			viewType: changeMatch[5],
			repoUri: changeMatch[6],
			branch: changeMatch[7],
			path: changeMatch[8],
			revInUrl: changeMatch[9], // only on previous versions
		};
		const sourcegraphUri = phabricatorInstance.getPhabricatorRepoFromMap(match.repoUri);
		if (!sourcegraphUri) {
			console.error(`could not map ${match.repoUri} to a valid git repository location.`);
			return null;
		}
		const phabricatorMode = PhabricatorMode.Change;
		if (!phabricatorMode) {
			console.error(`diffusion window.location not recognized.`);
			return null;
		}
		const rev = getRevFromPage();
		if (!rev) {
			console.error("cannot determine revision from page.");
			return null;
		}
		return {
			repoURI: sourcegraphUri,
			branch: match.branch,
			path: match.path,
			mode: phabricatorMode,
			rev: rev,
			prevRev: rev.concat("~1"),
		} as PhabChangeUrl;
	}
	return null;
}

export function getFilepathFromFile(fileContainer: HTMLElement): string {
	return fileContainer.children[3].textContent as string;
}

export function tryGetBlobElement(file: HTMLElement): HTMLElement | null {
	// TODO(@uforic): https://secure.phabricator.com/diffusion/ARC/browse/master/NOTICE , repository-crossreference doesn't work.
	return file.querySelector(".repository-crossreference") as HTMLElement | null;
}

/**
 * getCodeCellsForAnnotation code cells which should be annotated
 */
export function getCodeCellsForAnnotation(table: HTMLTableElement): CodeCell[] {
	const cells: CodeCell[] = [];
	// tslint:disable-next-line:prefer-for-of
	for (const row of Array.from(table.rows)) {
		let line: number; // line number of the current line
		let codeCell: HTMLTableDataCellElement; // the actual cell that has code inside; each row contains multiple columns
		let isAddition: boolean | undefined;
		let isDeletion: boolean | undefined;
		let isBlameEnabled = false;
		if (row.cells[0].classList.contains("diffusion-blame-link")) {
			isBlameEnabled = true;
		}
		line = parseInt(row.cells[isBlameEnabled ? 2 : 0].children[0].textContent as string, 10);
		codeCell = row.cells[isBlameEnabled ? 3 : 1];
		if (!codeCell) {
			continue;
		}

		const innerCode = codeCell.querySelector(".blob-code-inner"); // ignore extraneous inner elements, like "comment" button on diff views
		cells.push({
			cell: (innerCode || codeCell) as HTMLElement,
			line,
			isAddition,
			isDeletion,
		});
	}
	return cells;
}

/**
 * getCodeCellsForAnnotation code cells which should be annotated
 */
export function getCodeCellsForDifferentialAnnotations(table: HTMLTableElement, isSplitView: boolean, isBase: boolean): PhabricatorCodeCell[] {
	const cells: PhabricatorCodeCell[] = [];
	// tslint:disable-next-line:prefer-for-of
	for (const row of Array.from(table.rows)) {
		if (row.getAttribute("data-sigil")) {
			// skip rows that have expander links
			continue;
		}
		if (isSplitView) {
			const baseLine = parseInt(row.cells[0].textContent as string, 10);
			const headLine = parseInt(row.cells[2].textContent as string, 10);
			const baseCodeCell = row.cells[1];
			const headCodeCell = row.cells[4];
			if (isBase && baseLine && baseCodeCell) {
				cells.push({
					cell: baseCodeCell,
					line: baseLine,
					isAddition: false,
					isDeletion: false,
					isLeftColumnInSplit: true,
					isUnified: false,
				});
			}
			if (!isBase && headLine && headCodeCell) {
				cells.push({
					cell: headCodeCell,
					line: headLine,
					isAddition: false,
					isDeletion: false,
					isLeftColumnInSplit: false,
					isUnified: false,
				});
			}
		} else {
			const baseLine = parseInt(row.cells[0].textContent as string, 10);
			const headLine = parseInt(row.cells[1].textContent as string, 10);
			const codeCell = row.cells[3];
			if (isBase && baseLine && codeCell) {
				cells.push({
					cell: codeCell,
					line: baseLine,
					isAddition: false,
					isDeletion: false,
					isLeftColumnInSplit: false,
					isUnified: true,
				});
			} else if (!isBase && headLine && codeCell) {
				if (!codeCell.classList.contains("old") && !codeCell.classList.contains("new")) {
					// We don't want to add the white lines (unchanged) for head and base both times, so opt for base.
					continue;
				}
				cells.push({
					cell: codeCell,
					line: headLine,
					isAddition: false,
					isDeletion: false,
					isLeftColumnInSplit: false,
					isUnified: true,
				});
			}
		}
	}
	return cells;
}

/**
 * This injects code as a script tag into a web page body.
 * Needed to defeat the Phabricator Javelin library.
 */
export function javelinPierce(code: () => void, node: string): void {
	const th = document.getElementsByTagName(node)[0];
	const s = document.createElement("script");
	s.setAttribute("type", "text/javascript");
	s.textContent = code.toString() + ";" + code.name + "();";
	th.appendChild(s);
}

export const PHAB_PAGE_LOAD_EVENT_NAME = "phabPageLoaded";

/**
 * This hooks into the Javelin event queue by adding an additional onload function.
 * Needed to successfully detect when a Phabricator page has loaded.
 * Fires a new event type not caught by Javelin that we can listen to, phabPageLoaded.
 */
export function setupPageLoadListener(): void {
	const JX = (window as any).JX;
	JX.onload(() => document.dispatchEvent(new Event("phabPageLoaded", {})));
}

/**
 * This hacks javelin Stratcom to allow for the detection of blob expansion. Normally,
 * javelin Stratcom kills the mouse click event before it can propogate to detection code.
 * Instead, we check every event that passes by Stratcom and if we see a show-more event,
 * propogate it onwards.
 */
export function expanderListen(): void {
	const JX = (window as any).JX;
	if (JX.Stratcom._dispatchProxyPreExpander) {
		console.error("Error setting up expander listener - _dispatchProxyPreExpander already defined.");
		return;
	}
	JX.Stratcom._dispatchProxyPreExpander = JX.Stratcom._dispatchProxy;
	JX.Stratcom._dispatchProxy = proxyEvent => {
		if (proxyEvent.isNormalClick() && proxyEvent.getNodes()["show-more"]) {
			proxyEvent.__auto__target.parentElement.parentElement.parentElement.parentElement.dispatchEvent(new Event("expandClicked", {}));
		}
		return JX.Stratcom._dispatchProxyPreExpander(proxyEvent);
	};
}

/**
 * This hacks javelin Stratcom to ignore command + click actions on sg-clickable tokens.
 * Without this, two windows open when a user command + clicks on a token.
 */
export function metaClickOverride(): void {
	const JX = (window as any).JX;
	if (JX.Stratcom._dispatchProxyPreMeta) {
		console.error("Error setting up expander listener - _dispatchProxyPreMeta already defined.");
		return;
	}
	JX.Stratcom._dispatchProxyPreMeta = JX.Stratcom._dispatchProxy;
	JX.Stratcom._dispatchProxy = proxyEvent => {
		if (proxyEvent.__auto__type === "click" && proxyEvent.__auto__rawEvent.metaKey && proxyEvent.__auto__target.classList.contains("sg-clickable")) {
			return;
		}
		return JX.Stratcom._dispatchProxyPreMeta(proxyEvent);
	};
}

const USERNAME_URL_PATTERN = /\/p\/([A-Z0-9-]+)/i;
export function getPhabricatorUsername(): string | null {
	const coreMenuItems = document.getElementsByClassName("core-menu-item");
	for (const coreMenuItem of Array.from(coreMenuItems)) {
		const children = coreMenuItem.children;
		if (children.length === 0) {
			continue;
		}
		const possiblePersonUrl = children[0].getAttribute("href");
		if (!possiblePersonUrl) {
			continue;
		}
		const match = USERNAME_URL_PATTERN.exec(possiblePersonUrl);
		if (!match) {
			continue;
		}
		return match[1];
	}
	return null;
}
