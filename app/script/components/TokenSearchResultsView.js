var React = require("react");
var TokenSearchResult = require("./TokenSearchResult");

var TokenSearchResultsView = React.createClass({
	render() {
		if (this.props.results) {
			if (this.props.results.length === 0) {
				return <i>No token results found for {this.props.query}</i>;
			}
			var results = this.props.results.map((result) => {
				return <TokenSearchResult key={result.URL} result={result} />;
			});
		}

		return (
			<div className="search-results">
				{results}
			</div>
		);
	},
});

module.exports = TokenSearchResultsView;
