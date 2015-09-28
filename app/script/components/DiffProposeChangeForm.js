var React = require("react");
var $ = require("jquery");
var DiffActions = require("../actions/DiffActions");
var MarkdownTextarea = require("./MarkdownTextarea");

var ProposeChangeForm = React.createClass({
	getDefaultProps() {
		return {
			loading: false,
		};
	},

	_createChangeset() {
		if (!this.isMounted()) {
			return;
		}

		var root = this.getDOMNode();

		DiffActions.proposeChange(this.props.deltaSpec.Base.URI, {
			DeltaSpec: this.props.deltaSpec,
			Title: $(root).find("input.title").val(),
			Description: this.refs.description.value(),
		});
	},

	render() {
		return (
			<div className="changeset-propose-form">
				<input type="text" className="title" placeholder="Title" />
				<MarkdownTextarea ref="description" placeholder="Enter a description..." />
				<div className="actions">
					{this.props.changesetLoading ? <span>Loading...</span> : null}
					<a className="btn btn-success pull-right" onClick={this._createChangeset}>Submit</a>
					<a className="btn pull-right" onClick={this.props.onCancel}>Cancel</a>
				</div>
			</div>
		);
	},
});

module.exports = ProposeChangeForm;
