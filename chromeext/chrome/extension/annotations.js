import utf8 from "utf8";
import fetch from "../../app/actions/xhr";
import styles from "../../app/components/App.css";

// addAnnotations takes json annotation data returned from the
// Sourcegraph annotations API and manipulates the DOM to add
// hover-over tooltips and links.
//
// It assumes the caller has verified that the current view
// is "ready" to be annotated (e.g. DOM elements have all been rendered)
// and that there are no overlapping annotations in the json
// returned by the Sourcegraph API.
export default function addAnnotations(json) {
	if (document.getElementById("sourcegraph-annotation-marker")) {
		return;
	}

	let annsByStartByte = {};
	let annsByEndByte = {};
	for (let i = 0; i < json.Annotations.length; i++){
		if (json.Annotations[i].URL) {
			let ann = json.Annotations[i];
			annsByStartByte[ann.StartByte] = ann;
			annsByEndByte[ann.EndByte] = ann;
		}
	}
	traverseDOM(annsByStartByte, annsByEndByte);

	// Prevent double annotation on any file by adding some hidden
	// state to the page.
	const el = document.querySelector(".blob-wrapper");
	if (el) {
		const annotationMarker = document.createElement("div");
		annotationMarker.id = "sourcegraph-annotation-marker";
		annotationMarker.style.display = "none";
		el.appendChild(annotationMarker);
	}
}

let annotating = false; // HACK: private value indicating whether annotation is in progress for a single node (def)

// traverseDOM handles the actual DOM manipulation.
function traverseDOM(annsByStartByte, annsByEndByte){
	let table = document.querySelector("table");
	let count = 0;

	// get output HTML for each line and replace the original <td>
	for (let i = 0; i < table.rows.length; i++){
		let output = "";
		let row = table.rows[i];


		// Code is always the second <td> element; we want to replace code.innerhtml
		let code = row.cells[1]
		let children = code.childNodes;
		let startByte = count;
		count += utf8.encode(code.textContent).length;
		if (code.textContent !== "\n") {
			count++; // newline
		}
		// Go through each childNode
		for (let j = 0; j < children.length; j++) {
			let childNodeChars;
			let debug = false;

			if (children[j].nodeType === Node.TEXT_NODE){
				childNodeChars = children[j].nodeValue.split("");
			} else {
				const txt = document.createElement("textarea");
				// HACK: Quote marks for imported package were not getting linked
				// properly. The first mark was a separate anchor tag because GitHub
				// places quote marks in separate span tags. This hack makes it
				// such that if an element has childNodes, we merge the innerText
				// and set that as the innerHTML of the main span tag.
				if (children[j].children.length>0) {
					children[j].innerHTML = children[j].innerText
					txt.innerHTML = children[j].outerHTML
				}
				else {
					txt.innerHTML = children[j].outerHTML;
				}
				// childNodeChars = children[j].outerHTML.split("");
				childNodeChars = txt.value.split("");
			}

			// when we are returning the <span> element, we don"t want to increment startByte
			let consumingSpan = false;
			// keep track of whether we are currently on a definition with a link
			annotating = false;

			// go through each char of childNodes
			for (let k = 0; k < childNodeChars.length; k++) {
				if (childNodeChars[k] === "<" && (childNodeChars.slice(k, k+5).join("") === "<span" || childNodeChars.slice(k, k+6).join("") === "</span")) {
					consumingSpan = true;
				}

				if (!consumingSpan){
					output += next(childNodeChars[k], startByte, annsByStartByte, annsByEndByte)
					startByte += utf8.encode(childNodeChars[k]).length
				}
				else {
					output += childNodeChars[k]
				}

				if (childNodeChars[k] === ">" && consumingSpan) {
					consumingSpan = false;
				}
			}
		}

		// manipulate the DOM asynchronously so the page doesn't freeze while large
		// code files are being annotated
		setTimeout(() => {
			code.innerHTML = output;
			let newRows = code.childNodes
			for (let n = 0; n < newRows.length; n++) {
				addPopover(newRows[n]);
			}
		}, 0);
	}
}

// next is a helper method for traverseDOM
function next(c, byteCount, annsByStartByte, annsByEndByte) {
	let matchDetails = annsByStartByte[byteCount];
	c = `&#${c.charCodeAt(0)};`

	// if there is a match
	if (!annotating && matchDetails) {
		if (annsByStartByte[byteCount].EndByte - annsByStartByte[byteCount].StartByte === 1) {
			return `<a href="https://sourcegraph.com${matchDetails.URL}?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext" target="tab" class=${styles.sgdef}>${c}</a>`;
		}

		annotating = true;
		return `<a href="https://sourcegraph.com${matchDetails.URL}?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext" target="tab" class=${styles.sgdef}>${c}`;
	}

	// if we reach the end, close the tag.
	if (annotating && annsByEndByte[byteCount + 1]) {
		annotating = false;
		return `${c}</a>`;
	}

	return c;
}

let popoverCache = {};
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
			let url = activeTarget.href.split("https://sourcegraph.com")[1]
			url = url.split("?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext")[0]
			url = `https://sourcegraph.com/.api/repos${url}`;
			fetchPopoverData(url, function(html) {
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
		if (popoverCache[url]) return cb(popoverCache[url]);
		fetch(url)
			.then((json) => {
				let html;
				if (json.Data) {
					if (json.DocHTML){
						html = `<div><div class=${styles.popoverTitle}>${json.Kind} <b style="color:#4078C0">${json.Name}</b>${json.FmtStrings.Type.ScopeQualified}</div><div>${json.DocHTML.__html}</div><div class=${styles.popoverRepo}>${json.Repo}</div></div>`;
					} else {
						html = `<div><div class=${styles.popoverTitle}>${json.Kind} <b style="color:#4078C0">${json.Name}</b>${json.FmtStrings.Type.ScopeQualified}</div><div class=${styles.popoverRepo}>${json.Repo}</div></div>`;
					}
				}
				popoverCache[url] = html;
				cb(html);
			})
			.catch((err) => console.log("Error getting definition info.") && cb(null));
	}
}

function defQualifiedName(def) {
	if (!def || !def.FmtStrings) return "(unknown)";
	let f = def.FmtStrings;
	return `${escapeHTML(f.DefKeyword + " ")}<span style="font-weight:bold">${escapeHTML(f.Name.ScopeQualified)}</span>${escapeHTML(f.NameAndTypeSeparator + f.Type.ScopeQualified)}`;
}

const entityMap = {
	"&": "&amp;",
	"<": "&lt;",
	">": "&gt;",
	'"': "&quot;",
	"'": "&#39;",
	"/": "&#x2F;",
};

function escapeHTML(string) {
	return String(string).replace(/[&<>"'\/]/g, function (s) {
		return entityMap[s];
	});
};

