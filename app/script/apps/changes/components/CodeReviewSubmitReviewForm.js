var React = require("react");
var Backbone = require("backbone");
var MarkdownTextarea = require("../../../components/MarkdownTextarea");

/**
 * @description This component holds the view that contains the review submission
 * form.
 */
var CodeReviewSubmitReviewForm = React.createClass({

	propTypes: {
		// visible indicates whether the review form is hidden or shown.
		// This property may be externally altered when the 'onShow'
		// function is called.
		visible: React.PropTypes.bool.isRequired,

		// drafts is a backbone collection of inline comments that need
		// to be submitted along with this review. This property is used
		// to display how many drafts will be submitted with this review.
		drafts: React.PropTypes.instanceOf(Backbone.Collection),

		// onShow will be called when the button that is expected to show
		// the form is clicked. The expected behavior of this function is
		// to externally alter the 'visible' prop.
		onShow: React.PropTypes.func.isRequired,

		// onSubmit will be called when the review is submitted. It will
		// receive parameters 'body' and 'event', where 'body' is the text
		// that the user entered and 'event' is the click event.
		onSubmit: React.PropTypes.func.isRequired,

		// onCancel is the function that will be called when the review
		// submission is cancelled.
		onCancel: React.PropTypes.func.isRequired,
	},

	/**
	 * @description Triggered when the review is submitted.
	 * @param {Event} e - Event
	 * @returns {void}
	 * @private
	 */
	_submit(e) {
		this.props.onSubmit(this.refs.formBody.value(), e);
	},

	render() {
		return (
			<table className="changeset-timeline-block changeset-submit-review">
				<tbody>
					{this.props.visible ? (
						<tr className="changeset-review-submit-form">
							<td colSpan={2}>
								<MarkdownTextarea ref="formBody" placeholder="Enter a description..." />
								<div className="actions">
									<i className="pull-left">Includes {this.props.drafts.length} inline comments.</i>
									<a className="btn btn-success" onClick={this._submit} tabIndex="0">Submit</a>
									<a className="btn" onClick={this.props.onCancel} tabIndex="0">Cancel</a>
								</div>
							</td>
						</tr>
					) : (
						<tr className="changeset-timeline-header" onClick={this.props.onShow}>
							<td className="changeset-timeline-icon changeset-icon-submit">
								<a>
									<span className="octicon octicon-plus"></span>
								</a>
							</td>
							<td colSpan="3" className="timeline-header-message">
								Submit your review
							</td>
						</tr>
					)}
				</tbody>
			</table>
		);
	},
});

module.exports = CodeReviewSubmitReviewForm;
