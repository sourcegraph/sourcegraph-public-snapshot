import utf8 from "utf8";
import fetch from "../actions/xhr";
import styles from "../components/App.css";
import * as utils from "./index";
import _ from "lodash";
import EventLogger from "../analytics/EventLogger";

// addAnnotations is the entry point for injecting annotations onto a blob (el).
// An invisible marker is appended to the document to indicate that annotation
// has been completed; so this function expects that it will be called once all
// repo/annotation data is resolved from the server.
export default function addAnnotations(path, repoRevSpec, el, anns, lineStartBytes, isSplitDiff) {
	_applyAnnotations(el, path, repoRevSpec, indexAnnotations(anns).annsByStartByte, indexLineStartBytes(lineStartBytes), isSplitDiff);
}

// _applyAnnotations is a helper function for addAnnotations
export function _applyAnnotations(el, path, repoRevSpec, annsByStartByte, startBytesByLine, isSplitDiff) {
	// The blob is represented by a table; the first column is the line number,
	// the second is code. Each row is a line of code
	const table = el.querySelector("table");

	let cells = [];
	for (let i = 0; i < table.rows.length; ++i) {
		const row = table.rows[i];
		if (row.classList && row.classList.contains("inline-comments")) continue;

		let line, codeCell;
		if (repoRevSpec.isDelta) {
			if (isSplitDiff && row.cells.length !== 4) continue;

			let metaCell;
			if (isSplitDiff) {
				metaCell = repoRevSpec.isBase ? row.cells[0] : row.cells[2];
			} else {
				metaCell = repoRevSpec.isBase ? row.cells[0] : row.cells[1];
			}

			if (metaCell.classList && metaCell.classList.contains("blob-num-expandable")) {
				continue;
			}

			if (isSplitDiff) {
				codeCell = repoRevSpec.isBase ? row.cells[1] : row.cells[3];
			} else {
				codeCell = row.cells[2];
			}
			if (!codeCell) {
				continue;
			}

			if (codeCell.classList && codeCell.classList.contains("blob-code-empty")) {
				continue;
			}

			let isAddition = codeCell.classList && codeCell.classList.contains("blob-code-addition");
			let isDeletion = codeCell.classList && codeCell.classList.contains("blob-code-deletion");
			if (!isAddition && !isDeletion && !repoRevSpec.isBase && !isSplitDiff) {
				continue; // careful; we don't need to put head AND base on unmodified parts (but only for unified diff views)
			}
			if (isDeletion && !repoRevSpec.isBase) {
				continue;
			}
			if (isAddition && repoRevSpec.isBase) {
				continue;
			}

			line = metaCell.dataset.lineNumber;
			if (line === "..." || !line) {
				continue;
			}
		} else {
			line = row.cells[0].dataset.lineNumber;
			codeCell = row.cells[1];
		}

		// If the line has already been annotated,
		// restore event handlers if necessary otherwise move to next line
		if (el.dataset[`${line}_${repoRevSpec.rev}`]) {
			if (!el.onclick || !el.onmouseout || !el.onmouseover) {
				addEventListeners(codeCell, path, repoRevSpec, line);
			}
			continue;
		}
		el.dataset[`${line}_${repoRevSpec.rev}`] = true;

		const offset = startBytesByLine[line];

		// result is the new (annotated) innerHTML of the code cell
		const {result, bytesConsumed} = convertNode(codeCell, annsByStartByte, offset, offset, repoRevSpec);
		// manipulate the DOM asynchronously so the page doesn't freeze while large
		// code files are being annotated
		let cell = codeCell;
		setTimeout(() => {
			const inner = cell.querySelector(".blob-code-inner");
			if (inner) {
				inner.innerHTML = result;
			} else {
				cell.innerHTML = result;
			}
			addEventListeners(cell, path, repoRevSpec, line);
		});
	}
}

