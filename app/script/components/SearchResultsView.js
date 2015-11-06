var React = require("react");

var globals = require("../globals");
var SearchActions = require("../actions/SearchActions");
var SearchResultsStore = require("../stores/SearchResultsStore");
var TokenSearchResultsView = require("./TokenSearchResultsView");
var TextSearchResultsView = require("./TextSearchResultsView");

var SearchResultsView = React.createClass({
	getInitialState() {
		return SearchResultsStore.state;
	},

	componentDidMount() {
		// Listen to changes in stores and update accordingly.
		window.addEventListener(SearchResultsStore.onChange.type, this.onChange);
	},

	componentDidUpdate(prevProps, prevState) {
		if (prevState.query !== this.state.query) {
			// When initiating a new search query, scroll to top of page to
			// view new results.
			window.scrollTo(0, 0);
		}
	},

	onChange() {
		this.setState(SearchResultsStore.state);
	},

	_selectSearchType(searchType) {
		SearchActions.selectSearchType(searchType);
	},

	_statusBadge(search) {
		var loadingIcon = <i className="fa fa-circle-o-notch fa-spin"></i>;
		if (search.loading) {
			return loadingIcon;
		} else if (search.data) {
			return search.data.Total;
		}
		return 0;
	},

	_errorMessage(error, searchType) {
		var message;
		switch (error.statusText) {
		case "timeout":
			message = "This search took took too long to process.";
			break;
		default:
			message = `There was an error returning ${searchType} search results.`;
		}
		return (
			<div className="alert alert-warning">
				<i className="fa fa-warning"></i>	{message}
			</div>
		);
	},

	_getCurrentSearchView() {
		var sharedProps = {
			query: this.state.query,
			repo: this.state.repo,
		};

		switch (this.state.currentSearchType) {
		case globals.SearchType.TOKEN:
			var tokenSearch = this.state.tokenSearch;
			if (tokenSearch.error) {
				return this._errorMessage(tokenSearch.error, "definition");
			}
			if (tokenSearch.data) {
				return (
					<TokenSearchResultsView
						{...sharedProps}
						loading={tokenSearch.loading}
						total={tokenSearch.data.Total}
						results={tokenSearch.data.Results}
						buildInfo={tokenSearch.data.BuildInfo} />
				);
			}
			break;
		case globals.SearchType.TEXT:
			var textSearch = this.state.textSearch;
			if (textSearch.error) {
				return this._errorMessage(textSearch.error, "text");
			}
			if (textSearch.data) {
				return (
					<TextSearchResultsView
						{...sharedProps}
						loading={textSearch.loading}
						total={textSearch.data.Total}
						results={textSearch.data.Results} />
				);
			}
			break;
		default:
			return null;
		}
	},

	render() {
		var currentResultsView = this._getCurrentSearchView();

		var tokenSearchStatusBadge = this._statusBadge(this.state.tokenSearch);
		var textSearchStatusBadge = this._statusBadge(this.state.textSearch);

		return (
			<div className="search-results row">
				<div className="col-md-10 col-md-offset-1">
					<ul className="nav nav-pills">
						<li className={this.state.currentSearchType === globals.SearchType.TOKEN ? "active" : null}>
							<a onClick={this._selectSearchType.bind(this, globals.SearchType.TOKEN)}>
								<i className="fa fa-asterisk"></i> Definitions <span className="badge">{tokenSearchStatusBadge}</span>
							</a>
						</li>
						<li className={this.state.currentSearchType === globals.SearchType.TEXT ? "active" : null}>
							<a onClick={this._selectSearchType.bind(this, globals.SearchType.TEXT)}>
								<i className="fa fa-code"></i> Text <span className="badge">{textSearchStatusBadge}</span>
							</a>
						</li>
					</ul>
					{currentResultsView}
				</div>
			</div>
		);
	},
});

module.exports = SearchResultsView;
