var React = require("react");

var globals = require("../globals");
var Pagination = require("./Pagination");
var TokenSearchResult = require("./TokenSearchResult");
var SearchActions = require("../actions/SearchActions");

var TokenSearchResultsView = React.createClass({
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
		SearchActions.searchRepoTokens(this.props.query, this.props.repo, page);
	},

	render() {
		if (!this.props.results) return null;

		var results = this.props.results.map((result) => {
			return <TokenSearchResult key={result.URL} result={result} />;
		});

		var summary;
		if (this.props.results.length === 0) {
			summary = `No definition results found for "${this.props.query}"`;
		} else {
			var s = this.props.results.length === 1 ? "" : "s";
			summary = `${this.props.total} definition result${s} for "${this.props.query}"`;
			if (this.state.currentPage > 1) summary = `Page ${this.state.currentPage} of ${summary}`;
		}

		var buildInfo;
		if (this.props.buildInfo) {
			buildInfo = (
				<div className="alert alert-info" role="alert">
					<i className="fa fa-warning"></i> {this.props.buildInfo}
				</div>
			);
		}

		var pagination;
		if (this.props.results.length > 0) {
			pagination = (
				<Pagination
					currentPage={this.state.currentPage}
					totalPages={Math.ceil(this.props.total/globals.TokenSearchResultsPerPage)}
					pageRange={10}
					loading={this.props.loading}
					onPageChange={this.onPageChange} />
			);
		}

		return (
			<div className="token-search-results">
				{buildInfo}
				<p className="summary">{summary}</p>
				{results}
				<div className="search-pagination">
					{pagination}
				</div>
			</div>
		);
	},
});

module.exports = TokenSearchResultsView;
