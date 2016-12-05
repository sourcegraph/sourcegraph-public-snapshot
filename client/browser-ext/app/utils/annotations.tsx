import { doFetch as fetch } from "../backend/xhr";
import { EventLogger } from "../utils/EventLogger";
import * as github from "./github";
import * as utils from "./index";
import * as tooltips from "./tooltips";
import * as _ from "lodash";

interface RepoRevSpec {
	repoURI: string;
	rev: string;
	isDelta: boolean;
	isBase: boolean;
}

// addAnnotations is the entry point for injecting annotations onto a blob (el).
// An invisible marker is appended to the document to indicate that annotation
// has been completed; so this function expects that it will be called once all
// repo/annotation data is resolved from the server.
export function addAnnotations(path: string, repoRevSpec: RepoRevSpec, el: HTMLElement, isSplitDiff: boolean, loggingStruct: Object): void {
	prewarmLSP(path, repoRevSpec);

	// The blob is represented by a table; the first column is the line number,
	// the second is code. Each row is a line of code
	const table = el.querySelector("table");
	if (!table) {
		return;
	}

	const cells = github.getCodeCellsForAnnotation(table, Object.assign({ isSplitDiff }, repoRevSpec));
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
			try {
				const annLine = convertNode(cell.cell, 1, cell.line, repoRevSpec.isDelta);
				cell.cell.innerHTML = "";
				cell.cell.appendChild(annLine.resultNode);

				addEventListeners(cell.cell, path, repoRevSpec, cell.line, loggingStruct);
			} catch (e) {
				console.error(e);
			}
		});
	});
}

interface ConvertNodeResult<T extends Node> {
	resultNode: T;
	bytesConsumed: number;
}

function convertNodeHelper(node: Node, offset: number, line: number, ignoreFirstTextChar: boolean): ConvertNodeResult<Element> {
	switch (node.nodeType) {
		case Node.TEXT_NODE:
			return convertTextNode(node, offset, line, ignoreFirstTextChar);

		case Node.ELEMENT_NODE:
			return convertElementNode(node, offset, line, ignoreFirstTextChar);

		default:
			throw new Error(`unexpected node type(${node.nodeType})`);
	}
}

// convertNode takes a DOM node and returns an object containing the
// maybe-linkified version of the node as an HTML string as well as the number of bytes consumed.
// It is the entry point for converting a <td> "cell" representing a line of code.
// It may also be called recursively for children (which are assumed to be <span>
// code highlighting annotations from GitHub).
export function convertNode(currentNode: Node, offset: number, line: number, ignoreFirstTextChar: boolean): ConvertNodeResult<Node> {
	let wrapperNode;
	let c = convertNodeHelper(currentNode, offset, line, ignoreFirstTextChar);

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
export function convertTextNode(currentNode: Node, offset: number, line: number, ignoreFirstTextChar: boolean): ConvertNodeResult<Element> {
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
		const match = txt.match(/^(\w+)/);
		if (match) {
			return match[0];
		}
		return txt[0];
	}

	while (nodeText.length > 0) {
		const token = consumeNext(nodeText);

		wrapperNode.appendChild(createTextNode(token, offset + prevConsumed));
		prevConsumed += token.length;
		bytesConsumed += token.length;

		nodeText = nodeText.slice(token.length);
	}

	return { resultNode: wrapperNode, bytesConsumed };
}

// convertElementNode takes a DOM node which should be of NodeType.ELEMENT_NODE
// (this must be checked by the caller) and returns an object containing the
//  maybe-linkified version of the node as an HTML string as well as the number of bytes consumed.
export function convertElementNode(currentNode: Node, offset: number, line: number, ignoreFirstTextChar: boolean): ConvertNodeResult<Element> {
	let bytesConsumed = 0;
	const wrapperNode = document.createElement("SPAN");

	wrapperNode.setAttribute("data-byteoffset", `${offset}`);

	// The logic here is to simply recurse on each of the child nodes; everything is eventually
	// just a text node or the special-cased "quoted string node" (see below).
	for (let i = 0; i < currentNode.childNodes.length; ++i) {
		const res = convertNode(currentNode.childNodes[i], offset + bytesConsumed, line, i === 0 && ignoreFirstTextChar);
		wrapperNode.appendChild(res.resultNode);
		bytesConsumed += res.bytesConsumed;
	}

	return { resultNode: wrapperNode, bytesConsumed };
}

