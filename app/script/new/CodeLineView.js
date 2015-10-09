var React = require("react");

var CodeLineView = React.createClass({
	propTypes: {
		lineNumber: React.PropTypes.number,
		tokens: React.PropTypes.array,
	},

	render() {
		return (
			<tr className="line">
				<td className="line-number" data-line={this.props.lineNumber}></td>
				<td className="line-content">
					{this.props.tokens.map((token, i) => <span className={token.Class} key={i}>{token.Label}</span>)}
				</td>
			</tr>
		);
	},
});

module.exports = CodeLineView;
