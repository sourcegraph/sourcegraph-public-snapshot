import { injectReferencesWidget } from "app/references/inject";
import { setReferences } from "app/references/store";
import { addAnnotations } from "app/tooltips";
import { parseURL } from "app/util";
import { CodeCell } from "app/util/types";

document.addEventListener("DOMContentLoaded", () => {
	injectReferencesWidget();
	//do work
	// setReferences({
	// 	docked: true,
	// 	context: {
	// 		path: "mux.go",
	// 		repoRevSpec: {
	// 			repoURI: "github.com/gorilla/mux",
	// 			rev: "ac112f7d75a0714af1bd86ab17749b31f7809640",
	// 			isDelta: false,
	// 			isBase: false,
	// 		},
	// 		coords: {
	// 			line: 40,
	// 			char: 26,
	// 			word: "Handler",
	// 		},
	// 	},
	// });

	const cells = getCodeCellsForAnnotation();
	const url = parseURL();
	if (url.uri && url.rev && url.path) {
		// blob view, add tooltips
		// TODO(john): this won't work for empty (e.g. default branch) rev
		addAnnotations(url.path, { repoURI: url.uri, rev: url.rev, isBase: false, isDelta: false }, cells);
	}
});

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
