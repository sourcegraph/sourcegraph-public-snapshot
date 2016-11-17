import {doFetch as fetch} from "../actions/xhr";
import {EventLogger} from "../analytics/EventLogger";
import * as types from "../constants/types";
import * as github from "./github";
import * as utils from "./index";
import * as _ from "lodash"; // TODO(john): see if we can remove lodash
import {style} from "typestyle";
import * as utf8 from "utf8";

const tooltipClassName = style({
	backgroundColor: "#2D2D30",
	maxWidth: "500px",
	maxHeight: "250px",
	overflow: "auto",
	border: "solid 1px #555",
	fontFamily: `"Helvetica Neue", Helvetica, Arial, sans-serif`,
	color: "rgba(213, 229, 242, 1)",
	fontSize: "12px",
	zIndex: 100,
	position: "absolute",
	wordWrap: "break-word",
	padding: "5px 6px",
});

const tooltipTitleStyle = style({
	fontFamily: `Menlo, Monaco, Consolas, "Courier New", monospace`,
	wordWrap: "break-all",
});

interface RepoRevSpec {
	repoURI: string;
	rev: string;
	isDelta: boolean;
	isBase: boolean;
	relRev: string;
}

// addAnnotations is the entry point for injecting annotations onto a blob (el).
// An invisible marker is appended to the document to indicate that annotation
// has been completed; so this function expects that it will be called once all
// repo/annotation data is resolved from the server.
export function addAnnotations(path: string, repoRevSpec: RepoRevSpec, el: HTMLElement, anns: types.Annotation[], lineStartBytes: number[], isSplitDiff: boolean, loggingStruct: Object): void {
	_applyAnnotations(el, path, repoRevSpec, indexAnnotations(anns).annsByStartByte, indexLineStartBytes(lineStartBytes), isSplitDiff, loggingStruct);
}

// _applyAnnotations is a helper function for addAnnotations
export function _applyAnnotations(el: HTMLElement, path: string, repoRevSpec: RepoRevSpec, annsByStartByte: AnnotationsByByte, startBytesByLine: StartBytesByLine, isSplitDiff: boolean, loggingStruct: Object): void {
	// The blob is represented by a table; the first column is the line number,
	// the second is code. Each row is a line of code
	const table = el.querySelector("table");
	if (!table) {
		return;
	}

	const cells = github.getCodeCellsForAnnotation(table, Object.assign({isSplitDiff}, repoRevSpec));
	cells.forEach((cell) => {
		const dataKey = `data-${cell.line}-${repoRevSpec.rev}`;

		// If the line has already been annotated,
		// restore event handlers if necessary otherwise move to next line
		if (el.getAttribute(dataKey)) {
			if (!el.onclick || !el.onmouseout || !el.onmouseover) {
				addEventListeners(cell.cell, path, repoRevSpec, cell.line, loggingStruct);
			}
			return;
		}
		el.setAttribute(dataKey, "true");

		// parse, annotate and replace the node asynchronously.
		setTimeout(() => {
			const annLine = convertNode(cell.cell, annsByStartByte, startBytesByLine[cell.line], startBytesByLine[cell.line], repoRevSpec.isDelta);

			cell.cell.innerHTML = "";
			cell.cell.appendChild(annLine.resultNode);

			addEventListeners(cell.cell, path, repoRevSpec, cell.line, loggingStruct);
		});
	});
}

export type AnnotationsByByte = {[key: number]: types.Annotation};

// indexAnnotations creates a fast lookup structure optimized to query
// annotations by start or end byte.
export function indexAnnotations(anns: types.Annotation[]): {annsByStartByte: AnnotationsByByte, annsByEndByte: AnnotationsByByte} {
	let annsByStartByte: AnnotationsByByte = {};
	let annsByEndByte: AnnotationsByByte = {};
	for (let i = 0; i < anns.length; i++) {
		// From pkg/syntaxhighlight/html_annotator.go
		const annType = anns[i].Class;
		if (annType !== "com" && annType !== "lit" && annType !== "pun" && annType !== "kwd" && annType !== "str") {
			let ann = anns[i];
			annsByStartByte[ann.StartByte] = ann;
			annsByEndByte[ann.EndByte] = ann;
		}
	}
	return {annsByStartByte, annsByEndByte};
}

export type StartBytesByLine = {[key: number]: number};

// indexLineStartBytes creates a fast lookup structure optimized to query
// byte offsets by line number (1-indexed).
export function indexLineStartBytes(lineStartBytes: number[]): StartBytesByLine {
	let startBytesByLine: StartBytesByLine = {};
	for (let i = 0; i < lineStartBytes.length; i++) {
		startBytesByLine[i + 1] = lineStartBytes[i];
	}
	return startBytesByLine;
}

