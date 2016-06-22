import utf8 from "utf8";
import fetch from "../actions/xhr";
import styles from "../components/App.css";
import _ from "lodash";
import EventLogger from "../analytics/EventLogger";

// addAnnotations is the entry point for injecting annotations onto a blob (el).
// An invisible marker is appended to the document to indicate that annotation
// has been completed; so this function expects that it will be called once all
// repo/annotation data is resolved from the server.
export default function addAnnotations(path, revSpec, el, anns, lineStartBytes) {
	const markerID = `sourcegraph-annotation-marker-${path}-@${revSpec.rev}`;
	if (document.getElementById(markerID)) {
		// Don't let annotations be applied twice (it is not idempotent).
		return;
	} else {
		const marker = document.createElement("div");
		marker.id = markerID;
		marker.style.display = "none";
		el.appendChild(marker);
	}
	_applyAnnotations(el, revSpec, indexAnnotations(anns).annsByStartByte, indexLineStartBytes(lineStartBytes));
	_postProcess();
}

let _annURLs = []; // global state container for annotation URLs matched to tokens
// _postProcess fetches metadata for each of the repositories matched to the
// page within annotations
function _postProcess() {
	_annURLs.forEach((url) => setTimeout(() => cacheDefaultBranch(url)));
	_annURLs = [];
}
// _applyAnnotations is a helper function for addAnnotations
export function _applyAnnotations(el, {isDelta, isBase}, annsByStartByte, startBytesByLine) {
	// The blob is represented by a table; the first column is the line number,
	// the second is code. Each row is a line of code
	const table = el.querySelector("table");

	let cells = [];
	for (let i = 0; i < table.rows.length; ++i) {
		const row = table.rows[i];
		if (row.classList && row.classList.contains("inline-comments")) continue;

		function removeLeadingChar(cell) {
			const innerBlob = cell.querySelector(".blob-code-inner");
			if (!innerBlob) return;

			const val = innerBlob.firstChild.nodeValue;
			innerBlob.firstChild.nodeValue = val.substring(1, val.length);
		}

		function addChar(cell, char) {
			const innerBlob = cell.querySelector(".blob-code-inner");
			if (!innerBlob) return;

			if (!innerBlob.firstChild) {
				innerBlob.appendChild(document.createTextNode(char));
			} else if (innerBlob.firstChild.nodeType !== Node.TEXT_NODE) {
				innerBlob.insertBefore(document.createTextNode(char), innerBlob.firstChild);
			} else {
				innerBlob.firstChild.nodeValue = `${char}${innerBlob.firstChild.nodeValue}`;
			}
		}

		let line, codeCell, isAddition, isDeletion;
		if (isDelta) {
			let metaCell = isBase ? row.cells[0] : row.cells[1];
			if (metaCell.classList.contains("blob-num-expandable")) {
				continue;
			}

			codeCell = row.cells[2];
			if (!codeCell) {
				continue;
			}

			isAddition = codeCell.classList.contains("blob-code-addition");
			isDeletion = codeCell.classList.contains("blob-code-deletion");
			if (!isAddition && !isDeletion && !isBase) {
				continue; // careful; we don't need to put head AND base on unmodified parts
			}
			if (isDeletion && !isBase) {
				continue;
			}
			if (isAddition && isBase) {
				continue;
			}

			// +, -, or whitespace preceeds all code
			removeLeadingChar(codeCell);

			line = metaCell.dataset.lineNumber;
			if (line === "..." || !line) {
				continue;
			}
		} else {
			line = row.cells[0].dataset.lineNumber;
			codeCell = row.cells[1];
		}

		const offset = startBytesByLine[line];

		// result is the new (annotated) innerHTML of the code cell
		const {result, bytesConsumed} = convertNode(codeCell, annsByStartByte, offset);

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
			if (isAddition) {
				addChar(cell, "+");
			} else if (isDeletion) {
				addChar(cell, "-");
			} else {
				addChar(cell, " ");
			}
			addPopover(cell);
		});
	}
}

