var React = require("react");

var CodeFileRange = require("./CodeFileRange");

var TextSearchResult = React.createClass({
	render() {
		return (
			<CodeFileRange
				repo={this.props.repo.URI}
				rev={this.props.repo.DefaultBranch}
				path={this.props.result.File}
				startLine={this.props.result.StartLine}
				endLine={this.props.result.EndLine}
				lines={this.props.result.Lines}
				showFileRangeLink={true} />
		);
	},
});

module.exports = TextSearchResult;
