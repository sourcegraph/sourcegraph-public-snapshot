import { injectReferencesWidget } from "app/references/inject";
import { injectSearchForm, injectSearchResults } from "app/search/inject";
import { addAnnotations } from "app/tooltips";
import * as url from "app/util/url";
import { CodeCell } from "app/util/types";
import { triggerBlame } from "app/blame";
import * as moment from 'moment';
import { injectShareWidget } from "app/share";

window.addEventListener("DOMContentLoaded", () => {
	if (window.location.pathname === "/") {
		injectSearchForm();
	}
	if (window.location.pathname === "/search") {
		injectSearchResults();
	}

	injectReferencesWidget();
	injectShareWidget();
	const u = url.parseBlob();
	const hash = window.location.hash;
	let line;
	if (hash) {
		const split = hash.split("#L");
		if (split[1]) {
			line = parseInt(split[1].split(":")[0], 10)
		}
	}
	if (u.uri && u.path) {
		// blob view, add tooltips
		const pageVars = (window as any).pageVars;
		if (!pageVars || !pageVars.ResolvedRev) {
			throw new TypeError("expected window.pageVars to exist, but it does not");
		}
		const rev = pageVars.ResolvedRev;
		const cells = getCodeCellsForAnnotation();
		window.addEventListener("syntaxHighlightingFinished", () => {
			addAnnotations(u.path, { repoURI: u.uri, rev: rev, isBase: false, isDelta: false }, cells);
		});
		if (line) {
			highlightAndScrollToLine(u.uri, rev, u.path, line, cells);
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
	const oldSplit = hash.oldURL!.split("#L");
	let lastLine;
	if (oldSplit[1]) {
		lastLine = parseInt(oldSplit[1].split(":")[0], 10);
	}
	const newSplit = hash.newURL!.split("#L");
	if (newSplit[1]) {
		const line = parseInt(newSplit[1].split(":")[0], 10);
		if (lastLine !== line) {
			// prevent e.g. re-scrolling to same line on toggling refs group
			const pageVars = (window as any).pageVars;
			if (!pageVars || !pageVars.ResolvedRev) {
				throw new TypeError("expected window.pageVars to exist, but it does not");
			}
			const rev = pageVars.ResolvedRev;
			const u = url.parseBlob();
			const cells = getCodeCellsForAnnotation();
			if (u.uri && u.path) {
				highlightAndScrollToLine(u.uri, rev, u.path, line, cells);
			}
		}
	}
}

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
