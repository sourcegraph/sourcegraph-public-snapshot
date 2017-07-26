import { injectReferencesWidget } from "app/references/inject";
import { addAnnotations } from "app/tooltips";
import { parseURL } from "app/util";
import { CodeCell } from "app/util/types";
import { blameFile, Hunk } from "app/backend";
import * as moment from 'moment';

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
		const pageVars = (window as any).pageVars;
		// blob view, add tooltips
		if (!pageVars || !pageVars.ResolvedRev) {
			throw new TypeError("expected window.pageVars to exist, but it does not");
		}
		const rev = pageVars.ResolvedRev;
		const cells = getCodeCellsForAnnotation();
		addAnnotations(url.path, { repoURI: url.uri, rev: rev, isBase: false, isDelta: false }, cells);
		if (line) {
			highlightAndScrollToLine(url.uri, rev, url.path, line, cells);
		}

		// Add click handlers to all lines of code, which highlight and add
		// blame information to the line.
		document.querySelectorAll(".blobview tr").forEach((tr: HTMLElement, index: number) => {
			tr.addEventListener("click", () => {
				highlightLine(url.uri, rev, url.path, index + 1, cells);
			});
		});
	}
});

function limitString(s: string, n: number, dotdotdot: boolean): string {
	if (s.length > n) {
		if (dotdotdot) {
			return s.substring(0, n - 1) + '…';
		}
		return s.substring(0, n);
	}
	return s;
}

let blameRequest = 0;

function blameLine(repoURI: string, rev: string, path: string, line: number, cells: CodeCell[]): void {
	// Clear the blame content on whatever line it was previously on.
	setLineBlameContent(-1, "", cells);

	// Keep track of which request for blame information this was. If the user
	// performs another (e.g. by clicking another line before the request
	// promise resolves), the we are no-op.
	//
	// TODO(slimsag): ideally the user triggering a new request would actually
	// cancel the outbound HTTP requests, but this is good enough for now.
	blameRequest++;
	const req = blameRequest;

	// If the promise doesn't resolve in 250ms then display a loading text on
	// the line.
	var resolved = false;
	setTimeout(() => {
		const cancelled = blameRequest != req;
		if (!resolved && !cancelled) {
			setLineBlameContent(line, "loading ◌", cells);
		}
	}, 250);

	blameFile(repoURI, rev, path, line, line).then((hunks: Hunk[]) => {
		const cancelled = blameRequest != req;
		if (cancelled) {
			return;
		}
		resolved = true;
		if (!hunks) {
			return;
		}
		const timeSince = moment(hunks[0].author.date).fromNow();
		const blameContent = `${hunks[0].author.person.name}, ${timeSince} • ${limitString(hunks[0].message, 80, true)} ${limitString(hunks[0].rev, 6, false)}`;

		setLineBlameContent(line, blameContent, cells);
	});
}

function setLineBlameContent(line: number, blameContent: string, cells: CodeCell[]): void {
	// Remove blame class from all other lines.
	const currentlyBlamed = document.querySelectorAll(".code-cell>.blame");
	currentlyBlamed.forEach((blame: HTMLElement) => {
		blame.parentNode.removeChild(blame);
	});

	if (line > 0) {
		// Add blame class to the target line.
		const cell = cells[line - 1];
		const blame = document.createElement("span");
		blame.classList.add("blame");
		blame.appendChild(document.createTextNode(blameContent));
		cell.cell.appendChild(blame);
	}
}

function highlightLine(repoURI: string, rev: string, path: string, line: number, cells: CodeCell[]): void {
	blameLine(repoURI, rev, path, line, cells);

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
	const cell = cells[line - 1];
	const element = cell.cell;
	const elementRect = element.getBoundingClientRect();
	const absoluteElementTop = elementRect.top + window.pageYOffset;
	const middle = absoluteElementTop - (window.innerHeight / 2);
	window.scrollTo(0, middle);
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
			if (!window.pageVars || !(window.pageVars as any).ResolvedRev) {
				throw new TypeError("expected window.pageVars to exist, but it does not");
			}
			const rev = (window.pageVars as any).ResolvedRev;
			const url = parseURL();
			const cells = getCodeCellsForAnnotation();
			highlightAndScrollToLine(url.uri, rev, url.path, line, cells);
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
