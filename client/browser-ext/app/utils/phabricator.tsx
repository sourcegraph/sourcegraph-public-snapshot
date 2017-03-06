import { PhabricatorInstance } from "../../app/utils/classes";
import { CodeCell, PhabDifferentialUrl, PhabDiffusionUrl, PhabRevisionUrl, PhabricatorCodeCell, PhabricatorMode, PhabUrl } from "../../app/utils/types";
import { phabricatorInstance } from "./context";

function getRevFromPage(): string | null {
	const shaPattern = /r([0-9A-z]+)([0-9a-f]{40})/;
	const revElement = document.getElementsByClassName("phui-tag-core").item(0);
	if (!revElement) {
		return null;
	}
	const shaMatch = shaPattern.exec(revElement.children[0].getAttribute("href") as string);
	if (!shaMatch) {
		return null;
	}
	return shaMatch[2];
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
	repoUrl = repoUrl.substr("/source/".length);
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

const PHAB_DIFFUSION_REGEX = /^(https?):\/\/([A-Z\d\.-]{2,})\.([A-Z]{2,})(:\d{2,4})?\/(source|diffusion)\/([A-Za-z0-9]+)\/browse\/([A-Za-z0-9]+)\/(.*)/i;
const PHAB_DIFFERENTIAL_REGEX = /^(https?):\/\/([A-Z\d\.-]{2,})\.([A-Z]{2,})(:\d{2,4})?\/(D[0-9]+)/i;
const PHAB_REVISION_REGEX = /^(https?):\/\/([A-Z\d\.-]{2,})\.([A-Z]{2,})(:\d{2,4})?\/r([0-9A-z]+)([0-9a-f]{40})/i;

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
			console.error(`repository name ${repoUrl} could not be mapped to a URL.`);
			return null;
		}
		const stagingUrl = phabricatorInstance.getStagingRepoUriFromRepoUrl(repoUrl);
		if (!stagingUrl) {
			console.error(`repository url ${stagingUrl} could not be mapped to a Phabricator staging URL, required for differential views.`);
			return null;
		}
		return {
			baseRepoURI: stagingUrl,
			baseBranch: `phabricator/base/${diffId}`,
			headRepoURI: stagingUrl, // not clear if this should be repoURI or something else
			headBranch: `phabricator/diff/${diffId}`, // This will be blank on GitHub, but on a manually staged instance should exist
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
				});
			}
			if (!isBase && headLine && headCodeCell) {
				cells.push({
					cell: headCodeCell,
					line: headLine,
					isAddition: false,
					isDeletion: false,
					isLeftColumnInSplit: false,
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
