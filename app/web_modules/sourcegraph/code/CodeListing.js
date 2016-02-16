import React from "react";
import ReactDOM from "react-dom";

import Component from "sourcegraph/Component";
import CodeLineView from "sourcegraph/code/CodeLineView";
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
	}

	componentDidMount() {
		this._updateVisibleLines();
		window.addEventListener("scroll", this._updateVisibleLines);
		if (this.state.startLine) {
			this._scrollTo(this.state.startLine);
		}
	}

	componentWillUnmount() {
		window.removeEventListener("scroll", this._updateVisibleLines);
	}

	reconcileState(state, props) {
		state.startLine = props.startLine || 0;
		state.endLine = props.endLine || 0;
		state.lineNumbers = Boolean(props.lineNumbers);
		state.selectedDef = props.selectedDef;
		state.highlightedDef = props.highlightedDef;

		let updateAnns = false;

		if (state.annotations !== props.annotations) {
			state.annotations = props.annotations;
			updateAnns = true;
		}

		if (state.contents !== props.contents) {
			state.contents = props.contents;
			state.lines = props.contents.split("\n");
			state.lineStartBytes = this._computeLineStartBytes(state.lines);
			updateAnns = true;
		}

		if (updateAnns) {
			state.lineAnns = [];
			(state.annotations || []).forEach((ann) => {
				let line = state.lineStartBytes.findIndex((startByte, i) => (
					(ann.StartByte || 0) >= startByte && ann.EndByte < startByte + state.lines[i].length + 1
				));
				if (line === -1) {
					throw new Error(`No line found for ann: ${JSON.stringify(ann)}`);
				}
				if (!state.lineAnns[line]) {
					state.lineAnns[line] = [];
				}
				state.lineAnns[line].push(ann);
			});
		}
	}

	_computeLineStartBytes(lines) {
		let pos = 0;
		return lines.map((line) => {
			let start = pos;
			pos += line.length + 1; // add 1 to account for newline
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

	onStateTransition(prevState, nextState) {
		if (nextState.startLine && nextState.selectedDef && prevState.startLine !== nextState.startLine) {
			this._scrollTo(nextState.startLine);
		}
	}

	_scrollTo(line) {
		if (!this.refs.table) { return; }
		let rect = this.refs.table.getBoundingClientRect();
		window.scrollTo(0, rect.height / this.state.lines.length * (line - 1) - 100);
	}

	getOffsetTopForByte(byte) {
		let $el = ReactDOM.findDOMNode(this);
		let line = lineFromByte(this.state.lines, byte);
		let $line = $el.querySelector(`[data-line="${line}"]`);
		if ($line) return $line.offsetTop - $el.parentNode.querySelector(".code-file-toolbar").offsetTop;
		throw new Error(`No element found for line ${line}`);
	}

	render() {
		let visibleLinesStart = this.state.firstVisibleLine;
		let visibleLinesEnd = visibleLinesStart + this.state.visibleLinesCount;

		let lines = this.state.lines.map((line, i) => {
			const visible = i >= visibleLinesStart && i < visibleLinesEnd;
			const lineNumber = 1 + i;
			return (
				<CodeLineView
					lineNumber={this.state.lineNumbers ? lineNumber : null}
					startByte={this.state.lineStartBytes[i]}
					contents={line}
					annotations={visible ? (this.state.lineAnns[i] || null) : null}
					selected={this.state.startLine <= lineNumber && this.state.endLine >= lineNumber}
					selectedDef={visible ? this.state.selectedDef : null}
					highlightedDef={visible ? this.state.highlightedDef : null}
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
	endLine: React.PropTypes.number,
	selectedDef: React.PropTypes.string,
	highlightedDef: React.PropTypes.string,
};

export default CodeListing;
