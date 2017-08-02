import * as colors from "app/util/colors";
import * as _ from "lodash";

/**
 * Inserts an element after the reference node.
 * @param el The element to be rendered.
 * @param referenceNode The node to render the element after.
 */
export function insertAfter(el: HTMLElement, referenceNode: Node): void {
	if (referenceNode.parentNode) {
		referenceNode.parentNode.insertBefore(el, referenceNode.nextSibling);
	}
}

export function isMouseEventWithModifierKey(e: MouseEvent): boolean {
	return e.altKey || e.shiftKey || e.ctrlKey || e.metaKey || e.which === 2;
}

export function highlightNode(parentNode: HTMLElement, start: number, end: number): void {
	highlightNodeHelper(parentNode, 0, start, end);
}

function highlightNodeHelper(parentNode: HTMLElement, curr: number, start: number, length: number): { done: boolean, consumed: number } {
	let origCurr = curr;
	let numParentNodes = parentNode.childNodes.length;
	for (let i = 0; i < numParentNodes; ++i) {
		if (curr >= start + length) {
			return { done: true, consumed: 0 };
		}
		const isLastNode = i === parentNode.childNodes.length - 1;
		const node = parentNode.childNodes[i];
		if (node.nodeType === Node.TEXT_NODE) {
			let nodeText = _.unescape(node.textContent || "");


			const containerNode = document.createElement("span");

			if (curr <= start && curr + nodeText.length > start) {
				parentNode.removeChild(node);
				const rest = nodeText.substr(start - curr);
				if (nodeText.substr(0, start - curr) !== "") {
					containerNode.appendChild(document.createTextNode(nodeText.substr(0, start - curr)));
				}

				if (rest.length >= length) {
					const text = rest.substr(0, length);
					const highlight = document.createElement("span");
					highlight.className = "selection-highlight";
					highlight.style.backgroundColor = colors.selectionHighlight;
					highlight.appendChild(document.createTextNode(text));
					containerNode.appendChild(highlight);
					if (rest.substr(length)) {
						containerNode.appendChild(document.createTextNode(rest.substr(length)));
					}

					if (parentNode.childNodes.length === 0 || isLastNode) {
						parentNode.appendChild(containerNode);
					} else {
						parentNode.insertBefore(containerNode, parentNode.childNodes[i] || parentNode.firstChild);
					}

					return { done: true, consumed: nodeText.length };
				} else {
					console.error("we shouldn't be here...", nodeText);
				}
			}

			curr += nodeText.length;
		} else if (node.nodeType === Node.ELEMENT_NODE) {
			const elementNode = node as HTMLElement;
			if (elementNode.classList.contains("selection-highlight")) {
				return { done: true, consumed: 0 };
			}
			const res = highlightNodeHelper(elementNode, curr, start, length);
			if (res.done) {
				return res;
			} else {
				curr += res.consumed;
			}
		}
	}
	return { done: false, consumed: curr - origCurr };
}
