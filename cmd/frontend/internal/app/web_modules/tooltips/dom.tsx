import { fetchDependencyReferences } from "app/backend";
import { fetchJumpURL } from "app/backend/lsp";
import { setReferences, store as referencesStore } from "app/references/store";
import { clearTooltip, store, TooltipState } from "app/tooltips/store";
import * as styles from "app/tooltips/styles";
import { getModeFromExtension } from "app/util";
// import { getAssetURL } from "app/util/assets";
// import { eventLogger, searchEnabled, sourcegraphUrl } from "app/util/context";
import { highlightBlock } from "highlight.js";
import * as marked from "marked";

let tooltip: HTMLElement;
let loadingTooltip: HTMLElement;
let tooltipActions: HTMLElement;
let j2dAction: HTMLAnchorElement;
let findRefsAction: HTMLAnchorElement;
let searchAction: HTMLAnchorElement;
let moreContext: HTMLElement;

const sourcegraphUrl = "http://localhost:3080";

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
	document.body.appendChild(tooltip);

	loadingTooltip = document.createElement("DIV");
	loadingTooltip.appendChild(document.createTextNode("Loading..."));
	Object.assign(loadingTooltip.style, styles.loadingTooltip);

	tooltipActions = document.createElement("DIV");
	Object.assign(tooltipActions.style, styles.tooltipActions);

	moreContext = document.createElement("DIV");
	Object.assign(moreContext.style, styles.tooltipMoreActions);
	moreContext.appendChild(document.createTextNode("Click for more actions"));

	j2dAction = document.createElement("A") as HTMLAnchorElement;
	j2dAction.appendChild(document.createTextNode("Go to Definition"));
	j2dAction.className = `btn btn-sm BtnGroup-item`;
	Object.assign(j2dAction.style, styles.tooltipAction);
	j2dAction.onclick = (e) => {
		e.preventDefault();
		const { data, context } = store.getValue();
		if (data && context && context.coords && context.path && context.repoRevSpec) {
			fetchJumpURL(context.coords.char, context.path, context.coords.line, context.repoRevSpec)
				.then((defUrl) => {
					// eventLogger.logJumpToDef({ ...getTooltipEventProperties(data, context), hasResolvedJ2D: Boolean(defUrl) });
					if (defUrl) {
						window.location.href = defUrl;
						// const withModifierKey = isMouseEventWithModifierKey(e);
						// openSourcegraphTab(defUrl, withModifierKey);
					}
				});
		}
	};

	findRefsAction = document.createElement("A") as HTMLAnchorElement;
	findRefsAction.appendChild(document.createTextNode("Find References"));
	Object.assign(findRefsAction.style, styles.tooltipAction);
	findRefsAction.className = `btn btn-sm BtnGroup-item`;
	findRefsAction.onclick = (e) => {
		e.preventDefault();
		const { data, context } = store.getValue();
		console.log("setting references to be", { ...referencesStore.getValue(), docked: true });
		if (data && context && context.coords && context.path && context.repoRevSpec) {
			setReferences({ ...referencesStore.getValue(), docked: true, context: context as any });
			// eventLogger.logFindRefs({ ...getTooltipEventProperties(data, context) });
			fetchDependencyReferences(context.repoRevSpec.repoURI, context.repoRevSpec.rev, context.path, 40, 25).then(() => {
				console.log("HELLO SENOR!!");
			});
			// fetchReferences(context.coords.char, context.path, context.coords.line, context.repoRevSpec)
			// 	.then((references) => {
			// 		console.log("GOT REFERENCES", references);
			// 	});
			// const url = `${sourcegraphUrl}/${context.repoRevSpec.repoURI}@${context.repoRevSpec.rev}/-/blob/${context.path}?utm_source=${getPlatformName()}#L${context.coords.line}:${context.coords.char}$references`;
			// const withModifierKey = isMouseEventWithModifierKey(e);
			// openSourcegraphTab(url, withModifierKey);
		}
	};

	searchAction = document.createElement("A") as HTMLAnchorElement;
	searchAction.appendChild(document.createTextNode("Search"));
	Object.assign(searchAction.style, styles.tooltipAction);
	searchAction.className = `btn btn-sm BtnGroup-item`;
	searchAction.onclick = () => {
		// e.preventDefault();
		// const searchText = store.getValue().context && store.getValue().context!.selectedText ?
		// 	store.getValue().context!.selectedText! :
		// 	store.getValue().target!.textContent!;
		// const { data, context } = store.getValue();
		// if (data && context && context.repoRevSpec) {
		// 	// const url = `${sourcegraphUrl}/${context.repoRevSpec.repoURI}@${context.repoRevSpec.rev}?q=${encodeURIComponent(searchText)}`;
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
		findRefsAction.href = `${sourcegraphUrl}/${context.repoRevSpec.repoURI}@${context.repoRevSpec.rev}/-/blob/${context.path}#L${context.coords.line}:${context.coords.char}$references`;
	} else {
		findRefsAction.href = "";
	}

	const searchText = context!.selectedText ? context!.selectedText! : target!.textContent!;
	if (searchText) {
		searchAction.href = `${sourcegraphUrl}/${context.repoRevSpec.repoURI}@${context.repoRevSpec.rev}/-/blob/${context.path}&q=${searchText}`;
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
		}
	} else {
		loadingTooltip.style.visibility = "visible";
	}

	// Anchor it horizontally, prior to rendering to account for wrapping
	// changes to vertical height if the tooltip is at the edge of the viewport.
	const targetBound = target.getBoundingClientRect();
	tooltip.style.left = (targetBound.left + window.scrollX) + "px";

	// Anchor the tooltip vertically.
	const tooltipBound = tooltip.getBoundingClientRect();
	tooltip.style.top = (targetBound.top - (tooltipBound.height + 5) + window.scrollY) + "px";

	// Make it all visible to the user.
	tooltip.style.visibility = "visible";
}

window.addEventListener("keyup", (e: KeyboardEvent) => {
	if (e.keyCode === 27) {
		clearTooltip();
	}
});

store.subscribe(updateTooltip);