// indexAnnotations creates a fast lookup structure optimized to query
// annotations by start or end byte.
export function indexAnnotations(anns) {
	let annsByStartByte = {};
	let annsByEndByte = {};
	for (let i = 0; i < anns.length; i++) {
		// From pkg/syntaxhighlight/html_annotator.go
		const annType = anns[i].Class;
		if (annType !== "com" || annType !== "lit" || annType !== "pun") {
			let ann = anns[i];
			annsByStartByte[ann.StartByte] = ann;
			annsByEndByte[ann.EndByte] = ann;
		}
	}
	return {annsByStartByte, annsByEndByte};
}

// indexLineStartBytes creates a fast lookup structure optimized to query
// byte offsets by line number (1-indexed).
export function indexLineStartBytes(lineStartBytes) {
	let startBytesByLine = {};
	for (let i = 0; i < lineStartBytes.length; i++) {
		startBytesByLine[i + 1] = lineStartBytes[i];
	}
	return startBytesByLine;
}

// annGenerator returns a "match" object if an anotation is defined
// at the byte offset. The match result contains the number of bytes
// matched by the annotation, and a generator function which returns
// an HTML anchor tag string .
export function annGenerator(annsByStartByte, byte, lineStart) {
	const match = annsByStartByte[byte];
	if (!match) return null;

	const annLen = match.EndByte - match.StartByte;
	if (annLen <= 0) return null; // sometimes, there will be an "empty" annotation, e.g. for CommonJS modules

	return {annLen, annGen: function(innerHTML) {
		return `<span data-byteoffset=${byte + 1 - lineStart} style="cursor:pointer;">${innerHTML}</span>`;
	}};
}

// getOpeningTag returns the starting tag (with attributes) of the node
// (assumed to be of type NodeType.ELEMENT_NODE). E.g.
//     <span attr="foo">hello world</span>
// would return '<span attr="foo">'.
// This is a fairly naive implementation that may not work if we were writing a
// full-blown HTML parser; but since we only have to parse GitHub's blob HTML
// we can expect more regularity.
export function getOpeningTag(node) {
	let i;
	let inAttribute = false;
	const outerHTML = node.outerHTML;
	for (i = 0; i < outerHTML.length; ++i) {
		if (outerHTML[i] === "\"") inAttribute = !inAttribute;
		if (outerHTML[i] === ">" && !inAttribute) break;
	}
	return outerHTML.substring(0, i+1);
}

// convertNode takes a DOM node and returns an object containing the
// maybe-linkified version of the node as an HTML string as well as the number of bytes consumed.
// It is the entry point for converting a <td> "cell" representing a line of code.
// It may also be called recursively for children (which are assumed to be <span>
// code highlighting annotations from GitHub).
export function convertNode(node, annsByStartByte, offset, lineStart, repoRevSpec, ignoreFirstTextChar) {
	let result, bytesConsumed, c; // c is an intermediate result
	if (node.nodeType === Node.ELEMENT_NODE) {
		// The logic here is to:
		//    - convert as element node (which may be the special-cased "quoted string" node)
		//    - ^^ gives inner html; wrap this with the node's current syntax highlighting <span>
		//    - ^^ but don't do that if the top-level tag is the <td> element (entrypoint)

		const isTableCell = node.tagName === "TD";
		ignoreFirstTextChar = repoRevSpec.isDelta && isTableCell; // +, -, or whitespace preceeds all code
		if (isTableCell) {
			// For diff blobs, the td can have extraneous child (text) nodes of whitespace that shouldn't
			// be annotated; select the ".blob-code-inner" element which has only the code we
			// care to annotate. (For normal blobs, there is no .blob-code-inner).
			const inner = node.querySelector(".blob-code-inner");
			if (inner) node = inner;
		}

		c = isStringNode(node) || isCommentNode(node) ?
			convertStringNode(node, annsByStartByte, offset, lineStart) :
			convertElementNode(node, annsByStartByte, offset, lineStart, repoRevSpec, ignoreFirstTextChar);

		if (!isTableCell) {
			const openTag = getOpeningTag(node);
			const closeTag = "</span>";
			if (openTag.indexOf("<span") !== 0) {
				throw new Error(`element node tag is not SPAN, got(${node.tagName}), parsed(${openTag})`);
			}
			result = `${openTag}${c.result}${closeTag}`;
		} else {
			result = c.result;
		}
		bytesConsumed = c.bytesConsumed;
	} else if (node.nodeType === Node.TEXT_NODE) {
		c = convertTextNode(node, annsByStartByte, offset, lineStart, ignoreFirstTextChar);
		result = c.result;
		bytesConsumed = c.bytesConsumed;
	} else {
		throw new Error(`unexpected node type(${node.nodeType})`);
	}

	return {result, bytesConsumed};
}

