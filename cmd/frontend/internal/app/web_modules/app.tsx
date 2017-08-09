import { Tree, TreeHeader } from "@sourcegraph/components";
import { content, flex, vertical } from "csstips";
import * as moment from "moment";
import * as React from "react";
import { render } from "react-dom";
import * as backend from "sourcegraph/backend";
import * as xhr from "sourcegraph/backend/xhr";
import { triggerBlame } from "sourcegraph/blame";
import { injectReferencesWidget } from "sourcegraph/references/inject";
import { injectAdvancedSearchDrawer, injectAdvancedSearchToggle, injectSearchForm, injectSearchInputHandler, injectSearchResults } from "sourcegraph/search/inject";
import { injectShareWidget } from "sourcegraph/share";
import { addAnnotations } from "sourcegraph/tooltips";
import { handleQueryEvents } from "sourcegraph/tracking/analyticsUtils";
import { events, viewEvents } from "sourcegraph/tracking/events";
import { getModeFromExtension, getPathExtension, supportedExtensions } from "sourcegraph/util";
import * as activeRepos from "sourcegraph/util/activeRepos";
import { pageVars } from "sourcegraph/util/pageVars";
import { sourcegraphContext } from "sourcegraph/util/sourcegraphContext";
import * as syntaxhighlight from "sourcegraph/util/syntaxhighlight";
import { CodeCell } from "sourcegraph/util/types";
import * as url from "sourcegraph/util/url";
import { style } from "typestyle";

window.onhashchange = (hash) => {
	const oldURL = url.parseBlob(hash.oldURL!);
	const newURL = url.parseBlob(hash.newURL!);
	if (!newURL.path || !newURL.line) {
		return;
	}
	if (oldURL.line === newURL.line) {
		// prevent e.g. re-scrolling to same line on toggling refs group
		//
		// also prevent highlightLine from retriggering onhashchange
		// recursively.
		return;
	}
	const cells = getCodeCellsForAnnotation();
	highlightAndScrollToLine(newURL.uri!, pageVars.CommitID, newURL.path, newURL.line, cells);
};

window.addEventListener("DOMContentLoaded", () => {
	registerListeners();
	xhr.useAccessToken(sourcegraphContext.accessToken);

	// Be a bit proactive and try to fetch/store active repos now. This helps
	// on the first search query, and when the data in local storage is stale.
	activeRepos.get();

	if (window.location.pathname === "/") {
		viewEvents.Home.log();
		injectSearchForm();
	} else {
		injectSearchInputHandler();
		injectAdvancedSearchToggle();
		injectAdvancedSearchDrawer();
	}
	if (window.location.pathname === "/search") {
		viewEvents.SearchResults.log();
		injectSearchResults();
	}

	const cloning = document.querySelector("#cloning");
	if (cloning) {
		// TODO: Actually poll the backend instead of just reloading the page
		// every 5s.
		setTimeout(() => {
			window.location.reload(false);
		}, 5000);
	}

	injectTreeViewer();
	injectReferencesWidget();
	injectShareWidget();
	const u = url.parseBlob();
	if (u.uri && u.path) {
		const blob = document.querySelector("#blob") as HTMLElement;
		highlightAsync(u.path, blob.textContent!);
		syntaxhighlight.wait().then(() => {
			// blob view, add tooltips
			const rev = pageVars.Rev;
			const commitID = pageVars.CommitID;
			const cells = getCodeCellsForAnnotation();
			if (supportedExtensions.has(getPathExtension(u.path!))) {
				addAnnotations(u.path!, { repoURI: u.uri!, rev: rev, commitID: commitID }, cells);
			}
			if (u.line) {
				highlightAndScrollToLine(u.uri!, commitID, u.path!, u.line, cells);
			}

			// Log blob view
			viewEvents.Blob.log({ repo: u.uri!, commitID, path: u.path!, language: getPathExtension(u.path!) });

			// Add click handlers to all lines of code, which highlight and add
			// blame information to the line.
			Array.from(document.querySelectorAll(".blobview tr")).forEach((tr: HTMLElement, index: number) => {
				tr.addEventListener("click", () => {
					if (u.uri && u.path) {
						highlightLine(u.uri, commitID, u.path, index + 1, cells);
					}
				});
			});
		});
	} else if (u.uri) {
		// tree view
		viewEvents.Tree.log();
	}

	// Log events, if necessary, based on URL querystring, and strip tracking-related parameters
	// from the URL and browser history
	// Note that this is a destructive operation (it changes the page URL and replaces browser state)
	handleQueryEvents(window.location.href);

});

