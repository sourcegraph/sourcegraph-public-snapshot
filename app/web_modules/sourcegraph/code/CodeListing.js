import React from "react";
import ReactDOM from "react-dom";
import utf8 from "utf8";

import Component from "sourcegraph/Component";
import * as CodeActions from "sourcegraph/code/CodeActions";
import CodeLineView from "sourcegraph/code/CodeLineView";
import Dispatcher from "sourcegraph/Dispatcher";
import debounce from "lodash/function/debounce";
import fileLines from "sourcegraph/util/fileLines";
import lineFromByte from "sourcegraph/code/lineFromByte";

const tilingFactor = 50;

class CodeListing extends Component {
	constructor(props) {
		super(props);
		this.state = {
			firstVisibleLine: 0,
			visibleLinesCount: tilingFactor * 3,
			lineAnns: [],
		};
		this._updateVisibleLines = this._updateVisibleLines.bind(this);
		this._handleSelectionChange = debounce(this._handleSelectionChange.bind(this), 200, {
			leading: false,
			trailing: true,
		});
		this._isMounted = false;
	}

	componentDidMount() {
		this._updateVisibleLines();
		window.addEventListener("scroll", this._updateVisibleLines);
		if (this.state.startLine) {
			this._scrollTo(this.state.startLine);
		}
		document.addEventListener("selectionchange", this._handleSelectionChange);
		this._isMounted = true;
	}

	componentWillUnmount() {
		window.removeEventListener("scroll", this._updateVisibleLines);
		document.removeEventListener("selectionchange", this._handleSelectionChange);
		this._isMounted = false;
	}

	reconcileState(state, props) {
		state.startLine = props.startLine || null;
		state.endLine = props.endLine || null;
		state.contentsOffsetLine = props.contentsOffsetLine || 0;
		state.lineNumbers = Boolean(props.lineNumbers);
		state.highlightedDef = props.highlightedDef;
		state.activeDef = props.activeDef || null;
		state.highlightSelectedLines = Boolean(props.highlightSelectedLines);
		state.dispatchSelections = Boolean(props.dispatchSelections);

		let updateAnns = false;

		if (state.annotations !== props.annotations) {
			state.annotations = props.annotations;
			updateAnns = true;
		}

		if (state.contents !== props.contents) {
			state.contents = props.contents;
			state.lines = fileLines(props.contents);
			state.lineStartBytes = this._computeLineStartBytes(state.lines);
			updateAnns = true;
		}

		if (updateAnns) {
			state.lineAnns = new Array(state.lineStartBytes ? state.lineStartBytes.length : 0);
			if (state.lineStartBytes && state.annotations) {
				state.lineStartBytes.forEach((lineStartByte, i) => {
					let lineEndByte = state.lineStartBytes[i + 1] || lineStartByte + state.lines[i].length;
					state.lineAnns[i] = (state.annotations || []).filter((ann) => (
						ann.StartByte < lineEndByte && ann.EndByte >= lineStartByte
					));
				});
			}
		}
	}

	_computeLineStartBytes(lines) {
		let pos = 0;
		return lines.map((line) => {
			let start = pos;
			// Encode the line using utf8 to account for multi-byte unicode characters.
			pos += utf8.encode(line).length + 1; // add 1 to account for newline
			return start;
		});
	}

	_updateVisibleLines() {
		let rect = this.refs.table.getBoundingClientRect();
		let firstVisibleLine = Math.max(0, Math.floor(this.state.lines.length / rect.height * -rect.top / tilingFactor - 1) * tilingFactor);
		let visibleLinesCount = Math.ceil(this.state.lines.length / rect.height * window.innerHeight / tilingFactor + 2) * tilingFactor;
		if (this.state.firstVisibleLine !== firstVisibleLine || this.state.visibleLinesCount !== visibleLinesCount) {
			this.setState({
				firstVisibleLine: firstVisibleLine,
				visibleLinesCount: visibleLinesCount,
			});
		}
	}

