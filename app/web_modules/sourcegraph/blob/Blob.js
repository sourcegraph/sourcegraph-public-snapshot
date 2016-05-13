// @flow

import React from "react";
import ReactDOM from "react-dom";

import Component from "sourcegraph/Component";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import BlobLine from "sourcegraph/blob/BlobLine";
import BlobLineExpander from "sourcegraph/blob/BlobLineExpander";
import type {Range} from "sourcegraph/blob/BlobLineExpander";
import Dispatcher from "sourcegraph/Dispatcher";
import debounce from "lodash/function/debounce";
import fileLines from "sourcegraph/util/fileLines";
import lineFromByte from "sourcegraph/blob/lineFromByte";
import {computeLineStartBytes} from "sourcegraph/blob/lineFromByte";
import annotationsByLine from "sourcegraph/blob/annotationsByLine";
import s from "sourcegraph/blob/styles/Blob.css";
import type {Def} from "sourcegraph/def";

class Blob extends Component {
	static propTypes = {
		contents: React.PropTypes.string,
		annotations: React.PropTypes.shape({
			Annotations: React.PropTypes.array,
			LineStartBytes: React.PropTypes.array,
		}),
		lineNumbers: React.PropTypes.bool,
		skipAnns: React.PropTypes.bool,
		startLine: React.PropTypes.number,
		startCol: React.PropTypes.number,
		startByte: React.PropTypes.number,
		endLine: React.PropTypes.number,
		endCol: React.PropTypes.number,
		endByte: React.PropTypes.number,
		scrollToStartLine: React.PropTypes.bool,
		highlightedDef: React.PropTypes.string,
		highlightedDefObj: React.PropTypes.object,
		activeDef: React.PropTypes.string,
		activeDefNoRev: React.PropTypes.string,

		// For linking line numbers to the file they came from (e.g., in
		// ref snippets).
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string.isRequired,
		path: React.PropTypes.string.isRequired,

		// contentsOffsetLine indicates that the contents string does not
		// start at line 1 within the file, but rather some other line number.
		// It must be specified when startLine > 1 but the contents don't begin at
		// the first line of the file.
		contentsOffsetLine: React.PropTypes.number,

		highlightSelectedLines: React.PropTypes.bool,

		// dispatchSelections is whether this Blob should emit BlobActions.SelectCharRange
		// actions when the text selection changes. It should be true for the main file view but
		// not for secondary file views (e.g., usage examples).
		dispatchSelections: React.PropTypes.bool,
	};

	constructor(props: Blob.props) {
		super(props);
		this._expandRange = this._expandRange.bind(this);
		this._handleSelectionChange = debounce(this._handleSelectionChange.bind(this), 200, {
			leading: false,
			trailing: true,
		});
		this._updateVisibleLines = debounce(this._updateVisibleLines.bind(this), 200, {
			leading: false,
			trailing: true,
		});
		this._isMounted = false;
	}

	state: {
		repo: string;
		rev: string;
		path: string;
		lines: string[];
		highlightSelectedLines: boolean;
		highlightedDef: ?string;
		highlightedDefObj: ?Def;
		startLine: ?number;
		startCol: ?number;
		startByte: ?number;
		endLine: ?number;
		endCol: ?number;
		endByte: ?number;
		lineStartBytes: number[];
		activeDef: ?string;
		activeDefNoRev: ?string;
		highlightedDef: ?string;
		contentsOffsetLine: number;
		expandedRanges: Range[];
		visStartLine: number;
		visEndLine: number;
	} = {
		repo: "",
		rev: "",
		path: "",
		activeDef: null,
		activeDefNoRev: null,
		highlightSelectedLines: false,
		highlightedDef: null,
		highlightedDefObj: null,
		contentsOffsetLine: 0,
		lineAnns: [],
		lineStartBytes: [],
		lines: [],
		expandedRanges: [],
		startLine: null,
		startCol: null,
		startByte: null,
		endLine: null,
		endCol: null,
		endByte: null,
		visStartLine: 0,
		visEndLine: 0,
	};

	componentDidMount() {
		// TODO: This is hacky, but the alternative was too costly (time and code volume)
		//       and unreliable to implement. Revisit this later if it's neccessary.
		//
		// Delay scrolling to give BlobRouter a chance to populate startLine.
		setTimeout(() => {
			if (this.state.startLine && this.state.scrollToStartLine) {
				this._scrollTo(this.state.startLine);
			}
		}, 0);

		document.addEventListener("selectionchange", this._handleSelectionChange);
		document.addEventListener("scroll", this._updateVisibleLines);
		this._isMounted = true;
		this._updateVisibleLines();
	}

