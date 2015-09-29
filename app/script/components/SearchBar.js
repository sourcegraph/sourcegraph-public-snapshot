var React = require("react");

var SearchBar = React.createClass({
	componentDidMount() {
		if (this.props.searchOptions) {
			var query = this.props.searchOptions.Query;
			React.findDOMNode(this.refs.queryInput).value = query;
		}
	},

	render() {
		return (
			<form className="navbar-form" role="search">
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
