var React = require("react");
var SearchResultsStore = require("../stores/SearchResultsStore");

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
		return (
			<div className="search-results row">
				<div className="col-md-10 col-md-offset-1">
					<i>No search results found for {this.state.query}</i>
				</div>
			</div>
		);
	},
});

module.exports = SearchResultsView;