// indexAnnotations creates a fast lookup structure optimized to query
// annotations by start or end byte.
export function indexAnnotations(anns) {
	let annsByStartByte = {};
	let annsByEndByte = {};
	for (let i = 0; i < anns.length; i++) {
		if (anns[i].URL) { // without a URL, it is a syntax highlighting annotation
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
export function annGenerator(annsByStartByte, byte) {
	const match = annsByStartByte[byte];
	if (!match) return null;

	function urlToDef(origURL) {
		if (!origURL) return null;
		const parts = origURL.split("/-/");
		if (parts.length < 2) return null;
		const repo = parts[0]; // remove leading slash
		const def = parts.slice(1).join("/-/").replace("def/", "");
		if (repo.startsWith("/github.com/")) {
			return `https:/${repo}#sourcegraph&def=${def}`;
		}
		return `https://github.com/#sourcegraph&repo=${repo}&def=${def}`;
	}

	const defIsOnGitHub = match.URL && match.URL.includes("github.com/");
	const url = defIsOnGitHub ? urlToDef(match.URL) : `https://sourcegraph.com${match.URL}`;

	const annLen = match.EndByte - match.StartByte;
	return {annLen, annGen: function(innerHTML) {
		_annURLs.push(url);
		return `<a href="${url}" ${defIsOnGitHub ? "data-sourcegraph-ref" : "target=tab"} data-src="https://sourcegraph.com${match.URL}" class=${styles.sgdef}>${innerHTML}</a>`;
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
export function convertNode(node, annsByStartByte, offset) {
	let result, bytesConsumed, c; // c is an intermediate result
	if (node.nodeType === Node.ELEMENT_NODE) {
		// The logic here is to:
		//    - convert as element node (which may be the special-cased "quoted string" node)
		//    - ^^ gives inner html; wrap this with the node's current syntax highlighting <span>
		//    - ^^ but don't do that if the top-level tag is the <td> element (entrypoint)

		const isTableCell = node.tagName === "TD";
		if (isTableCell) {
			// For diff blobs, the td can have extraneous child (text) nodes of whitespace that shouldn't
			// be annotated; select the ".blob-code-inner" element which has only the code we
			// care to annotate. (For normal blobs, there is no .blob-code-inner).
			const inner = node.querySelector(".blob-code-inner");
			if (inner) node = inner;
		}

		c = isQuotedStringNode(node) ?
			convertQuotedStringNode(node, annsByStartByte, offset) :
			convertElementNode(node, annsByStartByte, offset);

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
		c = convertTextNode(node, annsByStartByte, offset);
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
export function convertTextNode(node, annsByStartByte, offset) {
	let innerHTML = [];
	let bytesConsumed;

	// Text could contain escaped character sequences (e.g. "&gt;") or UTF-8
	// decoded characters (e.g. "˟") which need to be properly counted in terms of bytes.
	const nodeText = utf8.encode(_.unescape(node.wholeText)).split("");
	for (bytesConsumed = 0; bytesConsumed < nodeText.length;) {
		const match = annGenerator(annsByStartByte, offset + bytesConsumed);
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
export function convertElementNode(node, annsByStartByte, offset) {
	let innerHTML = [];
	let bytesConsumed = 0;

	// The logic here is to simply recurse on each of the child nodes; everything is eventually
	// just a text node or the special-cased "quoted string node" (see below).
	for (let i = 0; i < node.childNodes.length; ++i) {
		const res = convertNode(node.childNodes[i], annsByStartByte, offset + bytesConsumed);
		innerHTML.push(res.result);
		bytesConsumed += res.bytesConsumed;
	}

	return {result: utf8.decode(innerHTML.join("")), bytesConsumed};
}

// isQuotedStringNode is a predicate function to identify the special-cased
// string identifier DOM element, which takes a like of go code like this:
//
//    import (
//        "fmt"
//    )
//
// and creates this (for the "fmt"):
//
//    "<span class="pl-s"><span class="pl-pds">"</span>fmt<span class="pl-pds">"</span></span>"
//
// Without the special-casing the <a class="sg-link" /> tag will be put around the opening quote,
// but the total # of bytes consumed would automatically count the rest of the fmt".
// This guarantees the annotation consumes the entire set of childNodes.
export function isQuotedStringNode(node) {
	return node.childNodes.length === 3 && node.querySelectorAll(".pl-pds").length === 2 &&
		node.innerText.startsWith("\"") && node.innerText.endsWith("\"");
}

// convertQuotedStringNode takes a DOM node which should pass the isQuotedStringNode predicate
// (this must be checked by the caller) and returns an object containing the
//  maybe-linkified version of the node as an HTML string as well as the number of bytes consumed.
export function convertQuotedStringNode(node, annsByStartByte, offset) {
	const text = `"${utf8.encode(_.unescape(node.childNodes[1].wholeText))}"`; // put quotes around the sanitized inner text
	const match = annGenerator(annsByStartByte, offset);

	// NOTE:
	// match could be undefined if the annotation doesn't consume the opening start quote;
	// we assume there is no chance that the string inside the quotes would otherwise have annotations.
	// ^^ this assumption may break some day, but I haven't seen it do so yet. Even so, we could
	// check/theoretically handle this case, but I think it is better not to add more special casing to
	// the special casing which is not yet observed in the wild.

	if (!match) return {result: node.innerHTML, bytesConsumed: text.length};
	if (match.annLen !== text.length) {
		throw new Error(`annotation for quoted string node has length mismatch, got(${match.annLen}) wanted(${text.length})`);
	}
	return {result: match.annGen(node.innerHTML), bytesConsumed: match.annLen};
}



// The rest of this file is responsible for fetching/caching annotation specific data from the server
// and adding interaction popover data to annotated elements.
// The sate management is done outside of the Redux container, thought it could be there; some of this
// stuff we don't need synchonized to browser local storage.

export const defaultBranchCache = {};
// fetchingDefaultBranchCache ensures we only make one API call per repo to get default branch.
export const fetchingDefaultBranchCache = {};
function cacheDefaultBranch(annURL) {
	// Assumes annURL has the form github.com/user/repo. If we can't fetch the default branch, we default to master.
	let annURLsplit = annURL.split(/[\/#]/);
	let annRepo = [annURLsplit[2], annURLsplit[3], annURLsplit[4]]
	let repo = annRepo.join("/");
	if (fetchingDefaultBranchCache[repo]) {
		return;
	}
	if (!defaultBranchCache[repo]) {
		fetchingDefaultBranchCache[repo] = true;
		fetch(`https://sourcegraph.com/.api/repos/${repo}`)
			.then((response) => {
				defaultBranchCache[repo] = response.ok ? response.DefaultBranch : "master";
				fetchingDefaultBranchCache[repo] = false;
			})
			.catch((err) => console.log("Error getting default branch"))
	}
}

let popoverCache = {};
export const defCache = {};
function addPopover(el) {
	let activeTarget, popover;

	el.addEventListener("mouseout", (e) => {
		hidePopover();
		activeTarget = null;
	});

	el.addEventListener("mouseover", (e) => {
		let t = getTarget(e.target);
		if (!t) return;
		if (activeTarget !== t) {
			activeTarget = t;
			let url = activeTarget.dataset.src.split("https://sourcegraph.com")[1];
			url = `https://sourcegraph.com/.api/repos${url}?ComputeLineRange=true&Doc=true`;
			fetchPopoverData(url, function(html, data) {
				if (activeTarget && html) showPopover(html, e.pageX, e.pageY);
			});
		}
	});

	function getTarget(t) {
		while (t && t.tagName === "SPAN") {t = t.parentNode;}
		if (t && t.tagName === "A" && t.classList.contains(styles.sgdef)) return t;
	}

	function showPopover(html, x, y) {
		if (!popover) {
			EventLogger.logEvent("HighlightDef");
			popover = document.createElement("div");
			popover.classList.add(styles.popover);
			popover.innerHTML = html;
			positionPopover(x, y);
			document.body.appendChild(popover);
		}
	}

	function hidePopover() {
		if (popover) {
			popover.remove();
			popover = null;
		}
	}

	function positionPopover(x, y) {
		if (popover) {
			popover.style.top = (y + 15) + "px";
			popover.style.left = (x + 15) + "px";
		}
	}

	function fetchPopoverData(url, cb) {
		if (popoverCache[url]) return cb(popoverCache[url], defCache[url]);
		fetch(url)
			.then((json) => {
				defCache[url] = json;
				let html;
				if (json.Data) {
					const f = json.FmtStrings;
					const doc = json.DocHTML ? `<div>${json.DocHTML.__html}</div>` : "";
					html = `<div><div class=${styles.popoverTitle}>${f.DefKeyword || ""}${f.DefKeyword ? " " : ""}<b style="color:#4078C0">${f.Name.Unqualified}</b>${f.NameAndTypeSeparator || ""}${f.Type.ScopeQualified === f.DefKeyword ? "" : f.Type.ScopeQualified || ""}</div>${doc}<div class=${styles.popoverRepo}>${json.Repo}</div></div>`;
				}
				popoverCache[url] = html;
				cb(html, json);
			})
			.catch((err) => console.log("Error getting definition info.") && cb(null, null));
	}
}
