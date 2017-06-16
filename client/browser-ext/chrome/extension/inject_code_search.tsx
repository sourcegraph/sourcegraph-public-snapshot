import * as backend from "../../app/backend";
import { eventLogger, searchEnabled, sourcegraphUrl } from "../../app/utils/context";
import { insertAfter } from "../../app/utils/dom";
import { getPlatformName, parseURL } from "../../app/utils/index";
import { GITHUB_LIGHT_THEME } from "../assets/themes/github_theme";

import * as querystring from "query-string";

const CODE_SEARCH_ELEMENT_ID = "sourcegraph-search-frame";

const GITHUB_REPOSITORY_CONTENT_CONTAINER = ".repository-content";
const GITHUB_CODE_SEARCH_CONTAINER_SELECTOR = ".container.new-discussion-timeline";
const GITHUB_RESULTS_CONTAINER_SELECTOR = ".clearfix.gutter-condensed";

const GITHUB_HEADER_SELECTOR = ".border-bottom";
const GITHUB_HEADER_HEIGHT = 117;

const GITHUB_FOOTER_SELECTOR = ".site-footer-container";

const SOURCEGRAPH_CODE_TOGGLE = "sourcegraph-code-toggle";
const SOURCEGRAPH_AUTH_PAGE = "sourcegraph-auth-page";

/**
 * injectCodeSearch is responsible for injecting our Sourcegraph Code Search into GitHub's DOM.
 */
export function injectCodeSearch(): void {
	if (!searchEnabled) {
		return;
	}

	if (isCodeSearchURL()) {
		renderSourcegraphSearchTab();
	}

	// Skip rendering all together if this is not a code search URL.
	const { repoURI } = parseURL(window.location);
	if (!repoURI) {
		return;
	}

	// Get the parent container for the search frame.
	const repoContent = document.querySelector(GITHUB_REPOSITORY_CONTENT_CONTAINER);
	if (!repoContent) {
		return;
	}
	const searchFrame = createCodeSearchFrame(repoContent as HTMLIFrameElement);
	hideCodeSearchFrame();
	const searchQuery = getSearchQuery();
	const frameLocation = searchQuery ?
		`${sourcegraphUrl}/${repoURI}?config=${searchURLConfig(repoURI, searchQuery)}&q=${searchQuery}` :
		`${sourcegraphUrl}/${repoURI}?config=${searchURLConfig(repoURI)}`;

	searchFrame.contentWindow.location.href = frameLocation;

	if (isSourcegraphSearchQuery() && searchQuery) {
		removeGitHubCodeSearchElements();
		showCodeSearchFrameForSearch(searchFrame);
	}
}