// convertTextNode takes a DOM node which should be of NodeType.TEXT_NODE
// (this must be checked by the caller) and returns an object containing the
//  maybe-linkified version of the node as an HTML string and the number
// of bytes consumed.
export function convertTextNode(node, annsByStartByte, offset, lineStart, ignoreFirstTextChar) {
	let innerHTML = [];
	let bytesConsumed;

	// Text could contain escaped character sequences (e.g. "&gt;") or UTF-8
	// decoded characters (e.g. "˟") which need to be properly counted in terms of bytes.
	let nodeText = utf8.encode(_.unescape(node.wholeText)).split("");
	if (ignoreFirstTextChar && nodeText.length > 0) {
		innerHTML.push(nodeText[0]);
		nodeText = nodeText.slice(1);
	}
	for (bytesConsumed = 0; bytesConsumed < nodeText.length;) {
		const match = annGenerator(annsByStartByte, offset + bytesConsumed, lineStart);
		if (!match) {
			innerHTML.push(_.escape(nodeText[bytesConsumed++]));
			continue;
		}

		innerHTML.push(match.annGen(_.escape(nodeText.slice(bytesConsumed, bytesConsumed + match.annLen).join(""))));
		bytesConsumed += match.annLen;
	}

	return {result: utf8.decode(innerHTML.join("")), bytesConsumed};
}

// convertElementNode takes a DOM node which should be of NodeType.ELEMENT_NODE
// (this must be checked by the caller) and returns an object containing the
//  maybe-linkified version of the node as an HTML string as well as the number of bytes consumed.
export function convertElementNode(node, annsByStartByte, offset, lineStart, repoRevSpec, ignoreFirstTextChar) {
	let innerHTML = [];
	let bytesConsumed = 0;

	// The logic here is to simply recurse on each of the child nodes; everything is eventually
	// just a text node or the special-cased "quoted string node" (see below).
	for (let i = 0; i < node.childNodes.length; ++i) {
		const res = convertNode(node.childNodes[i], annsByStartByte, offset + bytesConsumed, lineStart, repoRevSpec, i === 0 && ignoreFirstTextChar);
		innerHTML.push(res.result);
		bytesConsumed += res.bytesConsumed;
	}

	return {result: utf8.decode(innerHTML.join("")), bytesConsumed};
}

// isStringNode and isCommentNode are predicate functions to
// identify string identifier DOM elements, as per Github's code styling classes.
export function isCommentNode(node) {
	return node.classList.contains("pl-c");
}

export function isStringNode(node) {
	return node.classList.contains("pl-s");
}

// convertStringNode takes a DOM node which is a plain string and returns an object containing the
// maybe-linkified version of the node as an HTML string as well as the number of bytes consumed.
export function convertStringNode(node, annsByStartByte, offset, lineStart) {
	function getChildNodeText(node) {
		if (node.nodeType === Node.ELEMENT_NODE) {
			return [].map.call(node.childNodes, getChildNodeText).join("");
		} else if (node.nodeType === Node.TEXT_NODE) {
			return utf8.encode(_.unescape(node.wholeText));
		} else {
			throw new Error(`unexpected node type(${node.nodeType})`);
		}
	}

	const strTxt = getChildNodeText(node);
	const annGen = function(innerHTML) {
		return `<span data-byteoffset=${offset + 1 - lineStart} style=${isCommentNode(node) ? "" : "cursor:pointer;"}>${innerHTML}</span>`;
	};

	return {result: annGen(_.escape(strTxt)), bytesConsumed: strTxt.length};
}

