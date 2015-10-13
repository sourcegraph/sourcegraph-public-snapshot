var React = require("react");

var routing = require("../routing/router");

var TextSearchResult = React.createClass({
	_goToLine(lineNumber) {
		// TODO(renfred) link to user-specified rev instead of default branch.
		var lineURL = routing.gURL(
			this.props.repo.URI, this.props.repo.DefaultBranch, this.props.result.File, lineNumber, lineNumber
		);
		window.location.href = lineURL;
	},

	render() {
		var result = this.props.result;
		var gURL = routing.fileRangeURL(
			this.props.repo.URI, this.props.repo.DefaultBranch, result.File, result.StartLine, result.EndLine
		);

		var lines = result.Lines.map((line, i) => {
			// Sometimes there will be a blank line before EOF that is not counted as part of the line
			// range. Do not display these lines.
			if (i + result.StartLine > result.EndLine) return null;

			var lineNumber = result.StartLine + i;

			var fileRangeLink = null;
			if (i === 0) {
				fileRangeLink = <td className="file-range-link"><a href={gURL}>{result.StartLine}-{result.EndLine}</a></td>;
			}

			return (
				<tr className="line" key={i}>
					<td className="line-number" onClick={this._goToLine.bind(this, lineNumber)}>{lineNumber}</td>
					{/* This HTML should be sanitized in ui/search.go */}
					<td className="line-content" dangerouslySetInnerHTML={line}></td>
					{fileRangeLink}
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
