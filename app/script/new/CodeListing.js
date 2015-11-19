import React from "react";

import Component from "./Component";
import CodeLineView from "./CodeLineView";

import classNames from "classnames";

const tilingFactor = 500;
const emptyArray = [];

export default class CodeListing extends Component {
	constructor(props) {
		super(props);
		this.state = {
			firstVisibleLine: 0,
			visibleLinesCount: tilingFactor * 3,
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
		Object.assign(state, props);
		state.lineNumbers = Boolean(props.lineNumbers);
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

	render() {
		let visibleLinesStart = this.state.firstVisibleLine;
		let visibleLinesEnd = visibleLinesStart + this.state.visibleLinesCount;

		let offscreenCodeAbove = "";
		this.state.lines.slice(0, visibleLinesStart).forEach((lineData) => {
			(lineData.Tokens || []).forEach((token) => {
				offscreenCodeAbove += token.Label || "";
			});
			offscreenCodeAbove += "\n";
		});

		let offscreenCodeBelow = "";
		this.state.lines.slice(visibleLinesEnd).forEach((lineData) => {
			(lineData.Tokens || []).forEach((token) => {
				offscreenCodeBelow += token.Label || "";
			});
			offscreenCodeBelow += "\n";
		});

		let lines = this.state.lines.slice(visibleLinesStart, visibleLinesEnd).map((lineData, i) => {
			let lineNumber = 1 + visibleLinesStart + i;
			let selected = this.state.startLine <= lineNumber && this.state.endLine >= lineNumber;
			return (
				<CodeLineView
					lineNumber={this.state.lineNumbers ? lineNumber : null}
					tokens={lineData.Tokens || emptyArray}
					selected={selected}
					selectedDef={this.state.selectedDef}
					highlightedDef={this.state.highlightedDef}
					lineButton={Boolean(this.state.lineNumbers && this.state.onLineButtonClick)}
					onLineButtonClick={this.state.onLineButtonClick}
					alwaysShowLineButton={this.state.lineNumbers && !this.state.lineSelectionForm && lineNumber === this.state.endLine}
					key={visibleLinesStart + i} />
			);
		});

		if (this.state.lineSelectionForm) {
			let form = (
				<tr key="form">
					<td className="line-number"></td>
					<td>
						{this.state.lineSelectionForm}
					</td>
				</tr>
			);

			lines.splice(this.state.endLine, 0, form);
		}

		let listingClasses = classNames({
			"line-numbered-code": true,
			"fade-unselected-lines": this.state.lineSelectionForm,
		});

		return (
			<table className={listingClasses} ref="table">
				<tbody>
					{offscreenCodeAbove !== "" &&
						<tr className="line">
							<td className="line-number"></td>
							<td className="line-content">{offscreenCodeAbove}</td>
						</tr>
					}
					{lines}
					{offscreenCodeBelow !== "" &&
						<tr className="line">
							<td className="line-number"></td>
							<td className="line-content">{offscreenCodeBelow}</td>
						</tr>
					}
				</tbody>
			</table>
		);
	}
}

CodeListing.propTypes = {
	lines: React.PropTypes.array,
	lineNumbers: React.PropTypes.bool,
	startLine: React.PropTypes.number,
	endLine: React.PropTypes.number,
	selectedDef: React.PropTypes.string,
	highlightedDef: React.PropTypes.string,
	lineSelectionForm: React.PropTypes.element,
	onLineButtonClick: React.PropTypes.func,
};