let tooltipCache: { [key: string]: tooltips.TooltipData } = {};
let j2dCache = {};

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

function wrapLSP(req: { method: string, params: Object }, repoRevSpec: RepoRevSpec, path: string): Object[] {
	(req as any).id = 1;
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

function addEventListeners(el: HTMLElement, path: string, repoRevSpec: RepoRevSpec, line: number, loggingStruct: Object): void {
	tooltips.createTooltips();

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
		tooltips.clearContext();
		activeTarget = null;
	};

	el.onmouseover = (e) => {
		let t = getTarget(e.target as HTMLElement);
		if (!t || activeTarget === t) {
			// don't do anything unless target is defined and has changed
			return;
		}

		activeTarget = t;
		tooltips.setContext(t, loggingStruct);
		tooltips.queueLoading();
		getTooltip(t, (data) => tooltips.setTooltip(data, t as HTMLElement));
	};

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
		}, repoRevSpec, path);

		fetch("https://sourcegraph.com/.api/xlang/textDocument/definition", { method: "POST", body: JSON.stringify(body) })
			.then((resp) => resp.json().then((json) => {
				const respUri = json[1].result[0].uri.split("git://")[1];
				const prt0Uri = respUri.split("?");
				const prt1Uri = prt0Uri[1].split("#");

				const repoUri = prt0Uri[0];
				const frevUri = (repoUri === repoRevSpec.repoURI ? repoRevSpec.rev : prt1Uri[0]) || "master"; // TODO(john): preserve rev branch
				const pathUri = prt1Uri[1];
				const lineUri = parseInt(json[1].result[0].range.start.line, 10) + 1;

				j2dCache[cacheKey] = `https://sourcegraph.com/${repoUri}@${frevUri}/-/blob/${pathUri}${lineUri ? "#L" + lineUri : ""}`;
				cb(j2dCache[cacheKey]);
			})).catch((err) => cb(null));
	}

	function getTooltip(target: HTMLElement, cb: (val: tooltips.TooltipData) => void): void {
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
		}, repoRevSpec, path);

		fetch("https://sourcegraph.com/.api/xlang/textDocument/hover", { method: "POST", body: JSON.stringify(body) })
			.then((resp) => resp.json().then((json) => {
				if (json[1].result && json[1].result.contents && json[1].result.contents.length > 0) {
					const title = json[1].result.contents[0].value;
					let doc;
					json[1].result.contents.filter((markedString) => markedString.language === "markdown").forEach((content) => {
						// TODO(john): what if there is more than 1?
						doc = content.value;
					});
					tooltipCache[cacheKey] = { title, doc };
				} else {
					tooltipCache[cacheKey] = null;
				}
				cb(tooltipCache[cacheKey]);
			})).catch((err) => cb(null));
	}
}

const prewarmCache = new Set<string>();
function prewarmLSP(path: string, repoRevSpec: RepoRevSpec): void {
	const uri = `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}#${path}`;
	if (prewarmCache.has(uri)) {
		return;
	}
	prewarmCache.add(uri);

	const body = wrapLSP({
		method: "textDocument/hover?prepare",
		params: {
			textDocument: { uri },
			position: {
				character: 0,
				line: 0,
			},
		},
	}, repoRevSpec, path);

	fetch("https://sourcegraph.com/.api/xlang/textDocument/hover?prepare", { method: "POST", body: JSON.stringify(body) });
}
