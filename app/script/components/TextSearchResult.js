var React = require("react");

var routing = require("../routing/router");

var TextSearchResult = React.createClass({
	_goToLine(lineNumber) {
		// TODO(renfred) link to user-specified rev instead of default branch.
		var lineURL = routing.fileSnippetURL(
			this.props.repo.URI, this.props.repo.DefaultBranch, this.props.result.File, lineNumber, lineNumber
		);
		window.location.href = lineURL;
	},

	render() {
		var lines = this.props.result.Lines.map((line, i) => {
			var lineNumber = this.props.result.StartLine + i;
			return (
				<tr className="line" key={i}>
					<td className="line-number" onClick={this._goToLine.bind(this, lineNumber)}>{lineNumber}</td>
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
