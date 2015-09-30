var React = require("react");

var SearchResultsView = React.createClass({
	render() {
		return (
			<div className="search-results">
				<i>No search results found</i>
			</div>
		);
	},
});

module.exports = SearchResultsView;
