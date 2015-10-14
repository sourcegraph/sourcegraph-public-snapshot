import React from "react";

class CodeLineView extends React.Component {
	render() {
		return (
			<tr className="line">
				<td className="line-number" data-line={this.props.lineNumber}></td>
				<td className="line-content">
					{this.props.tokens.map((token, i) => <span className={token.Class} key={i}>{token.Label}</span>)}
				</td>
			</tr>
		);
	}
}

CodeLineView.propTypes = {
	lineNumber: React.PropTypes.number,
	tokens: React.PropTypes.array,
};

export default CodeLineView;
