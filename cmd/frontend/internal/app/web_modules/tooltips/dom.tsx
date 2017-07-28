import { fetchJumpURL } from "app/backend/lsp";
import { triggerReferences } from "app/references";
import { clearTooltip, store, TooltipState } from "app/tooltips/store";
import * as styles from "app/tooltips/styles";
import { getModeFromExtension } from "app/util";
// import { getAssetURL } from "app/util/assets";
// import { eventLogger, searchEnabled } from "app/util/context";
import { highlightBlock } from "highlight.js";
import * as marked from "marked";

let tooltip: HTMLElement;
let loadingTooltip: HTMLElement;
let tooltipActions: HTMLElement;
let j2dAction: HTMLAnchorElement;
let findRefsAction: HTMLAnchorElement;
let searchAction: HTMLAnchorElement;
let moreContext: HTMLElement;

const searchIconSVG = '<svg width="12px" height="12px"><path fill="#FFFFFF" xmlns="http://www.w3.org/2000/svg" id="path13_fill" d="M 4.75021 4.65905e-06C 7.36999 -0.00361534 9.49667 2.1172 9.50027 4.73698C 9.50167 5.7595 9.17264 6.7551 8.5622 7.5754L 11.1265 9.74432C 11.5382 10.0957 11.5872 10.7144 11.2358 11.1261C 10.8844 11.5379 10.2657 11.5868 9.85399 11.2355C 9.81473 11.202 9.77819 11.1654 9.74467 11.1261L 7.5752 8.56228C 5.46856 10.1236 2.49507 9.68156 0.933725 7.5749C -0.627615 5.46824 -0.185555 2.49476 1.92111 0.933425C 2.73957 0.326825 3.73145 -0.000435341 4.75021 4.65905e-06ZM 4.75021 8.5C 6.82127 8.5 8.50021 6.82106 8.50021 4.75C 8.50021 2.67894 6.82127 1 4.75021 1C 2.67915 1 1.00021 2.67894 1.00021 4.75C 1.00023 6.82106 2.67915 8.49998 4.75021 8.49998L 4.75021 8.5Z"/></svg>';

const referencesIconSVG = '<svg width="12px" height="8px"><path fill="#FFFFFF" xmlns="http://www.w3.org/2000/svg" id="path15_fill" d="M 6.00625 8C 2.33125 8 0.50625 5.075 0.05625 4.225C -0.01875 4.075 -0.01875 3.9 0.05625 3.775C 0.50625 2.925 2.33125 0 6.00625 0C 9.68125 0 11.5063 2.925 11.9563 3.775C 12.0312 3.925 12.0312 4.1 11.9563 4.225C 11.5063 5.075 9.68125 8 6.00625 8ZM 6.00625 1.25C 4.48125 1.25 3.25625 2.475 3.25625 4C 3.25625 5.525 4.48125 6.75 6.00625 6.75C 7.53125 6.75 8.75625 5.525 8.75625 4C 8.75625 2.475 7.53125 1.25 6.00625 1.25ZM 6.00625 5.75C 5.03125 5.75 4.25625 4.975 4.25625 4C 4.25625 3.025 5.03125 2.25 6.00625 2.25C 6.98125 2.25 7.75625 3.025 7.75625 4C 7.75625 4.975 6.98125 5.75 6.00625 5.75Z"/></svg>';

const definitionIconSVG = '<svg width="11px" height="9px"><path fill="#FFFFFF" xmlns="http://www.w3.org/2000/svg" id="path10_fill" d="M 6.325 8.4C 6.125 8.575 5.8 8.55 5.625 8.325C 5.55 8.25 5.5 8.125 5.5 8L 5.5 6C 2.95 6 1.4 6.875 0.825 8.7C 0.775 8.875 0.6 9 0.425 9C 0.2 9 -4.44089e-16 8.8 -4.44089e-16 8.575C -4.44089e-16 8.575 -4.44089e-16 8.575 -4.44089e-16 8.55C 0.125 4.825 1.925 2.675 5.5 2.5L 5.5 0.5C 5.5 0.225 5.725 8.88178e-16 6 8.88178e-16C 6.125 8.88178e-16 6.225 0.05 6.325 0.125L 10.825 3.875C 11.025 4.05 11.075 4.375 10.9 4.575C 10.875 4.6 10.85 4.625 10.825 4.65L 6.325 8.4Z"/></svg>';

/**
 * createTooltips initializes the DOM elements used for the hover
 * tooltip and "Loading..." text indicator, adding the former
 * to the DOM (but hidden). It is idempotent.
 */
