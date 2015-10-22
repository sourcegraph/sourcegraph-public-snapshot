import React from "react";

import CodeLineView from "./CodeLineView";

const emptyArray = [];

class CodeListing extends React.Component {
	shouldComponentUpdate(nextProps, nextState) {
		return nextProps.lines !== this.props.lines ||
			nextProps.selectedDef !== this.props.selectedDef ||
			nextProps.highlightedDef !== this.props.highlightedDef;
	}

	render() {
		// TODO tiled rendering for better performance on huge files
		return (
			<table className="line-numbered-code">
				<tbody>
					{this.props.lines.map((lineData, i) =>
						<CodeLineView
							lineNumber={this.props.lineNumbers ? i + 1 : null}
							tokens={lineData.Tokens || emptyArray}
							selectedDef={this.props.selectedDef}
							highlightedDef={this.props.highlightedDef}
							key={i} />
					)}
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

export default CodeListing;
