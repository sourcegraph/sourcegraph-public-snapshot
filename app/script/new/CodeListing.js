import React from "react";

import CodeLineView from "./CodeLineView";

const tilingFactor = 500;
const emptyArray = [];

export default class CodeListing extends React.Component {
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
	}

	shouldComponentUpdate(nextProps, nextState) {
		return nextProps.lines !== this.props.lines ||
			nextProps.selectedDef !== this.props.selectedDef ||
			nextProps.highlightedDef !== this.props.highlightedDef ||
			nextState.firstVisibleLine !== this.state.firstVisibleLine ||
			nextState.visibleLinesCount !== this.state.visibleLinesCount;
	}

	componentWillUnmount() {
		window.removeEventListener("scroll", this._updateVisibleLines);
	}

	_updateVisibleLines() {
		let rect = this.refs.table.getBoundingClientRect();
		let firstVisibleLine = Math.max(0, Math.floor(this.props.lines.length / rect.height * -rect.top / tilingFactor - 1) * tilingFactor);
		let visibleLinesCount = Math.ceil(this.props.lines.length / rect.height * window.innerHeight / tilingFactor + 2) * tilingFactor;
		if (this.state.firstVisibleLine !== firstVisibleLine || this.state.visibleLinesCount !== visibleLinesCount) {
			this.setState({
				firstVisibleLine: firstVisibleLine,
				visibleLinesCount: visibleLinesCount,
			});
		}
	}

	render() {
		let visibleLinesStart = this.state.firstVisibleLine;
		let visibleLinesEnd = visibleLinesStart + this.state.visibleLinesCount;

		let offscreenCodeAbove = "";
		this.props.lines.slice(0, visibleLinesStart).forEach((lineData) => {
			(lineData.Tokens || []).forEach((token) => {
				offscreenCodeAbove += token.Label || "";
			});
			offscreenCodeAbove += "\n";
		});

		let offscreenCodeBelow = "";
		this.props.lines.slice(visibleLinesEnd).forEach((lineData) => {
			(lineData.Tokens || []).forEach((token) => {
				offscreenCodeBelow += token.Label || "";
			});
			offscreenCodeBelow += "\n";
		});

		return (
			<table className="line-numbered-code" ref="table">
				<tbody>
					{offscreenCodeAbove !== "" &&
						<tr className="line">
							<td className="line-number"></td>
							<td className="line-content">{offscreenCodeAbove}</td>
						</tr>
					}
					{this.props.lines.slice(visibleLinesStart, visibleLinesEnd).map((lineData, i) =>
						<CodeLineView
							lineNumber={this.props.lineNumbers ? 1 + visibleLinesStart + i : null}
							tokens={lineData.Tokens || emptyArray}
							selectedDef={this.props.selectedDef}
							highlightedDef={this.props.highlightedDef}
							key={visibleLinesStart + i} />
					)}
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
	selectedDef: React.PropTypes.string,
	highlightedDef: React.PropTypes.string,
};
