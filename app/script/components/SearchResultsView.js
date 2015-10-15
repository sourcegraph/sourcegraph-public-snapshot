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

	_getCurrentSearchView() {
		var sharedProps = {
			query: this.state.query,
			repo: this.state.repo,
		};

		switch (this.state.currentSearchType) {
		case globals.SearchType.TOKEN:
			if (this.state.tokenSearch) {
				return (
					<TokenSearchResultsView
						{...sharedProps}
						loading={this.state.tokenSearchLoading}
						total={this.state.tokenSearch.Total}
						results={this.state.tokenSearch.Results} />
				);
			}
			break;
		case globals.SearchType.TEXT:
			if (this.state.textSearch) {
				return (
					<TextSearchResultsView
						{...sharedProps}
						loading={this.state.textSearchLoading}
						total={this.state.textSearch.Total}
						results={this.state.textSearch.Results} />
				);
			}
			break;
		default:
			return null;
		}
	},

	render() {
		var currentResultsView = this._getCurrentSearchView();

		var loadingIcon = <i className="fa fa-circle-o-notch fa-spin"></i>;
		var tokenSearchStatusBadge = null;
		if (this.state.tokenSearchLoading) {
			tokenSearchStatusBadge = loadingIcon;
		} else if (this.state.tokenSearch) {
			tokenSearchStatusBadge = this.state.tokenSearch.Total;
		}
		var textSearchStatusBadge = null;
		if (this.state.textSearchLoading) {
			textSearchStatusBadge = loadingIcon;
		} else if (this.state.textSearch) {
			textSearchStatusBadge = this.state.textSearch.Total;
		}

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