export function isCommentNode(node: Node): boolean {
	return (node as Element).className.includes("pl-c");
}

export function isStringNode(node: Node): boolean {
	return (node as Element).className.includes("pl-s") &&
		node.childNodes.length === 3 &&
		(node.childNodes[0] as Element).className.includes("pl-pds") &&
		(node.childNodes[2] as Element).className.includes("pl-pds");
}

interface ConvertNodeResult<T extends Node> {
	resultNode: T;
	bytesConsumed: number;
}

function convertNodeHelper(node: Node, annsByStartByte: AnnotationsByByte, offset: number, lineStart: number, ignoreFirstTextChar: boolean): ConvertNodeResult<Element> {
	switch (node.nodeType) {
		case Node.TEXT_NODE:
			return convertTextNode(node, annsByStartByte, offset, lineStart, ignoreFirstTextChar);

		case Node.ELEMENT_NODE:
			return isStringNode(node) || isCommentNode(node) ?
				convertStringNode(node, annsByStartByte, offset, lineStart) :
				convertElementNode(node, annsByStartByte, offset, lineStart, ignoreFirstTextChar);

		default:
			throw new Error(`unexpected node type(${node.nodeType})`);
	}
}

// convertNode takes a DOM node and returns an object containing the
// maybe-linkified version of the node as an HTML string as well as the number of bytes consumed.
// It is the entry point for converting a <td> "cell" representing a line of code.
// It may also be called recursively for children (which are assumed to be <span>
// code highlighting annotations from GitHub).
export function convertNode(currentNode: Node, annsByStartByte: AnnotationsByByte, offset: number, lineStart: number, ignoreFirstTextChar: boolean): ConvertNodeResult<Node> {
	let wrapperNode;
	let c = convertNodeHelper(currentNode, annsByStartByte, offset, lineStart, ignoreFirstTextChar);

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
	};
}

// convertTextNode takes a DOM node which should be of NodeType.TEXT_NODE
// (this must be checked by the caller) and returns an object containing the
//  maybe-linkified version of the node as an HTML string and the number
// of bytes consumed.
export function convertTextNode(currentNode: Node, annsByStartByte: AnnotationsByByte, offset: number, lineStart: number, ignoreFirstTextChar: boolean): ConvertNodeResult<Element> {
	let nodeText;
	let prevConsumed = 0;
	let bytesConsumed = 0;
	let lineOffset = offset;
	const wrapperNode = document.createElement("SPAN");
	wrapperNode.id = `text-node-wrapper-${lineOffset}`;

	function createTextNode(text: string[], start: number, end: number, off: number): Node {
		const wrapNode = document.createElement("SPAN");
		wrapNode.id = `text-node-${lineOffset}-${off}`;
		const textNode = document.createTextNode(utf8.decode(text.slice(start, end).join("")));

		wrapNode.setAttribute("data-byteoffset", `${off}`);
		wrapNode.appendChild(textNode);

		return wrapNode;
	}

	// Text could contain escaped character sequences (e.g. "&gt;") or UTF-8
	// decoded characters (e.g. "˟") which need to be properly counted in terms of bytes.
	nodeText = utf8.encode(_.unescape(currentNode.textContent || "")).split("");

	// Handle special case for pull requests (+/- character on diffs).
	if (ignoreFirstTextChar && nodeText.length > 0) {
		wrapperNode.appendChild(document.createTextNode(utf8.decode(nodeText[0])));
		nodeText = nodeText.slice(1);
	}

	for (bytesConsumed = 0; bytesConsumed < nodeText.length; ) {
		const match = annsByStartByte[offset + bytesConsumed];

		if (match) {
			if (prevConsumed < bytesConsumed) {
				// Consume the bytes that have been passed from no matches into a single text node.
				wrapperNode.appendChild(createTextNode(nodeText, prevConsumed, bytesConsumed, offset + prevConsumed + 1 - lineStart));
				prevConsumed = bytesConsumed;
			}

			bytesConsumed += (match.EndByte - match.StartByte);
			wrapperNode.appendChild(createTextNode(nodeText, prevConsumed, bytesConsumed, offset + prevConsumed + 1 - lineStart));
			prevConsumed = bytesConsumed;
		} else {
			bytesConsumed++;
		}
	}

	if (prevConsumed < bytesConsumed) {
		wrapperNode.appendChild(createTextNode(nodeText, prevConsumed, bytesConsumed, offset + prevConsumed + 1 - lineStart));
	}

	return {resultNode: wrapperNode, bytesConsumed};
}

