import * as _ from "lodash";
import { eventLogger, isBrowserExtension } from "../utils/context";
import * as github from "./github";
import { fetchJumpURL, getTooltip, prewarmLSP } from "./lsp";
import * as tooltips from "./tooltips";
import { CodeCell, PhabricatorCodeCell } from "./types";

export interface RepoRevSpec {
	repoURI: string;
	rev: string;
	isDelta: boolean;
	isBase: boolean;
}
/**
 * addAnnotations is the entry point for injecting annotations onto a blob (el).
 * An invisible marker is appended to the document to indicate that annotation
 * has been completed; so this function expects that it will be called once all
 * repo/annotation data is resolved from the server.
 *
 * el should be an element that changes when the dom significantly changes.
 * datakeys are stored as properites on el, and the code shortcuts if the datakey
 * is detected.
 *
 * spacesToTab, if nonzero, converts leading whitespace on each line to be converted
 * into tabs (some repository hosts like Bitbucket Server and Phabricator auto-convert
 * tabs into spaces in their code browrsers). The spacesToTab mechanism of reversing
 * this conversion works well enough for now, but is flawed. For BBS, a better
 * mechanism would be to take use the `cm-tab` DOM attribute. For Phabricator, no
 * better mechanism is known at this time (see https://secure.phabricator.com/T2495).
 */
export function addAnnotations(path: string, repoRevSpec: RepoRevSpec, el: HTMLElement, loggingStruct: Object, cells: CodeCell[], spacesToTab: number): void {
	cells.forEach((cell) => {
		let annotated = false;
		// parse, annotate and replace the node asynchronously when the user hovers over the code cell
		const annotateCb = () => {
			if (annotated) {
				return;
			}
			annotated = true;

			try {
				let ignoreFirstTextChar = repoRevSpec.isDelta;

				if ((cell as PhabricatorCodeCell).isLeftColumnInSplit || (cell as PhabricatorCodeCell).isUnified) {
					ignoreFirstTextChar = false;
				}

				const annLine = convertNode(cell.cell, 1, cell.line, ignoreFirstTextChar, spacesToTab);
				cell.cell.innerHTML = "";
				cell.cell.appendChild(annLine.resultNode);

				// addEventListeners(cell.cell, path, repoRevSpec, cell.line, loggingStruct);
			} catch (e) {
				console.error(e);
			}
		};

		const dataKey = `data-${cell.line}-${repoRevSpec.rev}`;
		// If the line has already been annotated,
		// restore event handlers if necessary otherwise move to next line
		// the first check works on GitHub, the second is required for phabricator
		// but is a no-op for GitHub
		// TODO(uforic):  && hasCellBeenAnnotated(cell.cell) - figure out why we need this.
		if (el.getAttribute(dataKey)) {
			if (!el.onclick || !el.onmouseout || !el.onmouseover) {
				addEventListeners(cell.cell, annotateCb, path, repoRevSpec, cell.line, loggingStruct);
			}
			return;
		}

		addEventListeners(cell.cell, annotateCb, path, repoRevSpec, cell.line, loggingStruct);
	});
}

interface ConvertNodeResult<T extends Node> {
	resultNode: T;
	bytesConsumed: number;

	// keeps track of whether the Node converted was all spaces. We do this, because we only want to treat leading
	// whitespace as candidates for conversion back to tabs (this is a heuristic).
	isAllSpaces: boolean;
}

function convertNodeHelper(node: Node, offset: number, line: number, ignoreFirstTextChar: boolean, spacesToTab: number): ConvertNodeResult<Element> {
	switch (node.nodeType) {
		case Node.TEXT_NODE:
			return convertTextNode(node, offset, line, ignoreFirstTextChar, spacesToTab);

		case Node.ELEMENT_NODE:
			return convertElementNode(node, offset, line, ignoreFirstTextChar, spacesToTab);

		default:
			throw new Error(`unexpected node type(${node.nodeType})`);
	}
}

