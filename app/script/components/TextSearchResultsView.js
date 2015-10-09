var React = require("react");

var globals = require("../globals");
var Pagination = require("./Pagination");
var TextSearchResult = require("../components/TextSearchResult");
var SearchActions = require("../actions/SearchActions");

var TextSearchResultsView = React.createClass({
	getInitialState() {
		return {currentPage: 1};
	},

	componentDidUpdate(prevProps) {
		if (prevProps.loading === true && this.props.loading === false) {
			// When finished loading a new page, scroll to top of page to
			// view new results.
			window.scrollTo(0, 0);
		}
	},

	onPageChange(page) {
		this.setState({currentPage: page});
		SearchActions.searchRepoText(this.props.query, this.props.repo, page);
	},

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

		var summary = `${this.props.total} results for "${this.props.query}"`;
		if (this.state.currentPage > 1) summary = `Page ${this.state.currentPage} of ${summary}`;

		return (
			<div className="text-search-results">
				<i className="summary">{summary}</i>
				{results}
				<div className="search-pagination">
					<Pagination
						currentPage={this.state.currentPage}
						totalPages={Math.ceil(this.props.total/globals.TokenSearchResultsPerPage)}
						pageRange={11}
						onPageChange={this.onPageChange} />
				</div>
			</div>
		);
	},
});

module.exports = TextSearchResultsView;