// convertStringNode takes a DOM node which is a plain string and returns an object containing the
// maybe-linkified version of the node as an HTML string as well as the number of bytes consumed.
export function convertStringNode(currentNode: Node, annsByStartByte: AnnotationsByByte, offset: number, lineStart: number): ConvertNodeResult<Element> {
	const wrapperNode = document.createElement("SPAN");
	wrapperNode.setAttribute("data-byteoffset", `${offset + 1 - lineStart}`);
	wrapperNode.appendChild(currentNode.cloneNode(true));

	return {
		resultNode: wrapperNode,
		bytesConsumed: (currentNode.textContent || "").length,
	};
}

// convertElementNode takes a DOM node which should be of NodeType.ELEMENT_NODE
// (this must be checked by the caller) and returns an object containing the
//  maybe-linkified version of the node as an HTML string as well as the number of bytes consumed.
export function convertElementNode(currentNode: Node, annsByStartByte: AnnotationsByByte, offset: number, lineStart: number, ignoreFirstTextChar: boolean): ConvertNodeResult<Element> {
	let bytesConsumed = 0;
	const wrapperNode = document.createElement("SPAN");

	wrapperNode.setAttribute("data-byteoffset", `${offset + 1 - lineStart}`);

	// The logic here is to simply recurse on each of the child nodes; everything is eventually
	// just a text node or the special-cased "quoted string node" (see below).
	for (let i = 0; i < currentNode.childNodes.length; ++i) {
		const res = convertNode(currentNode.childNodes[i], annsByStartByte, offset + bytesConsumed, lineStart, i === 0 && ignoreFirstTextChar);
		wrapperNode.appendChild(res.resultNode);
		bytesConsumed += res.bytesConsumed;
	}

	return {resultNode: wrapperNode, bytesConsumed};
}

// The rest of this file is responsible for fetching/caching annotation specific data from the server
// and adding interaction tooltip data to annotated elements.
// The sate management is done outside of the Redux container, thought it could be there; some of this
// stuff we don't need synchonized to browser local storage.

let tooltipCache = {};
let j2dCache = {};

const HOVER_DELAY_MS = 200;

let tooltip;
let loadingTooltip;
function createTooltips(): void {
	if (tooltip || loadingTooltip) {
		return;
	}

	tooltip	= document.createElement("DIV");
	tooltip.className = tooltipClassName;
	tooltip.classList.add("sg-tooltip");

	loadingTooltip = document.createElement("DIV");
	loadingTooltip.appendChild(document.createTextNode("Loading..."));
};

let activeTarget;
function getTarget(t: HTMLElement): HTMLElement | undefined {
	if (t.tagName === "TD") {
		// Not hovering over any token in particular.
		return;
	}
	while (t && t.tagName !== "TD" && !t.getAttribute("data-byteoffset")) {
		t = (t.parentNode as any); // TODO(john): remove cast
	}
	if (t && t.tagName === "SPAN" && t.getAttribute("data-byteoffset")) {
		return t;
	}
}

let hoverTimeout;
function clearTooltip(): void {
	clearTimeout(hoverTimeout);
	if (tooltip.firstChild) {
		tooltip.removeChild(tooltip.firstChild);
	}
	tooltip.style.visibility = "hidden"; // prevent black dot of empty content
}

