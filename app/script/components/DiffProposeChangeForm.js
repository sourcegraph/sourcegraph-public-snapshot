var React = require("react");
var ReactDOM = require("react-dom");
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

		var root = ReactDOM.findDOMNode(this);

		DiffActions.proposeChange(this.props.deltaSpec.Base.URI, {
			DeltaSpec: this.props.deltaSpec,
			Title: $(root).find("input.title").val(),
			Description: this.refs.description.value(),
		});
	},

	render() {
		return (
			<div className="changeset-propose-form">
				<input type="text" className="title" placeholder="Title" autoFocus="true" />
				<MarkdownTextarea ref="description" placeholder="Enter a description..." />
				<div className="actions">
					{this.props.changesetLoading ? <span>Loading...</span> : null}
					<div className="pull-right">
						<button className="btn btn-success" onClick={this._createChangeset} tabIndex="0">Submit</button>
						<button className="btn btn-neutral" onClick={this.props.onCancel} tabIndex="0">Cancel</button>
					</div>
				</div>
			</div>
		);
	},
});

module.exports = ProposeChangeForm;
