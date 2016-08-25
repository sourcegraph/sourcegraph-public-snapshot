import {Location} from "history";
import * as debounce from "lodash/debounce";
import * as React from "react";
import * as ReactDOM from "react-dom";
import {Component, EventListener} from "sourcegraph/Component";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {Props} from "sourcegraph/blob/Blob";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import {BlobLine} from "sourcegraph/blob/BlobLine";
import {BlobLineExpander, Range} from "sourcegraph/blob/BlobLineExpander";
import {annotationsByLine} from "sourcegraph/blob/annotationsByLine";
import {computeLineStartBytes, lineFromByte} from "sourcegraph/blob/lineFromByte";
import * as s from "sourcegraph/blob/styles/Blob.css";
import {withJumpToDefRedirect} from "sourcegraph/blob/withJumpToDefRedirect";
import {Def} from "sourcegraph/def/index";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {fileLines} from "sourcegraph/util/fileLines";

interface State {
	startlineCallback: any;
	location: Location;
	repo: any;
	rev: any;
	commitID: any;
	path: any;
	lines: any;
	highlightSelectedLines: boolean;
	highlightedDef: string | null;
	highlightedDefObj: Def | null;
	startLine: any;
	startCol: number | null;
	startByte: number | null;
	endLine: number | null;
	endCol: number | null;
	endByte: number | null;
	lineStartBytes: any;
	lineNumbers: boolean;
	activeDef: string | null;
	activeDefRepo: string | null;
	contentsOffsetLine: number;
	expandedRanges: Range[];
	displayLineExpanders: string | null;
	textSize: string;
	scrollTarget: number | null;
	scrollCallback: any;
	lineAnns: any;
	scrollToStartLine: any;
	displayRanges: any;
	dispatchSelections: any;
	annotations: any;
	contents: any;
};