// convertNode takes a DOM node and returns an object containing the
// maybe-linkified version of the node as an HTML string as well as the number of bytes consumed.
// It is the entry point for converting a <td> "cell" representing a line of code.
// It may also be called recursively for children (which are assumed to be <span>
// code highlighting annotations from GitHub).
export function convertNode(currentNode: Node, offset: number, line: number, ignoreFirstTextChar: boolean, spacesToTab: number): ConvertNodeResult<Node> {
	let wrapperNode;
	let c = convertNodeHelper(currentNode, offset, line, ignoreFirstTextChar, spacesToTab);

	// If this is the top level node for code, return a documentFragment
	// otherwise copy all the attributes of the original node.
	if ((currentNode as any).tagName === "TD") {
		wrapperNode = document.createDocumentFragment();
		wrapperNode.appendChild(c.resultNode);
	} else {
		wrapperNode = c.resultNode;
		if (currentNode.attributes && currentNode.attributes.length > 0) {
			[].map.call(currentNode.attributes, (attr) => wrapperNode.setAttribute(attr.name, attr.value));
		}
	}

	return {
		resultNode: wrapperNode,
		bytesConsumed: c.bytesConsumed,
		isAllSpaces: c.isAllSpaces,
	};
}

// convertTextNode takes a DOM node which should be of NodeType.TEXT_NODE
// (this must be checked by the caller) and returns an object containing the
//  maybe-linkified version of the node as an HTML string and the number
// of bytes consumed.
const VARIABLE_TOKENIZER = /(^\w+)/;
const ASCII_CHARACTER_TOKENIZER = /(^[\x21-\x2F|\x3A-\x40|\x5B-\x60|\x7B-\x7E])/;
const NONVARIABLE_TOKENIZER = /(^[^\x21-\x7E]+)/;
const SPACES = /(^[\x20]*$)/;
export function convertTextNode(currentNode: Node, offset: number, line: number, ignoreFirstTextChar: boolean, spacesToTab: number): ConvertNodeResult<Element> {
	let nodeText;
	let prevConsumed = 0;
	let bytesConsumed = 0;
	const wrapperNode = document.createElement("SPAN");
	wrapperNode.id = `text-node-wrapper-${line}-${offset}`;

	function createTextNode(text: string, off: number): Node {
		const wrapNode = document.createElement("SPAN");
		wrapNode.id = `text-node-${line}-${off}`;
		const textNode = document.createTextNode(text);

		wrapNode.setAttribute("data-byteoffset", `${off}`);
		wrapNode.appendChild(textNode);

		return wrapNode;
	}

	// Text could contain escaped character sequences (e.g. "&gt;")
	nodeText = _.unescape(currentNode.textContent || "");

	// Handle special case for pull requests (+/- character on diffs).
	if (ignoreFirstTextChar && nodeText.length > 0) {
		wrapperNode.appendChild(document.createTextNode(nodeText[0]));
		nodeText = nodeText.slice(1);
	}

	function consumeNext(txt: string): string {
		// first, check for real stuff, i.e. sets of [A-Za-z0-9_]
		const variableMatch = txt.match(VARIABLE_TOKENIZER);
		if (variableMatch) {
			return variableMatch[0];
		}
		// next, check for tokens that are not variables, but should stand alone
		// i.e. {}, (), :;. ...
		const asciiMatch = txt.match(ASCII_CHARACTER_TOKENIZER);
		if (asciiMatch) {
			return asciiMatch[0];
		}
		// finally, the remaining tokens we can combine into blocks, since they are whitespace
		// or UTF8 control characters. We had better clump these in case UTF8 control bytes
		// require adjacent bytes
		const nonVariableMatch = txt.match(NONVARIABLE_TOKENIZER);
		if (nonVariableMatch) {
			return nonVariableMatch[0];
		}
		return txt[0];
	}

	var allSpaces = true;
	while (nodeText.length > 0) {
		const token = consumeNext(nodeText);
		const isAllSpaces = SPACES.test(token);
		allSpaces = isAllSpaces && allSpaces;

		wrapperNode.appendChild(createTextNode(token, offset + prevConsumed));
		prevConsumed += isAllSpaces && spacesToTab > 0 && token.length % spacesToTab === 0 ? token.length / spacesToTab : token.length;
		bytesConsumed += isAllSpaces && spacesToTab > 0 && token.length % spacesToTab === 0 ? token.length / spacesToTab : token.length;
		if (!allSpaces && spacesToTab > 0) {
			// NOTE: this makes it so that if there are further spaces, they don't get divided by 2 for their byte offset.
			// only divide by 2 for initial code indents.
			spacesToTab = 0;
		}
		nodeText = nodeText.slice(token.length);
	}

	return { resultNode: wrapperNode, bytesConsumed, isAllSpaces: allSpaces };
}

