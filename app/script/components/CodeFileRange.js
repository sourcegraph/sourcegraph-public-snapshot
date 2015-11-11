var React = require("react");

var routing = require("../routing/router");

var CodeFileRange = React.createClass({
	_goToLine(lineNumber) {
		var lineURL = routing.fileRangeURL(
			this.props.repo, this.props.rev, this.props.path, lineNumber, lineNumber
		);
		window.location.href = lineURL;
	},

	render() {
		var lines = this.props.lines.map((line, i) => {
			// Sometimes there will be a blank line before EOF that is not counted as part of the line
			// range. Do not display these lines.
			if (i + this.props.startLine > this.props.endLine) return null;

			var lineNumber = this.props.startLine + i;

			var fileRangeLink = null;
			if (this.props.showFileRangeLink && i === 0) {
				var lineURL = routing.fileRangeURL(
					this.props.repo, this.props.rev, this.props.path, this.props.startLine, this.props.endLine
				);

				fileRangeLink = <td className="file-range-link"><a href={lineURL}>{this.props.startLine}-{this.props.endLine}</a></td>;
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
			<table className="line-numbered-code static-code-view theme-default">
				<tbody>
					{lines}
				</tbody>
			</table>
		);
	},
});

module.exports = CodeFileRange;

CodeFileRange.propTypes = {
	repo: React.PropTypes.string,
	rev: React.PropTypes.string,
	path: React.PropTypes.string,
	startLine: React.PropTypes.number,
	endLine: React.PropTypes.number,
	lines: React.PropTypes.arrayOf(React.PropTypes.object),
	showFileRangeLink: React.PropTypes.bool,
};
