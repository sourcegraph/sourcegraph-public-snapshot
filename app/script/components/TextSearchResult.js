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
		var result = this.props.result;
		var snippetURL = routing.fileSnippetURL(
			this.props.repo.URI, this.props.repo.DefaultBranch, result.File, result.StartLine, result.EndLine
		);

		var lines = result.Lines.map((line, i) => {
			var lineNumber = result.StartLine + i;

			var snippetLink = null;
			if (i === 0) {
				snippetLink = <td className="snippet-link"><a href={snippetURL}>{result.StartLine}-{result.EndLine}</a></td>;
			}

			return (
				<tr className="line" key={i}>
					<td className="line-number" onClick={this._goToLine.bind(this, lineNumber)}>{lineNumber}</td>
					<td className="line-content" dangerouslySetInnerHTML={{__html: line}}></td>
					{snippetLink}
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