export function createTooltips(): void {
	if (tooltip) {
		return; // idempotence
	}

	tooltip = document.createElement("DIV");
	Object.assign(tooltip.style, styles.tooltip);
	tooltip.classList.add("sg-tooltip");
	tooltip.style.visibility = "hidden";

	document.querySelector("#blob-table")!.appendChild(tooltip);

	loadingTooltip = document.createElement("DIV");
	loadingTooltip.appendChild(document.createTextNode("Loading..."));
	Object.assign(loadingTooltip.style, styles.loadingTooltip);

	tooltipActions = document.createElement("DIV");
	Object.assign(tooltipActions.style, styles.tooltipActions);

	moreContext = document.createElement("DIV");
	Object.assign(moreContext.style, styles.tooltipMoreActions);
	moreContext.appendChild(document.createTextNode("Click for more actions"));

	const definitionIcon = document.createElement("svg");
	definitionIcon.innerHTML = definitionIconSVG;
	Object.assign(definitionIcon.style, styles.definitionIcon);

	j2dAction = document.createElement("A") as HTMLAnchorElement;
	j2dAction.appendChild(definitionIcon);
	j2dAction.appendChild(document.createTextNode("Go to definition"));
	j2dAction.className = `btn btn-sm BtnGroup-item`;
	Object.assign(j2dAction.style, styles.tooltipAction);
	Object.assign(j2dAction.style, styles.tooltipActionNotLast);
	j2dAction.onclick = (e) => {
		e.preventDefault();
		const { data, context } = store.getValue();
		if (data && context && context.coords && context.path && context.repoRevSpec) {
			fetchJumpURL(context.coords.char, context.path, context.coords.line, context.repoRevSpec)
				.then((defUrl) => {
					// eventLogger.logJumpToDef({ ...getTooltipEventProperties(data, context), hasResolvedJ2D: Boolean(defUrl) });
					if (defUrl) {
						window.location.href = defUrl;
						clearTooltip();
						// const withModifierKey = isMouseEventWithModifierKey(e);
						// openSourcegraphTab(defUrl, withModifierKey);
					}
				});
		}
	};

	const referencesIcon = document.createElement("svg");
	referencesIcon.innerHTML = referencesIconSVG;
	Object.assign(referencesIcon.style, styles.referencesIcon);

	findRefsAction = document.createElement("A") as HTMLAnchorElement;
	findRefsAction.appendChild(referencesIcon);
	findRefsAction.appendChild(document.createTextNode("Find all references"));
	Object.assign(findRefsAction.style, styles.tooltipAction);
	Object.assign(findRefsAction.style, styles.tooltipActionNotLast);
	findRefsAction.className = `btn btn-sm BtnGroup-item`;
	findRefsAction.onclick = (e) => {
		e.preventDefault();
		const { context } = store.getValue();
		if (!context || !context.coords) {
			return;
		}
		const loc = {
			uri: context.repoRevSpec.repoURI,
			rev: context.repoRevSpec.rev,
			path: context.path,
			line: context.coords.line,
			char: context.coords.char,
		};
		triggerReferences({ loc, word: context.coords.word }, true);
	};

	const searchIcon = document.createElement("svg");
	searchIcon.innerHTML = searchIconSVG;
	Object.assign(searchIcon.style, styles.searchIcon);

	searchAction = document.createElement("A") as HTMLAnchorElement;
	searchAction.appendChild(searchIcon);
	searchAction.appendChild(document.createTextNode("Search..."));
	Object.assign(searchAction.style, styles.tooltipAction);
	searchAction.className = `btn btn-sm BtnGroup-item`;
	searchAction.onclick = () => {
		// e.preventDefault();
		// const searchText = store.getValue().context && store.getValue().context!.selectedText ?
		// 	store.getValue().context!.selectedText! :
		// 	store.getValue().target!.textContent!;
		// const { data, context } = store.getValue();
		// if (data && context && context.repoRevSpec) {
		// 	// const url = `/${context.repoRevSpec.repoURI}@${context.repoRevSpec.rev}?q=${encodeURIComponent(searchText)}`;
		// 	// const withModifierKey = isMouseEventWithModifierKey(e);
		// 	// return;
		// }
	};

	tooltipActions.appendChild(j2dAction);
	tooltipActions.appendChild(findRefsAction);
	tooltipActions.appendChild(searchAction);
}

function constructBaseTooltip(): void {
	tooltip.appendChild(loadingTooltip);
	tooltip.appendChild(moreContext);
	tooltip.appendChild(tooltipActions);
}

/**
 * hideTooltip makes the tooltip on the DOM invisible.
 */
export function hideTooltip(): void {
	if (!tooltip) {
		return;
	}

	while (tooltip.firstChild) {
		tooltip.removeChild(tooltip.firstChild);
	}
	tooltip.style.visibility = "hidden"; // prevent black dot of empty content
}

/**
 * updateTooltip displays the appropriate tooltip given current state (and may hide
 * the tooltip if no text is available).
 */
