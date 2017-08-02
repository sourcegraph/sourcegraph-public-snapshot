import * as Rx from "rxjs";
import { fetchJumpURL, getTooltip } from "app/backend/lsp";
import * as tooltips from "app/tooltips/dom";
import { clearTooltip, setTooltip, store, TooltipContext } from "app/tooltips/store";
import { CodeCell, TooltipData } from "app/util/types";
import * as _ from "lodash";
import { events } from "app/tracking/events";

export interface RepoRevSpec { // TODO(john): move to types.
	repoURI: string;
	rev: string;
	isDelta: boolean;
	isBase: boolean;
}

// activeTarget tracks the element which is currently hovered over / clicked
let activeTarget: HTMLElement | null;

interface DOMObservables {
	mouseover: Rx.Observable<MouseEvent>;
	mouseout: Rx.Observable<MouseEvent>;
	mouseup: Rx.Observable<MouseEvent>;
}

/**
 * addEventListeners registers various event listeners on a DOM node and wraps each listener in an Observable.
 * It is idempotent and returns `undefined` if listeners have already been registered on the element.
 * @param el The element to register listeners on.
 * @param hoverCb Calback to execute whenever hovering over the element.
 */
function addEventListeners(el: HTMLElement, hoverCb: () => void): DOMObservables | undefined {
	// Ensure we only add listeners once per element.
	if (el.className.indexOf("sg-annotated") !== -1) {
		return;
	}
	el.className = `${el.className} sg-annotated`;

	const mouseover = Rx.Observable.fromEvent<MouseEvent>(el, "mouseover").do(hoverCb);
	const mouseout = Rx.Observable.fromEvent<MouseEvent>(el, "mouseout");
	const mouseup = Rx.Observable.fromEvent<MouseEvent>(el, "mouseup");
	return { mouseout, mouseover, mouseup };
}

/**
 * convertNode modifies a DOM node so that we can identify precisely token a user has clicked or hovered over.
 * On a code view, source code is typically wrapped in a HTML table cell. It may look like this:
 *
 *     <td id="LC18" class="blob-code blob-code-inner js-file-line">
 *        <#textnode>\t</#textnode>
 *        <span class="pl-k">return</span>
 *        <#textnode>&amp;Router{namedRoutes: </#textnode>
 *        <span class="pl-c1">make</span>
 *        <#textnode>(</#textnode>
 *        <span class="pl-k">map</span>
 *        <#textnode>[</#textnode>
 *        <span class="pl-k">string</span>
 *        <#textnode>]*Route), KeepContext: </#textnode>
 *        <span class="pl-c1">false</span>
 *        <#textnode>}</#textnode>
 *     </td>
 *
 * The browser extension works by registering a hover event listeners on the <td> element. When the user hovers over
 * "return" (in the first <span> node) the event target will be the <span> node. We can use the event target to determine which line
 * and which character offset on that line to use to fetch tooltip data. But when the user hovers over "Router"
 * (in the second text node) the event target will be the <td> node, which lacks the appropriate specificity to request
 * tooltip data. To circumvent this, all we need to do is wrap every free text node in a <span> tag.
 *
 * In summary, convertNode effectively does this: https://gist.github.com/lebbe/6464236
 *
 * There are three additional edge cases we handle:
 *   1. some text nodes contain multiple discrete code tokens, like the second text node in the example above; by wrapping
 *     that text node in a <span> we lose the ability to distinguish whether the user is hovering over "Router" or "namedRoutes".
 *   2. there may be arbitrary levels of <span> nesting; in the example above, every <span> node has only one (text node) child, but
 *     in reality a <span> node could have multiple children, both text and element nodes
 *   3. on GitHub diff views (e.g. pull requests) the table cell contains an additional prefix character ("+" or "-" or " ", representing
 *     additions, deletions, and unchanged code, respectively); we want to make sure we don't count that character when computing the
 *     character offset for the line
 *   4. TODO(john) some code hosts transform source code before rendering; in the example above, the first text node may be a tab character
 *     or multiple spaces
 *
 * @param parentNode The node to convert.
 */