// convertElementNode takes a DOM node which should be of NodeType.ELEMENT_NODE
// (this must be checked by the caller) and returns an object containing the
//  maybe-linkified version of the node as an HTML string as well as the number of bytes consumed.
export function convertElementNode(currentNode: Node, offset: number, line: number, ignoreFirstTextChar: boolean, spacesToTab: number): ConvertNodeResult<Element> {
	let bytesConsumed = 0;
	const wrapperNode = document.createElement("SPAN");

	wrapperNode.setAttribute("data-byteoffset", `${offset}`);

	// The logic here is to simply recurse on each of the child nodes; everything is eventually
	// just a text node or the special-cased "quoted string node" (see below).
	var isAllSpaces = true;
	for (let i = 0; i < currentNode.childNodes.length; ++i) {
		const res = convertNode(currentNode.childNodes[i], offset + bytesConsumed, line, i === 0 && ignoreFirstTextChar, spacesToTab);
		isAllSpaces = isAllSpaces && res.isAllSpaces;
		// NOTE: this makes it so that if there are further spaces, they don't get divided by 2 for their byte offset.
		// only divide by 2 for initial code indents.
		if (!isAllSpaces && spacesToTab > 0) {
			spacesToTab = 0;
		}
		wrapperNode.appendChild(res.resultNode);
		bytesConsumed += res.bytesConsumed;
	}

	return { resultNode: wrapperNode, bytesConsumed, isAllSpaces: isAllSpaces };
}

let activeTarget;
function getTarget(t: HTMLElement): HTMLElement | undefined {
	if (t.tagName === "TD") {
		// Not hovering over any token in particular.
		return;
	}
	while (t && t.tagName !== "TD" && !t.getAttribute("data-byteoffset")) {
		t = (t.parentNode as HTMLElement);
	}

	if (t && t.tagName === "SPAN" && t.getAttribute("data-byteoffset")) {
		return t;
	}
}

function addEventListeners(el: HTMLElement, annotateCb: () => void, path: string, repoRevSpec: RepoRevSpec, line: number, loggingStruct: Object): void {
	tooltips.createTooltips();

	el.onclick = e => {
		let t = getTarget(e.target as HTMLElement);
		if (!t || t.style.cursor !== "pointer") {
			return;
		}

		fetchJumpURL(t.dataset["byteoffset"], path, line, repoRevSpec).then((defUrl) => {
			if (!defUrl) {
				return;
			}

			// If cmd/ctrl+clicked or middle button clicked, open in new tab/page otherwise
			// either move to a line on the same page, or refresh the page to a new blob view.
			eventLogger.logJumpToDef(Object.assign({}, repoRevSpec, loggingStruct));
			window.open(defUrl, "_blank");
		});
	};

	el.onmouseout = (e) => {
		tooltips.clearContext();
		activeTarget = null;
	};

	el.onmouseover = (e) => {
		annotateCb(); // annotate the cell

		let t = getTarget(e.target as HTMLElement);
		if (!t || activeTarget === t) {
			// don't do anything unless target is defined and has changed
			return;
		}

		activeTarget = t;
		tooltips.setContext(t, loggingStruct);
		tooltips.queueLoading();

		getTooltip(activeTarget, path, line, repoRevSpec).then((data) => tooltips.setTooltip(data, activeTarget)).catch((err) => tooltips.setTooltip(null, activeTarget));
	};
}
