var React = require("react");

var TextSearchResult = require("../components/TextSearchResult");

var TextSearchResultsView = React.createClass({
	render() {
		if (!this.props.results) return null;

		if (this.props.results.length === 0) {
			return <i>No token results found for {this.props.query}</i>;
		}

		var currentFile, header;
		var results = this.props.results.map((result) => {
			if (currentFile !== result.File) {
				header = <header>{result.File}</header>;
			} else {
				header = null;
			}
			currentFile = result.File;

			return (
				<div className="text-search-result" key={`${result.File}-${result.StartLine}`}>
					{header}
					<TextSearchResult result={result} />
				</div>
			);
		});

		return (
			<div>
				{results}
			</div>
		);
	},
});

module.exports = TextSearchResultsView;
