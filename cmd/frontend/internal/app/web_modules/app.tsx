import { injectReferencesWidget } from "app/references/inject";
import { addAnnotations } from "app/tooltips";
import { parseURL } from "app/util";
import { CodeCell } from "app/util/types";

window.addEventListener("DOMContentLoaded", () => {
	injectReferencesWidget();
	const url = parseURL();
	const hash = window.location.hash;
	let line;
	if (hash) {
		const split = hash.split("#L");
		if (split[1]) {
			line = parseInt(split[1].split(":")[0], 10)
		}
	}
	if (url.uri && url.path) {
		// blob view, add tooltips
		const rev = (window.pageVars as any).ResolvedRev;
		const cells = getCodeCellsForAnnotation();
		addAnnotations(url.path, { repoURI: url.uri, rev: rev, isBase: false, isDelta: false }, cells);
		if (line) {
			highlightAndScrollToLine(line, cells);
		}
	}
});

export function highlightAndScrollToLine(line: number, cells: CodeCell[]): void {
	const currentlyHighlighted = document.querySelectorAll(".sg-highlighted");
	Array.from(currentlyHighlighted).forEach((cell: HTMLElement) => {
		cell.classList.remove("sg-highlighted");
		cell.style.backgroundColor = "inherit";
	});

	const cell = cells[line - 1];
	cell.cell.style.backgroundColor = "rgba(255,255,0,.25)";
	cell.cell.classList.add("sg-highlighted");
	const element = cell.cell;
	const elementRect = element.getBoundingClientRect();
	const absoluteElementTop = elementRect.top + window.pageYOffset;
	const middle = absoluteElementTop - (window.innerHeight / 2);
	window.scrollTo(0, middle);
}

window.onhashchange = (hash) => {
	const split = hash.newURL!.split("#L");
	if (split[1]) {
		const line = parseInt(split[1].split(":")[0], 10)
		const cells = getCodeCellsForAnnotation();
		highlightAndScrollToLine(line, cells);
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
