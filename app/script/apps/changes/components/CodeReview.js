var React = require("react");
var router = require("../../../routing/router");
var CurrentUser = require("../../../CurrentUser");
var $ = require("jquery");

var CodeReviewStore = require("../stores/CodeReviewStore");
var CodeReviewActions = require("../actions/CodeReviewActions");

var Changes = require("./CodeReviewChanges");
var Timeline = require("./CodeReviewTimeline");
var SubmitForm = require("./CodeReviewSubmitReviewForm");
var ControlPanel = require("./CodeReviewControlPanel");
var CodeReviewHeader = require("./CodeReviewHeader");

/**
 * @description CodeReview is the main component that contains all of the functionality
 * for the "Changes" application".
 */
var CodeReview = React.createClass({

	propTypes: {
		// Data may hold the JSON object that is used
		// to popuplate this component. This can be attached server-side as a
		// preloading optimization.
		data: React.PropTypes.object,
	},

	getInitialState() {
		return CodeReviewStore.attributes;
	},

	componentDidMount() {
		CodeReviewStore.on("change", () => {
			this.setState(CodeReviewStore.attributes);
		});

		CodeReviewStore.on("scrollTop", X => {
			$("html, body").animate({scrollTop: X - 130}, 400, "linear");
		});

		if (this.props.data !== null) {
			CodeReviewActions.loadData(this.props.data);
		}
	},

	componentWillUnmount() {
		CodeReviewStore.off("change");
		CodeReviewStore.off("scrollTop");
	},

	/**
	 * @description Triggers the action that shows the review form. Called on
	 * user click.
	 * @param {Event} e - The (click) event that triggered the action.
	 * @returns {void}
	 * @private
	 */
	_submitReviewShow(e) {
		if (CurrentUser === null) {
			window.location = "/login";
			return;
		}
		// TODO(gbbr): Do an action here
		CodeReviewStore.set({reviewFormVisible: true});
	},

	/**
	 * @description Called when the user cancels submitting a review.
	 * @param {Event} e - The (click) event that triggered the action.
	 * @returns {void}
	 * @private
	 */
	_submitReviewHide(e) {
		// TODO(gbbr): Do an action here
		CodeReviewStore.set({reviewFormVisible: false});
	},

	/**
	 * @description Triggers the action to submit a review on the current changeset.
	 * @param {string} body - The text body of the review.
	 * @param {Event} e - The (click) event that triggered the action.
	 * @returns {void}
	 * @private
	 */
	_submitReview(body, e) {
		CodeReviewActions.submitReview(body);
	},

	render() {
		if (typeof this.state.Changeset === "undefined") return null;
		var url = `${router.changesetURL(this.state.Changeset.DeltaSpec.Base.URI, this.state.Changeset.ID)}/files`;
		var showingGuidelines = this.state.guidelinesVisible;

		// TODO(renfred) Move this into its own component/app.
		var jiraIssues = null;
		if (this.state.JiraIssues && Object.keys(this.state.JiraIssues).length > 0) {
			var issueList = Object.keys(this.state.JiraIssues).map((id) =>
				<li>
					<a href={this.state.JiraIssues[id]}>{id}</a>
				</li>
			);

			jiraIssues = (
				<div className="well jira-issues">
					<p>JIRA Issues</p>
					<ul>
						{issueList}
					</ul>
				</div>
			);
		}

		return (
			<div className="code-review-inner">
				<CodeReviewHeader
					changeset={this.state.Changeset}
					delta={this.state.Delta}
					commits={this.state.commits}
					onSubmitTitle={CodeReviewActions.submitTitle} />

				<div className="changeset-tab-content changeset-timeline">
					<div className="left-panel">

						<Timeline
							commits={this.state.commits}
							reviews={this.state.reviews}
							events={this.state.events}
							changeset={this.state.Changeset} />

						{this.state.ReviewGuidelines && this.state.ReviewGuidelines.__html ? (
							<div className="review-guidelines">
								<i className="fa fa-warning pull-left" /> There are guidelines for contributing to this repository!
								<a
									className="pull-right"
									onClick={() => this.setState({guidelinesVisible: !Boolean(showingGuidelines)})}>
										<i className={showingGuidelines ? "octicon octicon-triangle-up" : "octicon octicon-triangle-down"} />
										{showingGuidelines ? " Hide" : " Show"}
								</a>
								{showingGuidelines ? (
									<div className="markdown-view" dangerouslySetInnerHTML={this.state.ReviewGuidelines} />
								) : null}
							</div>
						) : null}

						<SubmitForm
							visible={this.state.reviewFormVisible}
							submitDisabled={this.state.submittingReview}
							drafts={this.state.reviews.drafts}
							onShow={this._submitReviewShow}
							onSubmit={this._submitReview}
							onCancel={this._submitReviewHide} />
					</div>

					<div className="right-panel">
						<ControlPanel
							changeset={this.state.Changeset}
							onStatusChange={CodeReviewActions.changeChangesetStatus}
							merging={this.state.merging}
							onMerge={CodeReviewActions.mergeChangeset} />
						{jiraIssues}
					</div>
				</div>

				<div className="changeset-tab-content tab-changes">
					{this.state.FileFilter ? (
						<div className="filter-warning">
							<i className="fa fa-icon fa-warning" />
							<span>Currently there is a filter applied to this view (<i className="backtick">{this.state.FileFilter}</i>). To clear it, you may <a href={url}>click here</a>.</span>
						</div>
					) : null}

					<Changes
						onTokenFocus={CodeReviewActions.focusToken}
						onCommentSubmit={CodeReviewActions.saveDraft}
						onCommentEdit={CodeReviewActions.updateDraft}
						onCommentDelete={CodeReviewActions.deleteDraft}
						onTokenBlur={CodeReviewActions.blurTokens}
						onTokenClick={CodeReviewActions.selectToken}
						onExpandHunk={CodeReviewActions.expandHunk}
						onFileClick={CodeReviewActions.selectFile}
						model={this.state.changes}
						urlBase={url}
						reviews={this.state.reviews} />
				</div>
			</div>
		);
	},
});

module.exports = CodeReview;
