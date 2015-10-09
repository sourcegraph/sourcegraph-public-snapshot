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

		var tokenSearchStatusBadge = this.state.tokenSearch ?
			<span className="badge">{this.state.tokenSearch.Total}</span> : null;
		var textSearchStatusBadge = this.state.textSearch ?
			<span className="badge">{this.state.textSearch.Total}</span> : null;

		return (
			<div className="search-results row">
				<div className="col-md-10 col-md-offset-1">
					<ul className="nav nav-pills">
						<li role="presentation"
							className={this.state.currentSearchType === globals.SearchType.TOKEN ? "active" : null}>
							<a href="#" onClick={this._selectSearchType.bind(this, globals.SearchType.TOKEN)}>
								<i className="fa fa-asterisk"></i> Token {tokenSearchStatusBadge}
							</a>
						</li>
						<li role="presentation"
							className={this.state.currentSearchType === globals.SearchType.TEXT ? "active" : null}>
							<a href="#" onClick={this._selectSearchType.bind(this, globals.SearchType.TEXT)}>
								<i className="fa fa-code"></i> Text {textSearchStatusBadge}
							</a>
						</li>
					</ul>
					<hr/>
					{currentResultsView}
				</div>
			</div>
		);
	},
});

module.exports = SearchResultsView;