// The rest of this file is responsible for fetching/caching annotation specific data from the server
// and adding interaction popover data to annotated elements.
// The sate management is done outside of the Redux container, thought it could be there; some of this
// stuff we don't need synchonized to browser local storage.

let popoverCache = {};
let jumptodefcache = {};

const HOVER_TIME = 200;
let hoverTimeout;

function addEventListeners(el, path, repoRevSpec, line) {
	let activeTarget, popover;

	el.onclick = function(e) {
		let t = getTarget(e.target);
		if (!t) return;

		let openInNewTab = e.ctrlKey || (navigator.platform.toLowerCase().indexOf('mac') >= 0 && e.metaKey) || e.button === 1;

		fetchJumpURL(t.dataset.byteoffset, function(defObj) {
			// If cmd/ctrl+clicked or middle button clicked, open in new tab/page otherwise
			// either move to a line on the same page, or refresh the page to a new blob view.
			if (defObj.defCurPage && !repoRevSpec.isDelta) {
				location.hash = defObj.defUrl.slice(defObj.defUrl.indexOf('#'));
			} else {
				if (openInNewTab) {
					window.open(defObj.defUrl, "_blank");
				} else {
					location.href = defObj.defUrl;
				}
			}
		});
	}

	el.onmouseout = function(e) {
		hidePopover();
		activeTarget = null;
	}

	el.onmouseover = function(e) {
		let t = getTarget(e.target);
		if (!t) return;

		if (activeTarget !== t) {
			activeTarget = t;

			hidePopover();
			hoverTimeout = setTimeout(() => {
				// Only show "Loading..." if it has been loading for a while. If we
				// show "Loading..." immediately, there will be a visible flash if
				// the actual hover text loads quickly thereafter.
				if (!popover) {
					showPopover(`<div><div class=${styles.popoverTitle}>Loading...</div></div>`, true);
				}
			}, 3 * HOVER_TIME);
			const hoverShowTime = Date.now() + HOVER_TIME;
			fetchPopoverData(activeTarget.dataset.byteoffset, function(html) {
				// Always wait at least HOVER_TIME before showing hover, to avoid
				// it obscuring text when you move your mouse rapidly across a code file.
				const hoverTimerRemaining = Math.max(0, hoverShowTime - Date.now());
				hidePopover();
				hoverTimeout = setTimeout(() => {
					showPopover(html, false);
				}, hoverTimerRemaining);
			});
		}
	}

	function getTarget(t) {
		while (t && t.tagName !== "TD" && !t.dataset && !t.dataset.byteoffset) {t = t.parentNode;}
		if (t && t.tagName === "SPAN" && t.dataset && t.dataset.byteoffset) return t;
	}

	function showPopover(html, isLoading) {
		if (!popover && activeTarget && html) {
			// Log event only when displaying a fetched tooltip
			if (!isLoading) {
				EventLogger.logEventForCategory("Def", "Hover", "HighlightDef", {isDelta: repoRevSpec.isDelta, language: utils.getPathExtension(path)});
			}

			// Create and style the element
			popover = document.createElement("div");
			popover.classList.add(styles.popover);
			popover.classList.add("sg-popover");
			popover.innerHTML = html;

			// Hide the popover initially while we add it to DOM to render and generate
			// bounding rectangle but hide as we haven't anchored it to a position yet.
			popover.style.visibility = "hidden";

			// Attach the node to DOM so the bounding rectangle is generated.
			document.body.appendChild(popover);

			// Anchor the popover above the symbol.
			const popoverBound = popover.getBoundingClientRect();
			const activeTargetBound = activeTarget.getBoundingClientRect();

			popover.style.top = (activeTargetBound.top - (popoverBound.height + 5) + window.scrollY) + "px";
			popover.style.left = (activeTargetBound.left + window.scrollX) + "px";

			// Make it all visible to the user.
			popover.style.visibility = "visible";
		}
	}

	function hidePopover() {
		if (hoverTimeout) {
			clearTimeout(hoverTimeout);
		}
		if (popover) {
			popover.remove();
			popover = null;
		}
	}

	function fetchJumpURL(col, cb) {
		if (typeof jumptodefcache[`${path}@${repoRevSpec.rev}:${line}@${col}`] !== 'undefined') {
			return cb(jumptodefcache[`${path}@${repoRevSpec.rev}:${line}@${col}`]);
		}

		const body = [
			{
				id: 0,
				method: "initialize",
				params: {
					rootPath: `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}`,
					mode: `${utils.getModeFromExtension(utils.getPathExtension(path))}`,
				},
			},
			{
				id: 1,
				method: "textDocument/definition",
				params: {
					textDocument: {
						uri: `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}#${path}`,
					},
					position: {
						character: parseInt(col, 10) - 1,
						line: parseInt(line, 10) - 1,
					},
				},
			},
			{
				id: 2,
				method: "shutdown",
			},
			{
				method: "exit",
			},
		];

		return fetch("https://sourcegraph.com/.api/xlang/textDocument/definition", {method: "POST", body: JSON.stringify(body)})
			.then((json) => {
				const respUri = json[1].result[0].uri.split("git://")[1];
				const prt0Uri = respUri.split("?");
				const prt1Uri = prt0Uri[1].split("#");

				const repoUri = prt0Uri[0];
				const frevUri = prt1Uri[0] || "master";
				const pathUri = prt1Uri[1];
				const lineUri = parseInt(json[1].result[0].range.start.line, 10) + 1;

				jumptodefcache[`${path}@${repoRevSpec.rev}:${line}@${col}`] = {
					defUrl: `https://${repoUri}/blob/${frevUri}/${pathUri}${lineUri ? "#L" + lineUri : "" }`,
					defCurPage: path === pathUri,
				};

				cb(jumptodefcache[`${path}@${repoRevSpec.rev}:${line}@${col}`]);
			})
			.catch((err) => {});
	}

	function fetchPopoverData(col, cb) {
		if (typeof popoverCache[`${path}@${repoRevSpec.rev}:${line}@${col}`] !== "undefined") {
			return cb(popoverCache[`${path}@${repoRevSpec.rev}:${line}@${col}`]);
		}

		const body = [
			{
				id: 0,
				method: "initialize",
				params: {
					rootPath: `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}`,
					mode: `${utils.getModeFromExtension(utils.getPathExtension(path))}`,
				},
			},
			{
				id: 1,
				method: "textDocument/hover",
				params: {
					textDocument: {
						uri: `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}#${path}`,
					},
					position: {
						character: parseInt(col, 10) - 1,
						line: parseInt(line, 10) - 1,
					},
				},
			},
			{
				id: 2,
				method: "shutdown",
			},
			{
				method: "exit",
			},
		];

		return fetch("https://sourcegraph.com/.api/xlang/textDocument/hover", {method: "POST", body: JSON.stringify(body)})
			.then((json) => {
				const keyCache = `${path}@${repoRevSpec.rev}:${line}@${col}`;
				try {
					popoverCache[keyCache] = `<div><div class=${styles.popoverTitle}>${json[1].result.contents[0].value}</div></div>`;
				} catch(err) {
					popoverCache[keyCache] = null;
				} finally {
					cb(popoverCache[keyCache]);
				}
			})
			.catch((err) => {cb(null)});
	}
}
