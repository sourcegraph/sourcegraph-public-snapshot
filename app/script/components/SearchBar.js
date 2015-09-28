var React = require("react");

var SearchBar = React.createClass({
	render() {
		return (
			<form className="navbar-form" role="search">
				<div className="form-group">
					<div className="input-group">
						<input className="form-control search-input-next"
							name="search"
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
