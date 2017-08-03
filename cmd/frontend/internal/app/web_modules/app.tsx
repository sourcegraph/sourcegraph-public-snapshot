import { highlightBlock } from "highlight.js";
import * as moment from "moment";
import * as xhr from "sourcegraph/backend/xhr";
import { triggerBlame } from "sourcegraph/blame";
import { injectReferencesWidget } from "sourcegraph/references/inject";
import { injectAdvancedSearchDrawer, injectAdvancedSearchToggle, injectSearchForm, injectSearchInputHandler, injectSearchResults } from "sourcegraph/search/inject";
import { injectShareWidget } from "sourcegraph/share";
import { addAnnotations } from "sourcegraph/tooltips";
import { handleQueryEvents } from "sourcegraph/tracking/analyticsUtils";
import { events, viewEvents } from "sourcegraph/tracking/events";
import { getModeFromExtension, getPathExtension } from "sourcegraph/util";
import * as activeRepos from "sourcegraph/util/activeRepos";
import { pageVars } from "sourcegraph/util/pageVars";
import { sourcegraphContext } from "sourcegraph/util/sourcegraphContext";
import * as syntaxhighlight from "sourcegraph/util/syntaxhighlight";
import { CodeCell } from "sourcegraph/util/types";
import * as url from "sourcegraph/util/url";

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

	injectReferencesWidget();
	injectShareWidget();
	const u = url.parseBlob();
	if (u.uri && u.path) {

		const blob = document.querySelector("#blob")!;
		blob.className = getModeFromExtension(getPathExtension(u.path));
		highlightBlock(document.querySelector("#blob"));

		const lines: Node[][] = [[]];

		const nodeProcessor = (node: Node, wrapperClass?: string) => {
			const wrap = (n: Node, className: string): any => {
				const wrapper = document.createElement("span");
				wrapper.className = className;
				wrapper.appendChild(n);
				return wrapper;
			};

			if (node.nodeType === Node.TEXT_NODE) {
				const text = node.textContent!;
				if (text.indexOf("\n") !== -1) {
					const split = text.split("\n");
					split.forEach((val, i) => {
						if (i !== 0) {
							lines.push([]);
						}
						let newNode = document.createTextNode(val);
						if (wrapperClass) {
							newNode = wrap(newNode, wrapperClass);
						}
						lines[lines.length - 1].push(newNode);
					});
				} else {
					if (wrapperClass) {
						node = wrap(node, wrapperClass);
					}
					lines[lines.length - 1].push(node);
				}
			} else {
				if (node.childNodes.length === 1) {
					const className = (node as HTMLElement).className;
					const text = node.textContent!;
					if (text.indexOf("\n") !== -1) {
						const split = text.split("\n");
						split.forEach((val, i) => {
							if (i !== 0) {
								lines.push([]);
							}
							let newNode = wrap(document.createTextNode(val), className);
							if (wrapperClass) {
								newNode = wrap(newNode, wrapperClass);
							}
							lines[lines.length - 1].push(newNode);
						});
					} else {
						let newNode = wrap(document.createTextNode(text), className);
						if (wrapperClass) {
							newNode = wrap(newNode, wrapperClass);
						}
						lines[lines.length - 1].push(newNode);
					}
				} else {
					if (node.textContent!.indexOf("\n") !== -1) {
						const className = (node as HTMLElement).className;
						for (const n of Array.from(node.childNodes)) {
							nodeProcessor(n, className);
						}
					} else {
						lines[lines.length - 1].push(node);
					}
				}
			}
		};
		for (const node of Array.from(document.querySelector("#blob")!.childNodes)) {
			nodeProcessor(node);
		}

		const table = document.createElement("table");
		const body = document.createElement("tbody");

		lines.forEach((l, i) => {
			const row = document.createElement("tr");
			const line = document.createElement("td");
			line.classList.add("line-number");
			line.appendChild(document.createTextNode("" + (i + 1)));

			const cell = document.createElement("td");
			cell.classList.add("code-cell");
			l.forEach(node => cell.appendChild(node));

			row.appendChild(line);
			row.appendChild(cell);

			body.appendChild(row);
		});

		table.appendChild(body);

		document.querySelector("#blob-table")!.appendChild(table);

		syntaxhighlight.done(); // mark syntax highlighting as finished

		// blob view, add tooltips
		const commitID = pageVars.CommitID;
		const cells = getCodeCellsForAnnotation();
		addAnnotations(u.path!, { repoURI: u.uri!, rev: commitID }, cells);
		if (u.line) {
			highlightAndScrollToLine(u.uri, commitID, u.path, u.line, cells);
		}

		// Log blog view
		viewEvents.Blob.log({ repo: u.uri!, commitID, path: u.path!, language: getPathExtension(u.path) });

		// Add click handlers to all lines of code, which highlight and add
		// blame information to the line.
		Array.from(document.querySelectorAll(".blobview tr")).forEach((tr: HTMLElement, index: number) => {
			tr.addEventListener("click", () => {
				if (u.uri && u.path) {
					highlightLine(u.uri, commitID, u.path, index + 1, cells);
				}
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

function highlightLine(repoURI: string, rev: string, path: string, line: number, cells: CodeCell[]): void {
	triggerBlame({
		time: moment(),
		repoURI: repoURI,
		rev: rev,
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

export function highlightAndScrollToLine(repoURI: string, rev: string, path: string, line: number, cells: CodeCell[]): void {
	highlightLine(repoURI, rev, path, line, cells);

	// Scroll to the line.
	const scrollingElement = document.querySelector("#blob-table")!;
	const viewportBound = scrollingElement.getBoundingClientRect();
	const blobTable = document.querySelector("#blob-table>table")!; // table that we're positioning tooltips relative to.
	const tableBound = blobTable.getBoundingClientRect(); // tables bounds
	const cell = cells[line - 1];
	const targetBound = cell.cell.getBoundingClientRect(); // our target elements bounds

	scrollingElement.scrollTop = targetBound.top - tableBound.top - (viewportBound.height / 2) + (targetBound.height / 2);
}

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
	const rev = pageVars.CommitID;
	const cells = getCodeCellsForAnnotation();
	highlightAndScrollToLine(newURL.uri!, rev, newURL.path, newURL.line, cells);
};

export function getCodeCellsForAnnotation(): CodeCell[] {
	const table = document.querySelector("table") as HTMLTableElement;
	const cells: CodeCell[] = [];
	for (let i = 0; i < table.rows.length; ++i) {
		const row = table.rows[i];

		const line = parseInt(row.cells[0].textContent!, 10);
		const codeCell: HTMLTableDataCellElement = row.cells[1]; // the actual cell that has code inside; each row contains multiple columns
		cells.push({
			cell: codeCell as HTMLElement,
			eventHandler: codeCell, // allways the TD element
			line,
		});
	}

	return cells;
}
