var React = require("react");
var SearchResultsStore = require("../stores/SearchResultsStore");
var TokenSearchResultsView = require("./TokenSearchResultsView");

var SearchResultsView = React.createClass({
	getInitialState() {
		return SearchResultsStore.state;
	},

	componentDidMount() {
		// Listen to changes in stores and update accordingly.
		window.addEventListener(SearchResultsStore.onChange.type, this.onChange);
	},

	onChange() {
		this.setState(SearchResultsStore.state);
	},

	render() {
		var currentResultsView;
		if (this.state.tokenSearch) {
			currentResultsView = (
				<TokenSearchResultsView
					query={this.state.query}
					results={this.state.tokenSearch.Results} />
			);
		}

		return (
			<div className="search-results row">
				<div className="col-md-10 col-md-offset-1">
					{currentResultsView}
				</div>
			</div>
		);
	},
});

module.exports = SearchResultsView;