function convertNode(parentNode: HTMLElement): void {
	for (let i = 0; i < parentNode.childNodes.length; ++i) {
		const node = parentNode.childNodes[i];
		const isLastNode = i === parentNode.childNodes.length - 1;
		if (node.nodeType === Node.TEXT_NODE) {
			let nodeText = _.unescape(node.textContent || "");
			if (nodeText === "") {
				continue;
			}
			parentNode.removeChild(node);
			let insertBefore = i;

			while (true) {
				const nextToken = consumeNextToken(nodeText);
				if (nextToken === "") {
					break;
				}
				const newTextNode = document.createTextNode(nextToken);
				const newTextNodeWrapper = document.createElement("SPAN");
				newTextNodeWrapper.appendChild(newTextNode);
				if (isLastNode) {
					parentNode.appendChild(newTextNodeWrapper);
				} else {
					// increment insertBefore as new span-wrapped text nodes are added
					parentNode.insertBefore(newTextNodeWrapper, parentNode.childNodes[insertBefore++]);
				}
				nodeText = nodeText.substr(nextToken.length);
			}
		} else if (node.nodeType === Node.ELEMENT_NODE) {
			const elementNode = node as HTMLElement;
			// if (elementNode.children.length > 0) {
			// The element is something more complicated than <span>text</span>; recurse.
			convertNode(elementNode);
			// }
		}
	}
}

const VARIABLE_TOKENIZER = /(^\w+)/;
const ASCII_CHARACTER_TOKENIZER = /(^[\x21-\x2F|\x3A-\x40|\x5B-\x60|\x7B-\x7E])/;
const NONVARIABLE_TOKENIZER = /(^[^\x21-\x7E]+)/;

/**
 * consumeNextToken parses the text content of a text node and returns the next "distinct"
 * code token. It handles edge case #1 from convertNode(). The tokenization scheme is
 * heuristic-based and uses simple regular expressions.
 * @param txt Aribitrary text to tokenize.
 */
