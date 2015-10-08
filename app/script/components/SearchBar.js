var React = require("react");
var ReactDOM = require("react-dom");
var SearchRouter = require("../routing/SearchRouter");

var SearchBar = React.createClass({
	componentDidMount() {
		if (this.props.searchOptions) {
			var query = this.props.searchOptions.Query;
			ReactDOM.findDOMNode(this.refs.queryInput).value = query;
			this._submitSearch();
		}
	},

	_submitSearch(e) {
		if (e) e.preventDefault();
		var query = ReactDOM.findDOMNode(this.refs.queryInput).value;
		if (this.props.repo) {
			SearchRouter.searchRepo(query, this.props.repo);
		}
	},

	render() {
		return (
			<form className="navbar-form" onSubmit={this._submitSearch}>
				<div className="form-group">
					<div className="input-group">
						<input className="form-control search-input-next"
							ref="queryInput"
							name="q"
							placeholder="Search"
							type="text" />
						<span className="input-group-addon"></span>
					</div>
				</div>
			</form>
		);
	},
});

module.exports = SearchBar;
