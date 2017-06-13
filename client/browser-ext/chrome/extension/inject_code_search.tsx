import * as backend from "../../app/backend";
import { eventLogger, sourcegraphUrl } from "../../app/utils/context";
import { insertAfter } from "../../app/utils/dom";
import * as github from "../../app/utils/github";
import { getDomain, getGitHubRoute, parseURL } from "../../app/utils/index";
import { Domain, GitHubBlobUrl, GitHubMode, GitHubUrl } from "../../app/utils/types";
import { GITHUB_LIGHT_THEME } from "../assets/themes/github_theme";

import * as querystring from "query-string";

const CODE_SEARCH_ELEMENT_ID = "sourcegraph-search-frame";

const GITHUB_REPOSITORY_CONTENT_CONTAINER = ".repository-content";
const GITHUB_CODE_SEARCH_CONTAINER_SELECTOR = ".container.new-discussion-timeline";
const GITHUB_RESULTS_CONTAINER_SELECTOR = ".clearfix.gut-sm";

const GITHUB_HEADER_SELECTOR = ".border-bottom";
const GITHUB_HEADER_HEIGHT = 117;

const GITHUB_FOOTER_SELECTOR = ".site-footer";
const GITHUB_FOOTER_HEIGHT = 139;

const CODE_SEARCH_FEATURE_FLAG_KEY = "sourcegraph-code-search";
const DEFAULT_CODE_SEARCH = "default-github-code-search";

const SOURCEGRAPH_CODE_TOGGLE = "sourcegraph-code-toggle";

/**
 * injectCodeSearch is responsible for injecting our Sourcegraph Code Search into GitHub's DOM.
 */
export function injectCodeSearch(): void {
	if (isCodeSearchURL()) {
		renderSourcegraphSearchTab();
	}

	// Skip rendering all together if this is not a code search URL.
	const { repoURI, path } = parseURL(window.location);
	if (!repoURI) {
		return;
	}

	// Get the parent container for the search frame.
	const repoContent = document.querySelector(GITHUB_REPOSITORY_CONTENT_CONTAINER);
	if (!repoContent) {
		return;
	}
	const searchFrame = createCodeSearchFrame(repoURI, repoContent as HTMLIFrameElement);
	hideCodeSearchFrame(searchFrame);
	const searchQuery = getSearchQuery();
	const frameLocation = searchQuery ?
		`${sourcegraphUrl}/${repoURI}/-/search?config=${searchURLConfig(repoURI, searchQuery)}` :
		`${sourcegraphUrl}/${repoURI}?config=${searchURLConfig(repoURI)}`;

	searchFrame.contentWindow.location.href = frameLocation;

	if (isSourcegraphSearchQuery() && searchQuery) {
		removeGitHubCodeSearchElements();
		showCodeSearchFrameForSearch(searchFrame, repoURI!, searchQuery);
	}
}

window.addEventListener("popstate", (e: any) => {
	if (!e.target || !e.target.location) {
		return;
	}
	if (isCodeSearchURL()) {
		renderSourcegraphSearchTab();
	}
	const { repoURI, path } = parseURL(window.location);
	if (!repoURI) {
		return;
	}

	// Get the parent container for the search frame.
	const repoContent = document.querySelector(GITHUB_REPOSITORY_CONTENT_CONTAINER);
	if (!repoContent) {
		return;
	}
	const searchFrame = createCodeSearchFrame(repoURI, repoContent as HTMLIFrameElement);
	const searchQuery = getSearchQuery();
	if (!searchQuery) {
		return;
	}

	if (isCodeSearchURL(e.target.location) && !isSourcegraphSearchQuery(e.target.location)) {
		hideCodeSearchFrame(searchFrame);
		addGitHubCodeSearchElements();
	}
	if (isCodeSearchURL(e.target.location) && isSourcegraphSearchQuery(e.target.location)) {
		removeGitHubCodeSearchElements();
		showCodeSearchFrameForSearch(searchFrame, repoURI!, searchQuery);
	}
});

/**
 * Returns if the user is currently viewing a GitHub code search URL.
 */
export function isCodeSearchURL(location: any = window.location): boolean {
	const { repoURI } = parseURL(location);
	return Boolean(location.pathname.includes("/search") && repoURI);
}

/**
 * Returns the rendered iframe if one exists.
 */
function renderedSearchFrame(): HTMLIFrameElement | undefined {
	return document.getElementById(CODE_SEARCH_ELEMENT_ID) as HTMLIFrameElement;
}

/**
 * Returns the search navigation bar.
 */
function getSearchNavBar(): Element | null {
	return document.querySelector(".underline-nav");
}

/**
 * Returns the Sourcegraph code toggle part.
 * @param navPart The GitHub navigation bar element.
 */