function consumeNextToken(txt: string): string {
	if (txt.length === 0) {
		return "";
	}

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

enum TooltipEventType {
	HOVER,
	CLICK,
	SELECT_TEXT,
}

/**
 * getTooltipObservable wraps the asynchronous "fetch" of tooltip data from the Sourcegraph API.
 * This Observable will emit exactly one value before it completes.
 * @param target The element tooltip is being requested for.
 * @param context The parameters for fetching tooltip data for `target`.
 */
function getTooltipObservable(target: HTMLElement, context: TooltipContext): Rx.Observable<{ target: HTMLElement, data: TooltipData }> {
	if (!context.coords) {
		throw new Error("cannot get tooltip without line/char");
	}
	return Rx.Observable.fromPromise(getTooltip(context.path, context.coords.line, context.coords.char, context.repoRevSpec))
		.do(data => {
			if (data && data.title) {
				// If non-empty tooltip data is returned, make the target "clickable" (via cursor pointer styling)
				target.style.cursor = "pointer";
				// TODO(john): remove sg-clickable styling if possible.
				if (!target.className.includes("sg-clickable")) {
					target.className = `${target.className} sg-clickable`;
				}
			}
		})
		.map(data => ({ target, data }));
}
/**
 * getTooltipObservable wraps the asynchronous "fetch" of tooltip data from the Sourcegraph API.
 * This Observable will emit exactly one value before it completes.
 * @param target The element tooltip is being requested for.
 * @param context The parameters for fetching tooltip data for `target`.
 */
function getJ2DObservable(context: TooltipContext): Rx.Observable<string | null> {
	if (!context.coords) {
		throw new Error("cannot get j2d without line/char");
	}
	return Rx.Observable.fromPromise(fetchJumpURL(context.coords.char, context.path, context.coords.line, context.repoRevSpec));
}

/**
 * getLoadingTooltipObervable emits "loading" tooltip data after a delay, iff another Observable hasn't already emitted a value.
 * @param target The element tooltip is being requested for.
 * @param tooltipObservable An Observable used to short-circuit emitting a "loading" tooltip (likely the Observable which wraps the asynchronous "fetch" of tooltip data from the Sourcegraph API).
 */
function getLoadingTooltipObservable(target: HTMLElement, tooltipObservable: Rx.Observable<any>): Rx.Observable<{ target: HTMLElement, data: TooltipData }> {
	// Show a loading tooltip after .5 seconds, ONLY if we haven't already gotten real tooltip data
	return Rx.Observable.fromPromise<TooltipData>(new Promise((resolve) => {
		setTimeout(() => resolve({ loading: true }), 500);
	}))
		.takeUntil(tooltipObservable) // short-circuit once actual tooltip is fetched
		.map(data => ({ target, data }));
}

/**
 * tooltipEvent is the "observer" for tooltip Observables. Pass it tooltip data (emitted from an Observable) and it will update the tooltip store.
 * @param ev The observable tooltip data event.
 * @param context Parameters used to fetch the tooltip data.
 * @param type Which type of user interaction caused the event.
 */
function tooltipEvent(ev: { target: HTMLElement, data: TooltipData }, context: TooltipContext, type: TooltipEventType, logEvent: boolean = true): void {
	if (store.getValue().docked && type === TooltipEventType.HOVER) {
		// While a tooltip is docked, hovers should not update the active tooltip.
		return;
	}
	if (ev.target === activeTarget && logEvent) {
		if (ev.data.title || ev.data.j2dUrl) {
			switch (type) {
				case TooltipEventType.HOVER:
					events.SymbolHovered.log();
					break;
				case TooltipEventType.CLICK:
					events.TooltipDocked.log();
					break;
				case TooltipEventType.SELECT_TEXT:
					events.TextSelected.log();
					break;

			}
		}
		setTooltip({
			target: ev.target,
			data: ev.data,
			docked: type === TooltipEventType.CLICK || type === TooltipEventType.SELECT_TEXT,
			context,
		});
	}
}

// TODO(john): add back this special-casing for spaces (consult Beyang)
// var allSpaces = true;
// while (nodeText.length > 0) {
// 	const token = consumeNextToken(nodeText);
// 	const isAllSpaces = SPACES.test(token);
// 	allSpaces = isAllSpaces && allSpaces;

// 	wrapperNode.appendChild(createTextNode(token, offset + prevConsumed));
// 	prevConsumed += isAllSpaces && spacesToTab > 0 && token.length % spacesToTab === 0 ? token.length / spacesToTab : token.length;
// 	bytesConsumed += isAllSpaces && spacesToTab > 0 && token.length % spacesToTab === 0 ? token.length / spacesToTab : token.length;
// 	if (!allSpaces && spacesToTab > 0) {
// 		// NOTE: this makes it so that if there are further spaces, they don't get divided by 2 for their byte offset.
// 		// only divide by 2 for initial code indents.
// 		spacesToTab = 0;
// 	}
// 	nodeText = nodeText.slice(token.length);
// }

/**
 * addAnnotations is the entry point for marking up a DOM element with source code in it.
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
export function addAnnotations(path: string, repoRevSpec: RepoRevSpec, cells: CodeCell[]): void {
	tooltips.createTooltips(); // TODO(john): can we just do this once in the module)?
	const ignoreFirstChar = repoRevSpec.isDelta;

	// TODO(john): figure out how to do this without looking at the cell itself.
	// if ((cell as PhabricatorCodeCell).isLeftColumnInSplit || (cell as PhabricatorCodeCell).isUnified) {
	// 	ignoreFirstTextChar = false;
	// }

	const domObservables = _.compact(cells.map((cell) => {
		let hovered = false;
		const hoverCb = () => {
			if (hovered) {
				return;
			}
			hovered = true;
			convertNode(cell.cell);
		};
		cell.cell.setAttribute("data-sg-line-number", `${cell.line}`);
		return addEventListeners(cell.eventHandler, hoverCb);
	})) as DOMObservables[];

	domObservables.forEach(observable => observable.mouseover.subscribe((e) => {
		const t = e.target as HTMLElement;
		if (!store.getValue().docked) {
			activeTarget = t;
		}
		const coords = getTargetLineAndOffset(e.target as HTMLElement, ignoreFirstChar);
		if (!coords) {
			return;
		}
		const context = { path, repoRevSpec, coords: coords! };
		const tooltipObservable = getTooltipObservable(t, context);
		const loadingTooltipObservable = getLoadingTooltipObservable(t, tooltipObservable);
		tooltipObservable.subscribe((ev) => tooltipEvent(ev, context, TooltipEventType.HOVER));
		tooltipObservable.zip(getJ2DObservable(context)).subscribe((ev) => {
			if (ev[1]) {
				ev[0].data.j2dUrl = ev[1]!;
				tooltipEvent(ev[0], context, TooltipEventType.HOVER, false);
			}
		});
		loadingTooltipObservable.subscribe((ev) => tooltipEvent(ev, context, TooltipEventType.HOVER));
	}));

	let lastSelectedText: string = "";

	domObservables.forEach(observable => observable.mouseout.subscribe((() => {
		if (!store.getValue().docked) {
			activeTarget = null;
			clearTooltip();
		}
	})));

	domObservables.forEach(observable => observable.mouseup.subscribe((e) => {
		const t = e.target as HTMLElement;
		const clickedActiveTarget = t === activeTarget;
		activeTarget = t;

		if (lastSelectedText !== "") {
			clearTooltip(t);
		}

		const selectedText = getSelectedText();
		if (selectedText !== "") {
			const shortCircuitTooltip = lastSelectedText === selectedText;
			lastSelectedText = selectedText;

			if (!shortCircuitTooltip) {
				const target = getSelectedTextTarget();
				activeTarget = target;

				const context = { path, repoRevSpec, selectedText };
				tooltipEvent({ target, data: { title: selectedText } }, context, TooltipEventType.SELECT_TEXT);
				return;
			} else {
				lastSelectedText = "";
			}
		} else {
			lastSelectedText = "";
		}

		const coords = getTargetLineAndOffset(e.target as HTMLElement, ignoreFirstChar);
		if (!coords) {
			clearTooltip(t);
			return;
		}
		if (store.getValue().docked && clickedActiveTarget) {
			setTooltip({ target: t, docked: false });
			return;
		}
		const context = { path, repoRevSpec, coords: coords! };
		const tooltipObservable = getTooltipObservable(t, context);
		const loadingTooltipObservable = getLoadingTooltipObservable(t, tooltipObservable);
		tooltipObservable.subscribe((ev) => {
			if (ev.data.loading || !ev.data.title) {
				clearTooltip(t);
				return;
			}
			tooltipEvent(ev, context, TooltipEventType.CLICK);
		});
		tooltipObservable.zip(getJ2DObservable(context)).subscribe((ev) => {
			if (ev[1]) {
				ev[0].data.j2dUrl = ev[1]!;
				tooltipEvent(ev[0], context, TooltipEventType.CLICK, false);
			}
		});
		loadingTooltipObservable.subscribe((ev) => tooltipEvent(ev, context, TooltipEventType.CLICK));
	}));
}

/**
 * getTargetLineAndOffset determines the line and character offset for some source code, identified by its HTMLElement wrapper.
 * It works by traversing the DOM until the HTMLElement's ancestor with a "data-sg-line-number" attribute is found, and short-circuits
 * when a <td> ancestor is found. (We expect all "data-sg-line-number"-annotated nodes to be nested within some <td> tag.) Once the
 * ancestor is found, we traverse the DOM again (this time the opposite direction) counting characters until the original target is found.
 * Returns undefined if line/char cannot be determined for the provided target.
 * @param target The element to compute line & character offset for.
 * @param ignoreFirstChar Whether to ignore the first character on a line when computing character offset.
 */
function getTargetLineAndOffset(target: HTMLElement, ignoreFirstChar: boolean = false): { line: number, char: number, word: string } | undefined {
	const origTarget = target;
	if (target.tagName === "TD") {
		// Short-circuit; we are hovering over a line of code, but no token in particular.
		return;
	}
	while (target && target.tagName !== "TD" && target.tagName !== "BODY" && !target.getAttribute("data-sg-line-number")) {
		// Find ancestor which wraps the whole line of code, not just the target token.
		target = (target.parentNode as HTMLElement);
	}
	if (!target.getAttribute("data-sg-line-number")) {
		// Make sure we're looking at an element we've annotated line number for (otherwise we have no idea )
		return;
	}
	const line = parseInt(target.getAttribute("data-sg-line-number")!, 10);

	let char = 1;
	// Iterate recursively over the current target's children until we find the original target;
	// count characters along the way. Return true if the original target is found.
	function findOrigTarget(root: HTMLElement): boolean {
		// tslint:disable-next-line
		for (let i = 0; i < root.childNodes.length; ++i) {
			const child = root.childNodes[i] as HTMLElement;
			if (child === origTarget) {
				return true;
			}
			if (child.children === undefined) {
				char += child.textContent!.length;
				continue;
			}
			if (child.children.length > 0 && findOrigTarget(child)) {
				// Walk over nested children, then short-circuit the loop to avoid double counting children.
				return true;
			}
			if (child.children.length === 0) {
				// Child is not the original target, but has no chidren to recurse on. Add to character offset.
				char += (child.textContent as string).length; // TODO(john): I think this needs to be escaped before we add its length...
				if (ignoreFirstChar) {
					char -= 1; // make sure not to count weird diff prefix
					ignoreFirstChar = false;
				}
			}
		}
		return false;
	}
	// Start recursion.
	if (findOrigTarget(target)) {
		return { line, char, word: origTarget.innerText };
	}
}

function getSelectedText(): string {
	let text = "";
	if (typeof window.getSelection !== "undefined") {
		text = window.getSelection().toString();
	} else if (typeof (document as any).selection !== "undefined" && (document as any).selection.type === "Text") {
		text = (document as any).selection.createRange().text;
	}
	return text;
}

function getSelectedTextTarget(): HTMLElement {
	return window.getSelection().getRangeAt(0).startContainer.parentElement!;
}
