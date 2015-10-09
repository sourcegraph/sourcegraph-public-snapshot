var React = require("react");

var TextSearchResult = React.createClass({
	render() {
		var lines = this.props.result.Lines.map((line, i) => {
			return (
				<tr className="line" key={i}>
					<td className="line-number">{this.props.result.StartLine + i}</td>
					<td className="line-content" dangerouslySetInnerHTML={{__html: line}}></td>
				</tr>
			);
		});

		return (
			<table className="line-numbered-code theme-default">
				<tbody>
					{lines}
				</tbody>
			</table>
		);
	},
});

module.exports = TextSearchResult;
