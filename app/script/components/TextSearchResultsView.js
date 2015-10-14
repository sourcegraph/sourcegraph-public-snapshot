var React = require("react");

var globals = require("../globals");
var routing = require("../routing/router");
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
			return <i>No text results found for "{this.props.query}"</i>;
		}

		var currentFile, header;
		var results = this.props.results.map((result) => {
			if (currentFile !== result.File) {
				// TODO(renfred) link to user-specified rev instead of default branch.
				var fileURL = routing.fileURL(this.props.repo.URI, this.props.repo.DefaultBranch, result.File);
				header = <header><a href={fileURL}>{result.File}</a></header>;
			} else {
				header = null;
			}
			currentFile = result.File;

			return (
				<div className="text-search-result" key={`${result.File}-${result.StartLine}`}>
					{header}
					<TextSearchResult result={result} repo={this.props.repo} />
				</div>
			);
		});

		var s = this.props.results.length === 1 ? "" : "s";
		var summary = `${this.props.total} result${s} for "${this.props.query}"`;
		if (this.state.currentPage > 1) summary = `Page ${this.state.currentPage} of ${summary}`;

		return (
			<div className="text-search-results">
				<i className="summary">{summary}</i>
				{results}
				<div className="search-pagination">
					<Pagination
						currentPage={this.state.currentPage}
						totalPages={Math.ceil(this.props.total/globals.TextSearchResultsPerPage)}
						pageRange={10}
						loading={this.props.loading}
						onPageChange={this.onPageChange} />
				</div>
			</div>
		);
	},
});

module.exports = TextSearchResultsView;
