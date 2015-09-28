var React = require("react");
var $ = require("jquery");
var router = require("../routing/router");

var CloseChangesetButton = React.createClass({
	propTypes: {
		label: React.PropTypes.string,
		repo: React.PropTypes.string.isRequired,
		changesetId: React.PropTypes.string.isRequired,
		afterClose: React.PropTypes.func,
		redirectUrl: React.PropTypes.string,
	},

	getDefaultProps() {
		return {
			label: "Close changeset",
			afterClose() {},
		};
	},

	_onClick(e) {
		e.preventDefault();

		return $.ajax({
			url: "/ui" + router.changesetURL(this.props.repo, this.props.changesetId) + "/update",
			method: "POST",
			data: JSON.stringify({Close: true}),
		}).then(function(data) {
			if (!data.hasOwnProperty("Error")) {
				if (this.props.redirectUrl) {
					window.location = this.props.redirectUrl;
				} else {
					this.props.afterClose();
				}
			}
		}.bind(this));
	},

	render() {
		return <a className="btn btn-default close-changeset-button" onClick={this._onClick}>{this.props.label}</a>;
	},
});

module.exports = CloseChangesetButton;