	componentWillUnmount() {
		document.removeEventListener("selectionchange", this._handleSelectionChange);
		document.removeEventListener("scroll", this._updateVisibleLines);
		this._isMounted = false;
	}

	reconcileState(state: Blob.state, props: Blob.props) {
		state.repo = props.repo || null;
		state.rev = props.rev || null;
		state.path = props.path || null;
		state.startLine = props.startLine || null;
		state.startCol = props.startCol || null;
		state.startByte = props.startByte || null;
		state.endLine = props.endLine || null;
		state.endCol = props.endCol || null;
		state.endByte = props.endByte || null;
		state.scrollToStartLine = Boolean(props.scrollToStartLine);
		state.contentsOffsetLine = props.contentsOffsetLine || 0;
		state.lineNumbers = Boolean(props.lineNumbers);
		state.highlightedDef = props.highlightedDef || null;
		state.highlightedDefObj = props.highlightedDefObj || null;
		state.activeDef = props.activeDef || null;
		state.activeDefNoRev = props.activeDefNoRev || null;
		state.highlightSelectedLines = Boolean(props.highlightSelectedLines);
		state.dispatchSelections = Boolean(props.dispatchSelections);
		state.displayRanges = props.displayRanges ? this._consolidateRanges(props.displayRanges.concat(state.expandedRanges)) : null;

		let updateAnns = false;

		if (state.annotations !== props.annotations && !props.skipAnns) {
			state.annotations = props.annotations;
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
	}

	_consolidateRanges(ranges: Range[]): ?Range[] {
		if (ranges.length === 0) return null;

		ranges = ranges.sort((a, b) => {
			if (a[0] < b[0]) return -1;
			if (a[0] > b[0]) return 1;
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

	_withinDisplayedRange(lineNumber: number): bool {
		if (!this.state.displayRanges) return false;
		for (let range of this.state.displayRanges) {
			if (range[0] <= lineNumber && lineNumber <= range[1]) return true;
		}
		return false;
	}

	_expandRange(range: Range): void {
		this.setState({
			expandedRanges: this.state.expandedRanges.concat([range]),
		});
	}

	_handleSelectionChange(ev: Event) {
		if (!this.state.dispatchSelections) return;

		const sel = document.getSelection();
		if (!sel || sel.rangeCount === 0) {
			return;
		}
		const rng = sel.getRangeAt(0);

		const getLineElem = (node: Node): ?HTMLElement => {
			if (!node) return null;
			if (node instanceof HTMLElement && node.nodeName === "TR" && node.dataset.line) {
				return node;
			}
			return getLineElem(node.parentNode);
		};

		const getColInLine = (lineElem, containerElem, offsetInContainer: number) => {
			let lineContentElem = lineElem.lastChild;
			let col = 0;
			let q = [lineContentElem];
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
		if (startLineElem && !$el.contains(startLineElem)) return;
		if (endLineElem && !$el.contains(endLineElem)) return;

		if (sel.isCollapsed) {
			// It's a click IN the file view with an empty (collapsed) selection.
			Dispatcher.Stores.dispatch(new BlobActions.SelectCharRange(this.state.repo, this.state.rev, this.state.path, null));
			return;
		}

		// Get start/end line + col. If the line elem isn't found, then the selection
		// extends beyond the bounds.
		let startLine = startLineElem ? parseInt(startLineElem.dataset.line, 10) : 1;
		let startCol = startLineElem ? getColInLine(startLineElem, rng.startContainer, rng.startOffset) : 0;
		let endLine = endLineElem ? parseInt(endLineElem.dataset.line, 10) : this.state.lines.length;
		let endCol = endLineElem ? getColInLine(endLineElem, rng.endContainer, rng.endOffset) : this.state.lines[this.state.lines.length - 1].length;

		let startByte = this.state.lineStartBytes[startLine - 1] + startCol;
		let endByte = this.state.lineStartBytes[endLine - 1] + endCol;

		Dispatcher.Stores.dispatch(new BlobActions.SelectCharRange(this.state.repo, this.state.rev, this.state.path, startLine, startCol, startByte, endLine, endCol, endByte));
	}

	onStateTransition(prevState: Blob.state, nextState: Blob.state) {
		if (nextState.startLine && prevState.startLine !== nextState.startLine) {
			if (!this._lineIsVisible(nextState.startLine)) {
				if (this.state.scrollToStartLine) {
					this._scrollTo(nextState.startLine);
				}
			}
		}
	}

	_updateVisibleLines() {
		let rect = this.refs.table.getBoundingClientRect();
		let lineCount = this.state.lines.length;
		if (this.state.displayRanges) {
			// Sum all elided lines. We don't incur rendering overhead for elided lines.
			lineCount += this.state.displayRanges.reduce((prev, cur) => prev + cur[1] - cur[0], 0);
		}
		let lineHeight = rect.height / lineCount;
		let firstLine = this.state.displayRanges && this.state.displayRanges[0] ? this.state.displayRanges[0][0] : this.state.contentsOffsetLine;
		let startLine = Math.max(0, (-1 * rect.top) / lineHeight) + firstLine;
		let numLines = window.innerHeight / lineHeight;
		this.setState({visStartLine: Math.ceil(startLine), visEndLine: Math.ceil(startLine + numLines)});
	}

	_scrollTo(line: number): void {
		if (!this.refs.table) { return; }
		let rect = this.refs.table.getBoundingClientRect();
		const y = rect.height / this.state.lines.length * (line - 1) - 100;
		window.scrollTo(0, y);
	}

	// _lineIsVisible returns true iff the line is scrolled into view.
	_lineIsVisible(line: number): boolean {
		if (!this._isMounted) return false;
		if (typeof document === "undefined" || typeof window === "undefined") return false;
		let top = document.body.scrollTop;
		let bottom = top + window.innerHeight;
		let $line = this.refs.table.querySelector(`[data-line="${line}"]`);
		if (!$line) return false;
		let elemTop = $line.offsetTop;
		let elemBottom = elemTop + $line.clientHeight;
		const deadZone = 150; // consider things not visible if they are near the screen top/bottom and hard to notice
		return elemBottom <= (bottom - deadZone) && elemTop >= (top + deadZone);
	}

	getOffsetTopForByte(byte: number): number {
		if (!this._isMounted) return 0;
		if (typeof byte === "undefined") throw new Error("getOffsetTopForByte byte is undefined");
		let $el = ReactDOM.findDOMNode(this);
		let line = lineFromByte(this.state.lines, byte);
		let $line = $el.querySelector(`[data-line="${line}"]`);
		let $toolbar = $el.parentNode.childNodes[0];
		if ($line) return $line.offsetTop + $toolbar.clientHeight + $toolbar.offsetTop;
		throw new Error(`No element found for line ${line}`);
	}

	render() {
		if (!this.state.lines) return null;

		let lastDisplayedLine = 0;
		let lastRangeEnd = 0;
		let lines = [];
		let renderedLines: number = 0;
		this.state.lines.forEach((line, i: number) => {
			const lineNumber = 1 + i + this.state.contentsOffsetLine;
			if (this.state.displayRanges && !this._withinDisplayedRange(lineNumber)) {
				return;
			}
			if (this.state.displayRanges && lastDisplayedLine !== lineNumber - 1) {
				// Prevent expanding above the last displayed range.
				let expandTo = [Math.max(lastRangeEnd, lineNumber-30), lineNumber-1];
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
					repo={this.state.repo}
					rev={this.state.rev}
					path={this.state.path}
					lineNumber={this.state.lineNumbers ? lineNumber : null}
					startByte={this.state.lineStartBytes[i]}
					contents={line}
					annotations={this.state.lineAnns ? (this.state.lineAnns[i] || null) : null}
					selected={this.state.highlightSelectedLines && this.state.startLine && this.state.endLine && this.state.startLine <= lineNumber && this.state.endLine >= lineNumber}
					highlightedDef={this.state.highlightedDef}
					highlightedDefObj={this.state.highlightedDefObj}
					activeDef={this.state.activeDef}
					activeDefNoRev={this.state.activeDefNoRev}
					key={i} />
			);
			renderedLines += 1;
		});
		if (this.state.lines && lastDisplayedLine < this.state.lines.length) {
			lines.push(
				<BlobLineExpander key={`expand-${this.state.lines.length}`}
					expandRange={[lastDisplayedLine, lastDisplayedLine+30]}
					onExpand={this._expandRange}
					direction={"down"} />
			);
		}

		return (
			<div className={s.scroller}>
				<table className={s.lines} ref="table">
					<tbody>
						{lines}
					</tbody>
				</table>
			</div>
		);
	}
}

export default Blob;
