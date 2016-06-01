import utf8 from "utf8";
import fetch from "../../app/actions/xhr";
import styles from "../../app/components/App.css";
import _ from "lodash";

// addAnnotations takes json annotation data returned from the
// Sourcegraph annotations API and manipulates the DOM to add
// hover-over tooltips and links.
//
// It assumes the caller has verified that the current view
// is "ready" to be annotated (e.g. DOM elements have all been rendered)
// and that there are no overlapping annotations in the json
// returned by the Sourcegraph API.
//
// It assumes that the formatted html provided by the Sourcegraph API
// for doc tooltips is "safe" to be injected into the page.
//
// It does *not* assume that the code that is being annotated is safe
// to be executed in our script, so we take care to properly escape
// characters during the annotation loop.
export default function addAnnotations(json) {
	if (document.getElementById("sourcegraph-annotation-marker")) {
		// This function is not idempotent; don't let it run twice.
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

let annotating = false; // imperative private value indicating whether annotation is in progress for a single token (def)

// traverseDOM handles the actual DOM manipulation.
function traverseDOM(annsByStartByte, annsByEndByte){
	let table = document.querySelector("table");
	let count = 0;

	// get output HTML for each line and replace the original <td>
	for (let i = 0; i < table.rows.length; i++){
		let output = "";
		let row = table.rows[i];

		// Code is always the second <td> element; we want to replace code.innerhtml
		// with a Sourcegraph-"linkified" version of the token, or the same token
		let code = row.cells[1]
		let children = code.childNodes;
		let startByte = count;
		count += utf8.encode(code.textContent).length;
		if (code.textContent !== "\n") {
			count++; // newline
		}

		for (let j = 0; j < children.length; j++) {
			let childNodeChars; // the "inner-stuff" of the code cell

			if (children[j].nodeType === Node.TEXT_NODE){
				childNodeChars = children[j].nodeValue.split("");
			} else {
				if (children[j].children.length > 1) {
					// HACK: combine children spans, e.g. for a quoted token
					// there may be 3 spans, one for each quote and one
					// for the token iteself
					children[j].innerHTML = _.escape(children[j].innerText);
				}
				childNodeChars = _.unescape(children[j].outerHTML).split("");
			}

			let consumingSpan = false;
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
					// when we are consuming the <span> element, don't increment startByte
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
		});
	}
}

// next is a helper method for traverseDOM which transforms a character
// into itself or wraps the character in a starting/ending anchor tag
function next(c, byteCount, annsByStartByte, annsByEndByte) {
	let matchDetails = annsByStartByte[byteCount];

	c = _.escape(c); // IMPORTANT: escape all markup injected in HTML

	// if there is a match
	if (!annotating && matchDetails) {
		// off-by-one case
		if (annsByStartByte[byteCount].EndByte - annsByStartByte[byteCount].StartByte === 1) {
			return `<a href="https://sourcegraph.com${matchDetails.URL}?utm_source=browser-ext&browser_type=chrome" target="tab" class=${styles.sgdef}>${c}</a>`;
		}

		annotating = true;
		return `<a href="https://sourcegraph.com${matchDetails.URL}?utm_source=browser-ext&browser_type=chrome" target="tab" class=${styles.sgdef}>${c}`;
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
			url = url.split("?utm_source=browser-ext&browser_type=chrome")[0]
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