	_handleSelectionChange(ev) {
		if (!this.state.dispatchSelections) return;

		const sel = document.getSelection();
		if (!sel || sel.rangeCount === 0) {
			return;
		}
		const rng = sel.getRangeAt(0);

		const getLineElem = (node) => {
			if (!node) return null;
			if (node.nodeName === "TR" && node.classList.contains("line")) {
				return node;
			}
			return getLineElem(node.parentNode);
		};

		const getColInLine = (lineElem, containerElem, offsetInContainer) => {
			let lineContentElem = lineElem.querySelector(".line-content");
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
			Dispatcher.dispatch(new CodeActions.SelectCharRange(null));
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

		// console.log("%d:%d[%d] - %d:%d[%d]", startLine, startCol, startByte, endLine, endCol, endByte);
		Dispatcher.dispatch(new CodeActions.SelectCharRange(startLine, startCol, startByte, endLine, endCol, endByte));
	}

	onStateTransition(prevState, nextState) {
		if (nextState.startLine && nextState.highlightedDef && prevState.startLine !== nextState.startLine) {
			this._scrollTo(nextState.startLine);
		}
	}

	_scrollTo(line) {
		if (!this.refs.table) { return; }
		let rect = this.refs.table.getBoundingClientRect();
		window.scrollTo(0, rect.height / this.state.lines.length * (line - 1) - 100);
	}

	getOffsetTopForByte(byte) {
		if (!this._isMounted) return 0;
		let $el = ReactDOM.findDOMNode(this);
		let line = lineFromByte(this.state.lines, byte);
		let $line = $el.querySelector(`[data-line="${line}"]`);
		if ($line) return $line.offsetTop + $el.parentNode.querySelector(".code-file-toolbar").clientHeight;
		throw new Error(`No element found for line ${line}`);
	}

	render() {
		let visibleLinesStart = this.state.firstVisibleLine;
		let visibleLinesEnd = visibleLinesStart + this.state.visibleLinesCount;

		let lines = this.state.lines.map((line, i) => {
			const visible = i >= visibleLinesStart && i < visibleLinesEnd;
			const lineNumber = 1 + i + this.state.contentsOffsetLine;
			return (
				<CodeLineView
					lineNumber={this.state.lineNumbers ? lineNumber : null}
					startByte={this.state.lineStartBytes[i]}
					contents={line}
					annotations={visible ? (this.state.lineAnns[i] || null) : null}
					selected={this.state.highlightSelectedLines && this.state.startLine <= lineNumber && this.state.endLine >= lineNumber}
					highlightedDef={visible ? this.state.highlightedDef : null}
					activeDef={visible ? this.state.activeDef : null}
					key={i} />
			);
		});

		return (
			<table className="line-numbered-code" ref="table">
				<tbody>
					{lines}
				</tbody>
			</table>
		);
	}
}

CodeListing.propTypes = {
	contents: React.PropTypes.string,
	annotations: React.PropTypes.array,
	lineNumbers: React.PropTypes.bool,
	startLine: React.PropTypes.number,
	startCol: React.PropTypes.number,
	endLine: React.PropTypes.number,
	endCol: React.PropTypes.number,
	highlightedDef: React.PropTypes.string,
	activeDef: React.PropTypes.string,

	// contentsOffsetLine indicates that the contents string does not
	// start at line 1 within the file, but rather some other line number.
	// It must be specified when startLine > 1 but the contents don't begin at
	// the first line of the file.
	contentsOffsetLine: React.PropTypes.number,

	highlightSelectedLines: React.PropTypes.bool,

	// dispatchSelections is whether this CodeListing should emit CodeActions.SelectCharRange
	// actions when the text selection changes. It should be true for the main file view but
	// not for secondary file views (e.g., usage examples).
	dispatchSelections: React.PropTypes.bool,
};

export default CodeListing;
