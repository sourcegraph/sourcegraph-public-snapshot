var React = require("react");

var CodeLineView = require("./CodeLineView");

var CodeListing = React.createClass({
	propTypes: {
		lines: React.PropTypes.array,
	},

	render() {
		// TODO tiled rendering for better performance on huge files
		return (
			<div className="code-view-react">
				<table className="line-numbered-code">
					<tbody>
						{this.props.lines.map((lineData, i) => <CodeLineView lineNumber={i + 1} tokens={lineData.Tokens || []} key={i} />)}
					</tbody>
				</table>
			</div>
		);
	},
});

module.exports = CodeListing;