function addEventListeners(el: HTMLElement, path: string, repoRevSpec: RepoRevSpec, line: number, loggingStruct: Object): void {
	createTooltips();

	el.onclick = (e) => {
		let t = getTarget(e.target as HTMLElement);
		if (!t || t.style.cursor !== "pointer") {
			return;
		}

		fetchJumpURL(t.dataset["byteoffset"], (defUrl) => {
			if (!defUrl) {
				return;
			}

			// If cmd/ctrl+clicked or middle button clicked, open in new tab/page otherwise
			// either move to a line on the same page, or refresh the page to a new blob view.
			EventLogger.logEventForCategory("Def", "Click", "JumpDef", Object.assign({}, repoRevSpec, loggingStruct));
			window.open(defUrl, "_blank");
		});
	};

	el.onmouseout = (e) => {
		clearTooltip();
		activeTarget = null;
	};

	el.onmouseover = (e) => {
		let t = getTarget(e.target as HTMLElement);
		if (!t) {
			return;
		}

		if (activeTarget !== t) {
			activeTarget = t;
			clearTooltip();
			// Only show "Loading..." if it has been loading for a while. If we
			// show "Loading..." immediately, there will be a visible flash if
			// the actual hover text loads quickly thereafter.
			hoverTimeout = setTimeout(() => showTooltip(loadingTooltip, true), 3 * HOVER_DELAY_MS);

			const hoverShowTime = Date.now() + HOVER_DELAY_MS;
			getTooltip(activeTarget, (elem) => {
				clearTimeout(hoverTimeout); // prevent delayed addition of loading indicator
				if (elem === null) { // add no tooltip for empty responses
					return;
				}

				// Always wait at least HOVER_DELAY_MS before showing hover, to avoid
				// it obscuring text when you move your mouse rapidly across a code file.
				const hoverTimerRemaining = Math.max(0, hoverShowTime - Date.now());
				hoverTimeout = setTimeout(() => showTooltip(elem, false), hoverTimerRemaining);
			});
		}
	};

	function showTooltip(elem: HTMLElement, isLoading: boolean): void {
		clearTooltip();

		if (activeTarget && elem) {
			// Log event only when displaying a fetched tooltip
			if (!isLoading) {
				EventLogger.logEventForCategory("Def", "Hover", "HighlightDef", Object.assign({}, repoRevSpec, loggingStruct));
			}

			tooltip.appendChild(elem);

			// Hide the tooltip initially while we add it to DOM to render and generate
			// bounding rectangle but hide as we haven't anchored it to a position yet.
			tooltip.style.visibility = "hidden";

			// Anchor it horizontally, prior to rendering to account for wrapping
			// changes to vertical height if the tooltip is at the edge of the viewport.
			const activeTargetBound = activeTarget.getBoundingClientRect();
			tooltip.style.left = (activeTargetBound.left + window.scrollX) + "px";

			// Attach the node to DOM so the bounding rectangle is generated.
			document.body.appendChild(tooltip);

			// Anchor the tooltip vertically.
			const tooltipBound = tooltip.getBoundingClientRect();
			tooltip.style.top = (activeTargetBound.top - (tooltipBound.height + 5) + window.scrollY) + "px";

			// Make it all visible to the user.
			tooltip.style.visibility = "visible";
		}
	}

	function wrapLSP(req: any): any[] {
		req.id = 1;
		return [
			{
				id: 0,
				method: "initialize",
				params: {
					rootPath: `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}`,
					mode: `${utils.getModeFromExtension(utils.getPathExtension(path))}`,
				},
			},
			req,
			{
				id: 2,
				method: "shutdown",
			},
			{
				method: "exit",
			},
		];
	}

	function fetchJumpURL(col: string, cb: (val: any) => void): void {
		const cacheKey = `${path}@${repoRevSpec.rev}:${line}@${col}`;
		if (typeof j2dCache[cacheKey] !== "undefined") {
			return cb(j2dCache[cacheKey]);
		}

		const body = wrapLSP({
			method: "textDocument/definition",
			params: {
				textDocument: {
					uri: `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}#${path}`,
				},
				position: {
					character: parseInt(col, 10) - 1,
					line: line - 1,
				},
			},
		});

		fetch("https://sourcegraph.com/.api/xlang/textDocument/definition", {method: "POST", body: JSON.stringify(body)})
			.then((resp) => resp.json().then((json) => {
				const respUri = json[1].result[0].uri.split("git://")[1];
				const prt0Uri = respUri.split("?");
				const prt1Uri = prt0Uri[1].split("#");

				const repoUri = prt0Uri[0];
				const frevUri = (repoUri === repoRevSpec.repoURI ? repoRevSpec.relRev : prt1Uri[0]) || "master";
				const pathUri = prt1Uri[1];
				const lineUri = parseInt(json[1].result[0].range.start.line, 10) + 1;

				j2dCache[cacheKey] = `https://sourcegraph.com/${repoUri}@${frevUri}/-/blob/${pathUri}${lineUri ? "#L" + lineUri : ""}`;
				cb(j2dCache[cacheKey]);
			})).catch((err) => cb(null));
	}

	function getTooltip(target: HTMLElement, cb: (val: any) => void): void {
		const cacheKey = `${path}@${repoRevSpec.rev}:${line}@${target.dataset["byteoffset"]}`;
		if (typeof tooltipCache[cacheKey] !== "undefined") {
			return cb(tooltipCache[cacheKey]);
		}

		const body = wrapLSP({
			method: "textDocument/hover",
			params: {
				textDocument: {
					uri: `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}#${path}`,
				},
				position: {
					character: parseInt(target.dataset["byteoffset"], 10) - 1,
					line: line - 1,
				},
			},
		});

		fetch("https://sourcegraph.com/.api/xlang/textDocument/hover", {method: "POST", body: JSON.stringify(body)})
			.then((resp) => resp.json().then((json) => {
				const tooltipText = document.createElement("DIV");
				if (json[1].result && json[1].result.contents && json[1].result.contents.length > 0) {
					target.style.cursor = "pointer";
					target.className = `${target.className} sg-clickable`;

					tooltipText.className = tooltipTitleStyle;
					tooltipText.appendChild(document.createTextNode(json[1].result.contents[0].value));
					tooltipCache[cacheKey] = tooltipText;
				} else {
					tooltipCache[cacheKey] = null;
				}
				cb(tooltipCache[cacheKey]);
			})).catch((err) => cb(null));
	}
}