function injectTreeViewer(): void {
	const mount = document.querySelector("#tree-viewer");
	if (!mount) {
		return;
	}

	const repoURL = url.parse();
	const blobURL = url.parseBlob();
	const treeURL = url.parseTree();
	const uri = blobURL.uri || treeURL.uri || repoURL.uri;
	const rev = blobURL.rev || treeURL.rev || repoURL.rev;
	const path = blobURL.path || treeURL.path || "/";

	// Force show the tree viewer on any non-blob page.
	const forceShow = !url.isBlob(blobURL);

	showExplorerTreeIfNecessary(forceShow);
	document.querySelector("#file-explorer")!.addEventListener("click", () => {
		handleToggleExplorerTree();
	});
	backend.localStoreListAllFiles(uri!, pageVars.CommitID).then(resp => {
		const el = <div className={style(vertical)}>
			<TreeHeader className={style(content)} title="Files" onDismiss={() => handleToggleExplorerTree()} />
			<Tree initSelectedPath={path} onSelectFile={(selectedPath) => window.location.href = url.toBlob({ uri: uri, rev: rev, path: selectedPath })} className={style(flex)} paths={resp.map(res => res.name)} />
		</div>;
		render(el, mount);
	});
}

function handleToggleExplorerTree(): void {
	// TODO(slimsag): add eventLogger calls
	//eventLogger.logFileTreeToggleClicked({toggled: toggled});
	const isShown = window.localStorage.getItem("show-explorer") === "true";
	window.localStorage.setItem("show-explorer", isShown ? "false" : "true");
	const treeViewer = document.querySelector("#tree-viewer")! as HTMLElement;
	treeViewer.style.display = isShown ? "none" : "flex";
}

function showExplorerTreeIfNecessary(force: boolean): void {
	// TODO(slimsag): add eventLogger calls
	//eventLogger.logFileTreeToggleClicked({toggled: toggled});
	const shouldShow = force || window.localStorage.getItem("show-explorer") === "true";
	const treeViewer = document.querySelector("#tree-viewer")! as HTMLElement;
	treeViewer.style.display = shouldShow ? "flex" : "none";
}

function registerListeners(): void {
	const openOnGitHub = document.querySelector(".github")!;
	if (openOnGitHub) {
		openOnGitHub.addEventListener("click", () => events.OpenInCodeHostClicked.log());
	}
	const openOnDesktop = document.querySelector(".open-on-desktop")!;
	if (openOnDesktop) {
		openOnDesktop.addEventListener("click", () => events.OpenInNativeAppClicked.log());
	}
}

function highlightLine(repoURI: string, commitID: string, path: string, line: number, cells: CodeCell[]): void {
	triggerBlame({
		time: moment(),
		repoURI: repoURI,
		commitID: commitID,
		path: path,
		line: line,
	});

	const currentlyHighlighted = document.querySelectorAll(".sg-highlighted");
	Array.from(currentlyHighlighted).forEach((cellElem: HTMLElement) => {
		cellElem.classList.remove("sg-highlighted");
		cellElem.style.backgroundColor = "inherit";
	});

	const cell = cells[line - 1];
	cell.cell.style.backgroundColor = "#1c2736";
	cell.cell.classList.add("sg-highlighted");

	// Update the URL.
	const u = url.parseBlob();
	u.line = line;

	if (url.toBlob(u) !== (window.location.pathname + window.location.hash)) {
		// Prevent duplicating history state for the same line.
		window.history.pushState(null, "", url.toBlobHash(u));
	}
}

function highlightAndScrollToLine(repoURI: string, commitID: string, path: string, line: number, cells: CodeCell[]): void {
	highlightLine(repoURI, commitID, path, line, cells);

	// Scroll to the line.
	const scrollingElement = document.querySelector("#blob-table")!;
	const viewportBound = scrollingElement.getBoundingClientRect();
	const blobTable = document.querySelector("#blob-table>table")!; // table that we're positioning tooltips relative to.
	const tableBound = blobTable.getBoundingClientRect(); // tables bounds
	const cell = cells[line - 1];
	const targetBound = cell.cell.getBoundingClientRect(); // our target elements bounds

	scrollingElement.scrollTop = targetBound.top - tableBound.top - (viewportBound.height / 2) + (targetBound.height / 2);
}

function highlightAsync(path: string, textContent: string): void {
	const worker = new Worker(`${(window as any).context.assetsRoot}/scripts/highlighter.bundle.js`);
	const lang = getModeFromExtension(getPathExtension(path));
	worker.onmessage = (event) => {
		const blob = document.querySelector("#blob") as HTMLElement;
		blob.innerHTML = event.data.innerHTML;
		syntaxhighlight.processBlock(blob);
	};
	worker.postMessage({ textContent, lang });
}

function getCodeCellsForAnnotation(): CodeCell[] {
	const table = document.querySelector("table") as HTMLTableElement;
	const cells: CodeCell[] = [];
	for (let i = 0; i < table.rows.length; ++i) {
		const row = table.rows[i];

		const line = parseInt(row.cells[0].getAttribute("data-line-number")!, 10);
		const codeCell: HTMLTableDataCellElement = row.cells[1]; // the actual cell that has code inside; each row contains multiple columns
		cells.push({
			cell: codeCell as HTMLElement,
			eventHandler: codeCell, // allways the TD element
			line,
		});
	}

	return cells;
}