// BlobTestOnly should only be used on its own for testing purposes. Normally, 
// you should be using Blob that's at the bottom of this file. 
export class BlobLegacyTestOnly extends Component<Props, State> {

	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
	};

	_isMounted: boolean;

	state: State = {
		startlineCallback: x => { /* empty */ },
		location: {
			key: "",
			pathname: "",
			search: "",
			action: "",
			query: {},
			state: {},
		},
		repo: "",
		rev: "",
		commitID: "",
		path: "",
		activeDef: null,
		activeDefRepo: null,
		highlightSelectedLines: false,
		highlightedDef: null,
		highlightedDefObj: null,
		contentsOffsetLine: 0,
		lineAnns: [],
		lineStartBytes: [],
		lines: [],
		lineNumbers: false,
		expandedRanges: [],
		startLine: null,
		startCol: null,
		startByte: null,
		endLine: null,
		endCol: null,
		endByte: null,
		displayLineExpanders: null,
		textSize: "normal",
		scrollTarget: null,
		scrollCallback: null,
		scrollToStartLine: null,
		displayRanges: null,
		dispatchSelections: null,
		annotations: null,
		contents: null,
	};

	constructor(props: Props) {
		super(props);
		this._expandRange = this._expandRange.bind(this);
		this._handleSelectionChange = debounce(this._handleSelectionChange.bind(this), 200, {
			leading: false,
			trailing: true,
		});
		this._isMounted = false;
	}

	componentDidMount(): void {
		// TODO: This is hacky, but the alternative was too costly (time and code volume)
		//       and unreliable to implement. Revisit this later if it's neccessary.
		//
		// Delay scrolling to give BlobRouter a chance to populate startLine.
		setTimeout(() => {
			if (this.state.startLine && this.state.scrollToStartLine) {
				this._scrollTo(this.state.startLine);
			}
		}, 0);
		this._isMounted = true;
	}

	componentWillUnmount(): void {
		this._isMounted = false;
	}

	reconcileState(state: State, props: Props): void {
		state.startlineCallback = props.startlineCallback || (x => { /* empty */ });
		state.repo = props.repo || null;
		state.rev = props.rev || null;
		state.commitID = props.commitID || null;
		state.path = props.path || null;

		let oldStartLine = state.startLine;
		state.startLine = props.startLine || null;

		state.startCol = props.startCol || null;
		state.startByte = props.startByte || null;
		state.endLine = props.endLine || null;
		state.endCol = props.endCol || null;
		state.endByte = props.endByte || null;
		state.textSize = props.textSize || "normal";
		state.scrollToStartLine = Boolean(props.scrollToStartLine);
		state.contentsOffsetLine = props.contentsOffsetLine || 0;
		state.lineNumbers = Boolean(props.lineNumbers);
		state.highlightedDef = props.highlightedDef || null;
		state.highlightedDefObj = props.highlightedDefObj || null;
		state.activeDef = props.activeDef || null;
		state.activeDefRepo = props.activeDefRepo || null;
		state.highlightSelectedLines = Boolean(props.highlightSelectedLines);
		state.dispatchSelections = Boolean(props.dispatchSelections);
		state.displayRanges = props.displayRanges ? this._consolidateRanges(props.displayRanges.concat(state.expandedRanges)) : null;
		state.displayLineExpanders = props.displayLineExpanders || null;

		let updateAnns = false;

		if (state.annotations !== props.annotations && !props.skipAnns) {
			state.annotations = props.annotations || null;
			updateAnns = true;
		}

		if (state.contents !== props.contents) {
			state.contents = null;
			state.lines = null;
			state.lineStartBytes = null;
			updateAnns = true;
			if (props.contents) {
				state.contents = props.contents;
				state.lines = fileLines(props.contents);
				state.lineStartBytes = state.annotations && state.annotations.LineStartBytes ? state.annotations.LineStartBytes : computeLineStartBytes(state.lines);
			}
		}

		if (updateAnns) {
			state.lineAnns = state.lineStartBytes && state.annotations && state.annotations.Annotations ? annotationsByLine(state.lineStartBytes, state.annotations.Annotations, state.lines) : null;
		}

		if (state.contents && state.startByte && state.endByte && state.startLine === null && state.endLine === null) {
			state.startLine = lineFromByte(state.lines, state.startByte);
			state.endLine = lineFromByte(state.lines, state.endByte);
		}
		if (state.startLine && oldStartLine !== state.startLine) {
			let scrollCallback = () => {
				if (this.state.scrollToStartLine) {
					this._scrollTo(state.startLine);
				}
			};
			let isVisible = this._lineIsVisible(state.startLine);

			// In the case when the BlobLine that we want to scroll to hasn't loaded yet,
			// particularly when navigating back to a blob page using the "back" button,
			// we register a callback with the BlobLine that it invokes when it has finished loading.
			if (isVisible === null) {
				state.scrollTarget = state.startLine;
				state.scrollCallback = scrollCallback;
			} else if (isVisible === false) {
				scrollCallback();
			}
		}
	}

	_consolidateRanges(ranges: Range[]): Range[] | null {
		if (ranges.length === 0) {
			return null;
		}
		ranges = ranges.sort((a, b) => {
			if (a[0] < b[0]) {
				return -1;
			}
			if (a[0] > b[0]) {
				return 1;
			}
			return 0;
		});

		let newRanges = [ranges[0]];
		for (let range of ranges) {
			let lastRange = newRanges[newRanges.length - 1];
			if (lastRange[1] < range[0]) {
				newRanges.push(range);
			} else if (lastRange[1] < range[1]) {
				// If the ranges overlap and range's ending value is greater, update
				// the ending value of lastRange.
				lastRange[1] = range[1];
			}
		}
		return newRanges;
	}

	_withinDisplayedRange(lineNumber: number): boolean {
		if (!this.state.displayRanges) {
			return false;
		}
		for (let range of this.state.displayRanges) {
			if (range[0] <= lineNumber && lineNumber <= range[1]) {
				return true;
			}
		}
		return false;
	}

	_expandRange(range: Range): void {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DEF_INFO, AnalyticsConstants.ACTION_CLICK, "BlobLineExpanderClicked", {repo: this.state.repo, active_def: this.state.activeDef, path: this.state.path});
		this.setState({
			expandedRanges: this.state.expandedRanges.concat([range]),
		} as State);
	}

	_handleSelectionChange(ev: Event): void {
		if (!this.state.dispatchSelections) {
			return;
		}

		const sel = document.getSelection();
		if (!sel || sel.rangeCount === 0) {
			return;
		}
		const rng = sel.getRangeAt(0);

		let getLineElem: (node: Node | null) => HTMLElement | null;
		getLineElem = (node) => {
			if (!node) {
				return null;
			}
			if (node instanceof HTMLElement && node.nodeName === "TR" && node.dataset["line"]) {
				return node;
			}
			return getLineElem(node.parentNode);
		};

		const getColInLine = (lineElem, containerElem, offsetInContainer: number) => {
			let lineContentElem = lineElem.lastChild;
			let col = 0;
			let q: any[] = [lineContentElem];
			while (q.length > 0) {
				let e = q.pop();
				if (e === containerElem) {
					return col + offsetInContainer;
				}
				if (e.nodeType === Node.TEXT_NODE) {
					col += e.textContent.length;
				}
				if (e.childNodes) {
					for (let i = e.childNodes.length - 1; i >= 0; i--) {
						q.push(e.childNodes[i]);
					}
				}
			}
			return 0;
		};

		let startLineElem = getLineElem(rng.startContainer);
		let endLineElem = getLineElem(rng.endContainer);

		// Don't let selections OUTSIDE this file view affect the selected range.
		// But if one extent (start or end) is in, then treat the other extent as
		// occurring at the extreme (start or end).
		if (!startLineElem && !endLineElem) {
			return;
		}
		const $el = ReactDOM.findDOMNode(this);
		if (startLineElem && !$el.contains(startLineElem)) {
			return;
		}
		if (endLineElem && !$el.contains(endLineElem)) {
			return;
		}

		if (sel.isCollapsed) {
			// It's a click IN the file view with an empty (collapsed) selection.
			Dispatcher.Stores.dispatch(new BlobActions.SelectCharRange(this.state.repo, this.state.rev, this.state.path, null, null, null, null, null, null));
			return;
		}

		// Get start/end line + col. If the line elem isn't found, then the selection
		// extends beyond the bounds.
		let startLine = startLineElem ? parseInt(startLineElem.dataset["line"], 10) : 1;
		let startCol = startLineElem ? getColInLine(startLineElem, rng.startContainer, rng.startOffset) : 0;
		let endLine = endLineElem ? parseInt(endLineElem.dataset["line"], 10) : this.state.lines.length;
		let endCol = endLineElem ? getColInLine(endLineElem, rng.endContainer, rng.endOffset) : this.state.lines[this.state.lines.length - 1].length;

		let startByte = this.state.lineStartBytes[startLine - 1] + startCol;
		let endByte = this.state.lineStartBytes[endLine - 1] + endCol;

		Dispatcher.Stores.dispatch(new BlobActions.SelectCharRange(this.state.repo, this.state.rev, this.state.path, startLine, startCol, startByte, endLine, endCol, endByte));
	}

	_scrollTo(line: number): void {
		if (!this.refs["table"]) { return; }
		let rect = (this.refs["table"] as Element).getBoundingClientRect();
		const y = rect.height / this.state.lines.length * (line - 1) - 100;
		((document as any).getElementById("scroller") as any).scrollTop = y;
	}

	// _lineIsVisible returns true the line has loaded and is scrolled into view, false
	// if the line has loaded and isn't scrolled into view, or null if the line hasn't loaded yet.
	_lineIsVisible(line: number): boolean | null {
		if (!this._isMounted) {
			return false;
		}
		if (typeof document === "undefined" || typeof window === "undefined") {
			return false;
		}
		let top = document.body.scrollTop;
		let bottom = top + window.innerHeight;
		let $line: any = (this.refs["table"] as Element).querySelector(`[data-line="${line}"]`);
		if (!$line) {
			return null;
		}
		let elemTop = $line.offsetTop;
		let elemBottom = elemTop + $line.clientHeight;
		const deadZone = 150; // consider things not visible if they are near the screen top/bottom and hard to notice
		return elemBottom <= (bottom - deadZone) && elemTop >= (top + deadZone);
	}

	getOffsetTopForByte(byte: number): number {
		if (!this._isMounted) {
			return 0;
		}
		if (typeof byte === "undefined") {
			throw new Error("getOffsetTopForByte byte is undefined");
		}
		let $el = ReactDOM.findDOMNode(this);
		let line = lineFromByte(this.state.lines, byte);
		let $line: any = $el.querySelector(`[data-line="${line}"]`);
		let $toolbar: any = $el.parentNode.childNodes[0];
		if ($line) {
			return $line.offsetTop + $toolbar.clientHeight + $toolbar.offsetTop;
		}
		throw new Error(`No element found for line ${line}`);
	}

	render(): JSX.Element | null {
		if (!this.state.lines) {
			return null;
		}
		let lastDisplayedLine = 0;
		let lastRangeEnd = 0;
		let lines: any[] = [];
		let renderedLines: number = 0;
		this.state.lines.forEach((line, i: number) => {
			const lineNumber = 1 + i + this.state.contentsOffsetLine;
			if (this.state.displayRanges && !this._withinDisplayedRange(lineNumber)) {
				return;
			}
			if (this.state.displayRanges && lastDisplayedLine !== lineNumber - 1 && (this.state.displayLineExpanders === null || this.state.displayLineExpanders !== "bottom")) {
				// Prevent expanding above the last displayed range.
				let expandTo = [Math.max(lastRangeEnd, lineNumber - 30), lineNumber - 1];
				lines.push(
					<BlobLineExpander key={`expand-${i}`}
						direction={renderedLines === 0 ? "up" : null}
						expandRange={expandTo}
						onExpand={this._expandRange} />
				);
				lastRangeEnd = lineNumber;
			}
			lastDisplayedLine = lineNumber;

			lines.push(
				<BlobLine
					onMount={this.state.scrollTarget && this.state.scrollTarget === i && this.state.scrollCallback ? this.state.scrollCallback : null}
					location={this.props.location}
					ref={this.state.startLine === lineNumber ? this.state.startlineCallback : undefined}
					repo={this.state.repo}
					rev={this.state.rev}
					commitID={this.state.commitID}
					path={this.state.path}
					lineNumber={lineNumber}
					showLineNumber={this.state.lineNumbers}
					startByte={this.state.lineStartBytes[i]}
					contents={line}
					textSize={this.state.textSize}
					annotations={this.state.lineAnns ? (this.state.lineAnns[i] || null) : null}
					selected={this.state.highlightSelectedLines && this.state.startLine !== null && this.state.endLine !== null && this.state.startLine <= lineNumber && this.state.endLine >= lineNumber}
					highlightedDef={this.state.highlightedDef}
					highlightedDefObj={this.state.highlightedDefObj}
					activeDef={this.state.activeDef}
					activeDefRepo={this.state.activeDefRepo}
					key={i} />
			);
			renderedLines += 1;
		});
		if (this.state.lines && lastDisplayedLine < this.state.lines.length && (this.state.displayLineExpanders === null || this.state.displayLineExpanders !== "top")) {
			lines.push(
				<BlobLineExpander key={`expand-${this.state.lines.length}`}
					expandRange={[lastDisplayedLine, lastDisplayedLine + 30]}
					onExpand={this._expandRange}
					direction={"down"}/>
			);
		}
		return (
			<div className={s.scroller}>
				<table className={s.lines} ref="table">
					<tbody>
						{lines}
					</tbody>
				</table>
				<EventListener target={global.document} event="selectionchange" callback={this._handleSelectionChange} />
			</div>
		);
	}
}

let blob = withJumpToDefRedirect(BlobLegacyTestOnly);
export {blob as BlobLegacy};