function updateTooltip(state: TooltipState): void {
	hideTooltip(); // hide before updating tooltip text

	const { target, data, docked, context } = state;

	if (!target) {
		// no target to show hover for; tooltip is hidden
		return;
	}
	if (!data) {
		// no data; bail
		return;
	}
	if (!context || (context.selectedText && context.selectedText.trim()) === "") {
		// no context or selected text is only whitespace; bail
		return;
	}

	constructBaseTooltip();
	loadingTooltip.style.display = data.loading ? "block" : "none";
	moreContext.style.display = docked || data.loading ? "none" : "flex";
	tooltipActions.style.display = docked ? "flex" : "none";

	if (context && context.selectedText) {
		j2dAction.style.display = "none";
		findRefsAction.style.display = "none";
	} else {
		j2dAction.style.display = "block";
		findRefsAction.style.display = "block";
	}

	j2dAction.href = data.j2dUrl ? data.j2dUrl : "";

	if (data && context && context.coords && context.path && context.repoRevSpec) {
		findRefsAction.href = `/${context.repoRevSpec.repoURI}@${context.repoRevSpec.rev}/-/blob/${context.path}#L${context.coords.line}:${context.coords.char}$references`;
	} else {
		findRefsAction.href = "";
	}

	const searchText = context!.selectedText ? context!.selectedText! : target!.textContent!;
	if (searchText) {
		searchAction.href = `/${context.repoRevSpec.repoURI}@${context.repoRevSpec.rev}/-/blob/${context.path}&q=${searchText}`;
	} else {
		searchAction.href = "";
	}

	if (!data.loading) {
		loadingTooltip.style.visibility = "hidden";

		if (!data.title) {
			// no tooltip text / search context; tooltip is hidden
			return;
		}

		const container = document.createElement("DIV");
		Object.assign(container.style, styles.divider);

		const tooltipText = document.createElement("DIV");
		tooltipText.className = `${getModeFromExtension(context.path)}`;
		Object.assign(tooltipText.style, styles.tooltipTitle);
		tooltipText.appendChild(document.createTextNode(data.title));

		const icon = document.createElement("img");
		// icon.src = getAssetURL("sourcegraph-mark.svg");
		Object.assign(icon.style, styles.sourcegraphIcon);

		container.appendChild(icon);
		container.appendChild(tooltipText);
		tooltip.insertBefore(container, moreContext);

		const closeContainer = document.createElement("a");
		Object.assign(closeContainer.style, styles.closeIcon);
		closeContainer.onclick = () => clearTooltip();

		if (docked) {
			const closeButton = document.createElement("img");
			// closeButton.src = getAssetURL("close-icon.svg");
			closeContainer.appendChild(closeButton);
			container.appendChild(closeContainer);
		}

		highlightBlock(tooltipText);

		if (data.doc) {
			const tooltipDoc = document.createElement("DIV");
			Object.assign(tooltipDoc.style, styles.tooltipDoc);
			tooltipDoc.innerHTML = marked(data.doc, { gfm: true, breaks: true, sanitize: true });
			tooltip.insertBefore(tooltipDoc, moreContext);

			// Handle scrolling ourselves so that scrolling to the bottom of
			// the tooltip documentation does not cause the page to start
			// scrolling (which is a very jarring experience).
			tooltip.addEventListener("wheel", (e: WheelEvent) => {
				e.preventDefault();
				tooltipDoc.scrollTop += e.deltaY;
			});
		}
	} else {
		loadingTooltip.style.visibility = "visible";
	}

	const blobScroll = document.querySelector("#blob-table")!; // the scroll view
	const blobTable = blobScroll.querySelector("table")!; // the overflowing content (can have negative positions)
	const tableBound = blobTable.getBoundingClientRect();


	// Anchor it horizontally, prior to rendering to account for wrapping
	// changes to vertical height if the tooltip is at the edge of the viewport.
	const targetBound = target.getBoundingClientRect();
	const relLeft = targetBound.left - tableBound.left;
	tooltip.style.left = relLeft + "px";

	// Anchor the tooltip vertically.
	const tooltipBound = tooltip.getBoundingClientRect();
	const relTop = targetBound.top - tableBound.top;
	const margin = 5;
	let tooltipTop = relTop - (tooltipBound.height + margin);
	if (tooltipTop < 0) {
		// Tooltip wouldn't be visible from the top, so display it at the
		// bottom.
		const relBottom = targetBound.bottom - tableBound.top;
		tooltipTop = relBottom - margin;
	}
	tooltip.style.top = tooltipTop + "px";

	// Make it all visible to the user.
	tooltip.style.visibility = "visible";
}

window.addEventListener("keyup", (e: KeyboardEvent) => {
	if (e.keyCode === 27) {
		clearTooltip();
	}
});

store.subscribe(updateTooltip);