function createSourcegraphTogglePart(navPart: HTMLElement): HTMLDivElement {
	const toggle = document.createElement("div");
	toggle.id = SOURCEGRAPH_CODE_TOGGLE;
	toggle.style.cursor = "pointer";
	const query = querystring.parse(window.location.search);
	if (isSourcegraphSearchQuery()) {
		const currentSelected = document.querySelector(".underline-nav-item.selected");
		if (currentSelected) {
			currentSelected.className = "underline-nav-item";
		}
		toggle.className = "underline-nav-item selected";
	} else {
		toggle.className = "underline-nav-item";
	}
	// Update search type to Sourcegraph
	const { repoURI } = parseURL(window.location);
	query["type"] = "Sourcegraph";
	const decodedSearch = decodeURIComponent((query["q"] || "" + "").replace(/\+/g, "%20"));
	toggle.onclick = (e) => {
		e.preventDefault();
		eventLogger.logSourcegraphSearchTabClicked({ query: getSearchQuery() });
		const repoContent = document.querySelector(GITHUB_REPOSITORY_CONTENT_CONTAINER);
		if (!repoContent) {
			return;
		}

		window.location.hash = "#sourcegraph";
		renderSourcegraphSearchTab();

		const searchFrame = createCodeSearchFrame(repoURI, repoContent as HTMLIFrameElement);
		if (removeGitHubCodeSearchElements()) {
			showCodeSearchFrameForSearch(searchFrame!, repoURI!, decodedSearch);
		}
	};
	toggle.innerText = "Code (Sourcegraph)";
	backend.searchText(repoURI!, decodedSearch).then((textSearch: backend.ResolvedSearchTextResp) => {
		if (textSearch.results) {
			let totalResults = 0;
			textSearch.results.forEach(file => {
				totalResults += file.lineMatches.length;
			});
			const resultsBubble = document.createElement("span");
			resultsBubble.className = "Counter ml-2";
			resultsBubble.innerText = totalResults.toString();
			toggle.appendChild(resultsBubble);
		}
	});
	return toggle;
}

/**
 * Renders the Sourcegraph code search toggle button.
 */
function renderSourcegraphSearchTab(): void {
	const togglePart = getSourcegraphCodeTogglePart();
	if (togglePart) {
		togglePart.remove();
	}
	const navbar = getSearchNavBar();
	if (navbar) {
		const firstChild = navbar.firstElementChild!;
		if (firstChild.hasChildNodes) {
			const textNode = firstChild.childNodes[0];
			textNode.textContent = "Code (GitHub)";
		}
		insertAfter(createSourcegraphTogglePart(navbar as HTMLElement), navbar.firstChild!.nextSibling!);
	}
}

/**
 * Returns the rendered Sourcegraph code search toggle button.
 */
function getSourcegraphCodeTogglePart(): HTMLAnchorElement | null {
	return document.getElementById(SOURCEGRAPH_CODE_TOGGLE) as HTMLAnchorElement;
}

/**
 * Update's the iframe's visibility to "visible" and updates the iFrame's src attribute to the current GitHub search query.
 * @param searchFrame The iFrame to be set visible.
 * @param repoURI The URI of the repository being searched.
 */
function showCodeSearchFrameForSearch(searchFrame: HTMLIFrameElement, repoURI: string, search: string): void {
	searchFrame.style.visibility = "visible";
	searchFrame.style.position = "";
	const searchEncoded = searchURLConfig(repoURI, search);
	const searchURL = `${sourcegraphUrl}/${repoURI}/-/search?config=${searchEncoded}`;
	eventLogger.logSourcegraphSearch({ query: getSearchQuery() });
}

/**
 * Update's the iframe's visibility to hidden.
 * @param searchFrame The search frame to set hidden.
 */
function hideCodeSearchFrame(searchFrame: HTMLIFrameElement): void {
	searchFrame.style.visibility = "hidden";
}

/**
 * Returns true if the current search query type is sourcegraph.
 */
function isSourcegraphSearchQuery(location: any = window.location): boolean {
	return location.hash === "#sourcegraph";
}

/**
 * Creates the code search iFrame used to display Sourcegraph code search results instead of GitHub's native code search. Default visiblility is hidden.
 * @param repoURI The URI of the current repository.
 * @return {HTMLElement} The code search iFrame.
 */
