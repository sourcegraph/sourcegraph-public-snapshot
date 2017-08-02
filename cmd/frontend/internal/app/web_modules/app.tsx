import * as xhr from "app/backend/xhr";
import { triggerBlame } from "app/blame";
import { injectReferencesWidget } from "app/references/inject";
import { injectSearchForm, injectSearchInputHandler, injectSearchResults } from "app/search/inject";
import { injectShareWidget } from "app/share";
import { addAnnotations } from "app/tooltips";
import { CodeCell } from "app/util/types";
import * as url from "app/util/url";
import * as moment from "moment";

window.addEventListener("DOMContentLoaded", () => {
	const context = (window as any).context;
	xhr.useAccessToken(context.accessToken);

	if (window.location.pathname === "/") {
		injectSearchForm();
	} else {
		injectSearchInputHandler();
	}
	if (window.location.pathname === "/search") {
		injectSearchResults();
	}

	injectReferencesWidget();
	injectShareWidget();
	const u = url.parseBlob();
	if (u.uri && u.path) {
		// blob view, add tooltips
		const pageVars = (window as any).pageVars;
		if (!pageVars || !pageVars.ResolvedRev) {
			throw new TypeError("expected window.pageVars to exist, but it does not");
		}
		const rev = pageVars.ResolvedRev;
		const cells = getCodeCellsForAnnotation();
		window.addEventListener("syntaxHighlightingFinished", () => {
			addAnnotations(u.path!, { repoURI: u.uri!, rev: rev, isBase: false, isDelta: false }, cells);
		});
		if (u.line) {
			highlightAndScrollToLine(u.uri, rev, u.path, u.line, cells);
		}

		// Add click handlers to all lines of code, which highlight and add
		// blame information to the line.
		Array.from(document.querySelectorAll(".blobview tr")).forEach((tr: HTMLElement, index: number) => {
			tr.addEventListener("click", () => {
				if (u.uri && u.path) {
					highlightLine(u.uri, rev, u.path, index + 1, cells);
				}
			});
		});
	}
});

function highlightLine(repoURI: string, rev: string, path: string, line: number, cells: CodeCell[]): void {
	triggerBlame({
		time: moment(),
		repoURI: repoURI,
		rev: rev,
		path: path,
		line: line,
	});

	const currentlyHighlighted = document.querySelectorAll(".sg-highlighted");
	Array.from(currentlyHighlighted).forEach((cell: HTMLElement) => {
		cell.classList.remove("sg-highlighted");
		cell.style.backgroundColor = "inherit";
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
	const scrollable = document.querySelector("#blob-table")!;
	const scrollableRect = scrollable.getBoundingClientRect(); // e.x. the navbar height

	const cell = cells[line - 1];
	const cellRect = cell.cell.getBoundingClientRect(); // e.x. distance from top of code cell to top of table
	scrollable.scrollTop = (cellRect.top + (cellRect.height / 2)) - (scrollableRect.top + (scrollableRect.height / 2));
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
	const pageVars = (window as any).pageVars;
	if (!pageVars || !pageVars.ResolvedRev) {
		throw new TypeError("expected window.pageVars to exist, but it does not");
	}
	const rev = pageVars.ResolvedRev;
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
