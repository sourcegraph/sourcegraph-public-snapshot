/**
 * TODO: This module should be rewritten! We shouldn't be using DOM events
 * anymore since we're not communicating across separate JS <-> TS scripts.
 */

/**
 * done marks syntax highlighting as done.
 */
export function done(): void {
	const finishEvent = document.createEvent("Event");
	finishEvent.initEvent("syntaxHighlightingFinished", true, true);
	window.dispatchEvent(finishEvent);
}

let syntaxHighlightingFinished = false;

window.addEventListener("syntaxHighlightingFinished", () => {
	syntaxHighlightingFinished = true;
}, false);

/**
 * wait returns a promise that waits for syntax highlighting to be finished.
 */
export function wait(): Promise<void> {
	if (syntaxHighlightingFinished) {
		return Promise.resolve();
	}
	return new Promise((resolve, _reject) => {
		window.addEventListener("syntaxHighlightingFinished", () => {
			resolve();
		}, false);
	});
}

export function processBlock(el: HTMLElement): void {
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

	for (const node of Array.from(el.childNodes)) {
		nodeProcessor(node);
	}

	const table = document.createElement("table");
	table.id = "processed-blob";
	const body = document.createElement("tbody");

	lines.forEach((l, i) => {
		const row = document.createElement("tr");
		const line = document.createElement("td");
		line.classList.add("line-number");
		line.setAttribute("data-line-number", "" + (i + 1));

		const cell = document.createElement("td");
		cell.classList.add("code-cell");
		l.forEach(node => cell.appendChild(node));

		row.appendChild(line);
		row.appendChild(cell);

		body.appendChild(row);
	});

	table.appendChild(body);

	document.querySelector("#blob-table")!.appendChild(table);

	done(); // mark syntax highlighting as finished
}