window.addEventListener("popstate", (e: any) => {
	if (!e.target || !e.target.location) {
		return;
	}
	if (isCodeSearchURL()) {
		renderSourcegraphSearchTab();
	}
	const { repoURI } = parseURL(window.location);
	if (!repoURI) {
		return;
	}

	// Get the parent container for the search frame.
	const repoContent = document.querySelector(GITHUB_REPOSITORY_CONTENT_CONTAINER);
	if (!repoContent) {
		return;
	}
	const searchFrame = createCodeSearchFrame(repoContent as HTMLIFrameElement);
	const searchQuery = getSearchQuery();
	if (!searchQuery) {
		return;
	}

	const authPage = createAuthPage(repoContent as HTMLElement);
	if (isCodeSearchURL(e.target.location) && !isSourcegraphSearchQuery(e.target.location)) {
		hideAuthPage(authPage);
		hideCodeSearchFrame();
		addGitHubCodeSearchElements();
	}
	if (isCodeSearchURL(e.target.location) && isSourcegraphSearchQuery(e.target.location)) {
		removeGitHubCodeSearchElements();
		showCodeSearchFrameForSearch(searchFrame);
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
function createSourcegraphTogglePart(): HTMLDivElement {
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

		const searchFrame = createCodeSearchFrame(repoContent as HTMLIFrameElement);
		if (removeGitHubCodeSearchElements()) {
			showCodeSearchFrameForSearch(searchFrame!);
		}
	};
	toggle.innerText = "Code (Sourcegraph)";
	backend.searchText(repoURI!, decodedSearch).then((textSearch: backend.ResolvedSearchTextResp) => {
		if (textSearch.notFound) {
			hideCodeSearchFrame();
			if (isSourcegraphSearchQuery()) {
				const repoContent = document.querySelector(GITHUB_REPOSITORY_CONTENT_CONTAINER);
				if (repoContent) {
					const authPage = createAuthPage(repoContent as HTMLElement);
					showAuthPage(authPage);
				}
			}
			return;
		}

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
		const firstChild = navbar.firstElementChild! as HTMLElement;
		if (isGitHubCodeSearch()) {
			firstChild.className = "underline-nav-item selected";
		}
		firstChild.onclick = function(e: MouseEvent): void {
			window.location.hash = "";
			if (isGitHubCodeSearch()) {
				e.preventDefault();
			}
		};
		if (firstChild.hasChildNodes) {
			const textNode = firstChild.childNodes[0];
			textNode.textContent = "Code (GitHub)";
		}
		insertAfter(createSourcegraphTogglePart(), navbar.firstChild!.nextSibling!);
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
function showCodeSearchFrameForSearch(searchFrame: HTMLIFrameElement): void {
	searchFrame.style.visibility = "visible";
	searchFrame.style.position = "";
	searchFrame.style.height = `calc(100vh - ${GITHUB_HEADER_HEIGHT}px)`;
	eventLogger.logSourcegraphSearch({ query: getSearchQuery() });
}

/**
 * Update's the iframe's visibility to hidden.
 * @param searchFrame The search frame to set hidden.
 */
function hideCodeSearchFrame(): void {
	const searchFrame = renderedSearchFrame();
	if (!searchFrame) {
		return;
	}
	searchFrame.style.visibility = "hidden";
	searchFrame.style.height = "0px";
}

/**
 * Updates the auth page's visibility to visible.
 * @param authPage The auth page container element to show.
 */
function showAuthPage(authPage: HTMLElement): void {
	authPage.style.visibility = "visible";
	authPage.style.display = "block";
	authPage.style.position = "absolute";
}

/**
 * Updates the auth page's visibility to hidden.
 * @param authPage The auth page container element to hide.
 */
function hideAuthPage(authPage: HTMLElement): void {
	authPage.style.visibility = "hidden";
	authPage.style.display = "none";
}

/**
 * Returns true if the current search query type is sourcegraph.
 */
function isSourcegraphSearchQuery(location: any = window.location): boolean {
	return location.hash === "#sourcegraph";
}

/**
 * Returns true if the current search query type is GitHub code search.
 */
function isGitHubCodeSearch(): boolean {
	const query = querystring.parse(window.location.search);
	return !query["type"] || query["type"] === "Code";
}

/**
 * Creates the auth page to be rendered in place of the iframe when a user is viewing a private repository, but not
 * authed with the chrome ext
 */
function createAuthPage(parent: HTMLElement): HTMLElement {
	const authPage = document.getElementById(SOURCEGRAPH_AUTH_PAGE);
	if (authPage) {
		return authPage;
	}

	const container = document.createElement("div");
	container.id = SOURCEGRAPH_AUTH_PAGE;
	container.style.height = `calc(100vh - ${GITHUB_HEADER_HEIGHT}px)`;
	container.style.width = "100%";
	container.style.display = "none";
	container.style.visibility = "hidden";
	container.style.textAlign = "center";
	container.style.top = `${GITHUB_HEADER_HEIGHT}px`;

	const icon = document.createElement("img");
	icon.src = (window as any).chrome.extension.getURL("img/sourcegraph-mark.svg");
	icon.style.display = "block";
	icon.style.margin = "auto";
	icon.style.width = "125px";
	icon.style.height = "125px";
	icon.style.position = "relative";
	icon.style.top = "50px";
	icon.style.paddingBottom = "25px";

	const header = document.createElement("div");
	header.style.display = "block";
	header.style.margin = "auto";
	header.style.paddingBottom = "15px";
	header.style.paddingTop = "40px";

	const headerText = document.createElement("h2");
	headerText.innerHTML = "Sign in to view search results";
	header.appendChild(headerText);

	const authButton = document.createElement("button") as HTMLElement;
	authButton.style.width = "200px";
	authButton.style.height = "35px";
	authButton.style.position = "relative";
	authButton.innerHTML = "Sign in";
	authButton.className = "btn btn-sm btn-primary";
	authButton.onclick = () => {
		window.location.href = `${sourcegraphUrl}/login?private=true&utm_source=${getPlatformName()}`;
	};

	container.appendChild(icon);
	container.appendChild(header);
	container.appendChild(authButton);
	parent.appendChild(container);
	return container;
}

/**
 * Creates the code search iFrame used to display Sourcegraph code search results instead of GitHub's native code search. Default visiblility is hidden.
 * @param repoURI The URI of the current repository.
 * @return {HTMLElement} The code search iFrame.
 */
function createCodeSearchFrame(parent: HTMLElement): HTMLIFrameElement {
	if (renderedSearchFrame()) {
		return renderedSearchFrame()!;
	}
	const searchFrame = document.createElement("iframe") as HTMLIFrameElement;
	searchFrame.id = CODE_SEARCH_ELEMENT_ID;
	searchFrame.style.height = `calc(100vh - ${GITHUB_HEADER_HEIGHT}px)`;
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
	return searchFrame;
}

/**
 * Returns the encoded configuration parameters.
 * @param repo The current repo to execute the query on.
 * @param query The search query.
 * @param initialStartup The boolean flag to tell the workbench how to load contents.
 */
function searchURLConfig(repo: string, query?: string, initialStartup?: boolean): string {
	const url = query ? `${sourcegraphUrl}/${repo}?q=${query}` : `${sourcegraphUrl}/${repo}`;
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
	footer.style.visibility = "hidden";
	footer.style.display = "none";

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
	footer.style.display = "block";
	footer.style.visibility = "visible";

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