function createCodeSearchFrame(repoURI: string = "", parent: HTMLElement): HTMLIFrameElement {
	if (renderedSearchFrame()) {
		return renderedSearchFrame()!;
	}
	const searchFrame = document.createElement("iframe") as HTMLIFrameElement;
	searchFrame.id = CODE_SEARCH_ELEMENT_ID;
	searchFrame.style.height = `calc(100vh - ${GITHUB_FOOTER_HEIGHT + GITHUB_HEADER_HEIGHT}px)`;
	searchFrame.style.display = "block";
	searchFrame.style.visibility = "visible";
	searchFrame.style.zIndex = "-500";
	searchFrame.style.position = "absolute";
	searchFrame.style.top = `${GITHUB_HEADER_HEIGHT}px`;
	searchFrame.style.left = "0px";
	searchFrame.style.width = "100%";
	searchFrame.style.bottom = "0px";
	searchFrame.style.border = "none";

	// Style the iframe to fit in the code search UI.
	// searchFrame.style.cssText = "top:0;left:0;display:block;" + `width:100%;height:calc(100vh - ${GITHUB_FOOTER_HEIGHT + GITHUB_HEADER_HEIGHT}px);z-index:1000; border: none;`;
	parent.appendChild(searchFrame);
	const iframeWindow = searchFrame.contentWindow || (searchFrame.contentDocument as any).parentWindow;
	iframeWindow.document.domain = "github.com";

	const searchQuery = getSearchQuery();
	const searchEncoded = searchURLConfig(repoURI, searchQuery, true);
	return searchFrame;
}

/**
 * Returns the encoded configuration parameters.
 * @param repo The current repo to execute the query on.
 * @param query The search query.
 * @param initialStartup The boolean flag to tell the workbench how to load contents.
 */
function searchURLConfig(repo: string, query?: string, initialStartup?: boolean): string {
	const url = query ? `${sourcegraphUrl}/${repo}/-/search?q=${query}&repos=${repo}` : `${sourcegraphUrl}/${repo}`;
	query = decodeURIComponent((query || "" + "").replace(/\+/g, "%20"));
	const queryParams = {
		_: [],
		isInitialStartup: initialStartup,
		workspacePath: `repo://${repo}`,
		url,
		urlRoutePathPrefix: `/${repo}`,
		baseTheme: "vs",
		backgroundColor: "#FFFFF",
		themeData: GITHUB_LIGHT_THEME,
		isChromeExtension: true,
		standaloneWorkbench: true,
		activityBarHidden: true,
	};

	const queryString = encodeURIComponent(JSON.stringify(queryParams));
	return queryString;
}

/**
 * Removes native GitHub code search elements from the DOM prior to injecting our elements.
 * @return {boolean} returns if we 'true' if we removed the elements. Returns 'false' if we did not remove the elements.
 */
function removeGitHubCodeSearchElements(): boolean {
	const searchContainer = document.querySelector(GITHUB_RESULTS_CONTAINER_SELECTOR) as HTMLElement | undefined;
	if (!searchContainer || !searchContainer.parentNode) {
		return false;
	}
	const container = document.querySelector(GITHUB_CODE_SEARCH_CONTAINER_SELECTOR) as HTMLElement;
	if (!container) {
		return false;
	}
	const footer = document.querySelector(GITHUB_FOOTER_SELECTOR) as HTMLElement;
	if (!footer) {
		return false;
	}
	const header = document.querySelector(GITHUB_HEADER_SELECTOR) as HTMLElement;
	if (!header) {
		return false;
	}

	// Update styles.
	container.style.marginTop = "0px";
	if (header.classList) { // kludge
		header.classList.remove("mb-4");
	}
	header.style.marginBottom = "0px";
	container.style.width = "100vw";

	searchContainer.style.visibility = "hidden";
	searchContainer.style.display = "none";
	return true;
}

function addGitHubCodeSearchElements(): boolean {
	const searchContainer = document.querySelector(GITHUB_RESULTS_CONTAINER_SELECTOR) as HTMLElement | undefined;
	if (!searchContainer || !searchContainer.parentNode) {
		return false;
	}
	const container = document.querySelector(GITHUB_CODE_SEARCH_CONTAINER_SELECTOR) as HTMLElement;
	if (!container) {
		return false;
	}
	const footer = document.querySelector(GITHUB_FOOTER_SELECTOR) as HTMLElement;
	if (!footer) {
		return false;
	}
	const header = document.querySelector(GITHUB_HEADER_SELECTOR) as HTMLElement;
	if (!header) {
		return false;
	}

	// Update styles.
	footer.style.marginTop = "40px";
	header.style.marginBottom = "24px";
	container.style.width = "980px";

	searchContainer.style.visibility = "visible";
	searchContainer.style.display = "block";
	return true;
}

/**
 * Returns the current search query from the query params.
 * @returns {string | undefined} The current code search query if one is present.
 */
function getSearchQuery(): string | undefined {
	const queryParams = querystring.parse(window.location.search);
	if (queryParams && queryParams["q"]) {
		return queryParams["q"];
	}
}
