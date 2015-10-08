var React = require("react");
var MarkdownTextarea = require("./MarkdownTextarea");

/*
 * @description Displays the discussion create form. See propTypes for configuration
 * options.
 */
var DiscussionCreateForm = React.createClass({

	propTypes: {
		// defName is the qualified name of the definition that will
		// be displayed in the UI.
		defName: React.PropTypes.object.isRequired,

		// onCancel is called when the user clicks the Cancel button
		// on the form.
		onCancel: React.PropTypes.func.isRequired,

		// onCreate is called when the user clicks the Create button.
		onCreate: React.PropTypes.func.isRequired,
	},

	_create() {
		var title = this.refs.titleText.value;
		var body = this.refs.bodyText.value();
		this.props.onCreate(title, body);
	},

	render() {
		return (
			<div className="discussion-create">
				<div className="form">
					<h1>Create a discussion</h1>
					<p>You are starting a new discussion on <b className="backtick" dangerouslySetInnerHTML={this.props.defName} />.</p>
					<input type="text" ref="titleText" className="title" placeholder="Title" />
					<MarkdownTextarea ref="bodyText" className="body" placeholder="Description" />
					<div className="buttons pull-right">
						<a ref="createBtn" className="btn btn-sgblue" onClick={this._create}>Create</a>
						<a ref="cancelBtn" className="btn btn-default" onClick={this.props.onCancel}>Cancel</a>
					</div>
				</div>
			</div>
		);
	},
});

module.exports = DiscussionCreateForm;
