var React = require("react");
var CurrentUser = require("../CurrentUser");

var ModelPropWatcherMixin = require("./mixins/ModelPropWatcherMixin");

var CodeLine = require("./CodeLineView");
var CodeReviewComment = require("./CodeReviewComment");
var ReviewCollection = require("../stores/collections/ReviewCollection");

var HunkView = React.createClass({

	propTypes: {
		// Token event callback.
		// The function to be called on click. It will receive as arguments the
		// CodeTokenModel that was clicked and the event. Default is automatically
		// prevented.
		onTokenClick: React.PropTypes.func,

		// Token event callback.
		// The function to be called on 'mouseenter'. It will receive as arguments the
		// CodeTokenModel and the event. Default is automatically prevented.
		onTokenFocus: React.PropTypes.func,

		// Token event callback.
		// The function to be called on 'mouseleave'. It will receive as arguments the
		// CodeTokenModel and the event. Default is automatically prevented.
		onTokenBlur: React.PropTypes.func,

		// Function is called when the expand hunk is pressed in either direction.
		// It will call the function using parameters: hunk, direction and event.
		onExpandHunk: React.PropTypes.func,

		// allowComments will display the comment '+' button next to each line of code.
		allowComments: React.PropTypes.bool,

		// onCommentSubmit is triggered when a comment is submitted. It is passed
		// hunk model, line model, body (string) and event.
		onCommentSubmit: React.PropTypes.func,

		// onCommentEdit is triggered when a comment is edited. It is passed
		// hunk model, line model, comment model, the new body (string) and event.
		onCommentEdit: React.PropTypes.func,

		// onCommentDelete is triggered when a comment is deleted. It is passed
		// hunk model, line model, comment model and event.
		onCommentDelete: React.PropTypes.func,

		// reviews contains a reference to the collection owning all of the reviews for
		// the changeset that his hunk is part of.
		reviews: React.PropTypes.instanceOf(ReviewCollection),
	},

	mixins: [ModelPropWatcherMixin],

	_onHunkExpand(isDirectionUp, evt) {
		if (typeof this.props.onExpandHunk === "function") {
			this.props.onExpandHunk(this.props.model, isDirectionUp, evt);
		}
	},

	_onCommentOpen(line, evt) {
		if (CurrentUser === null) {
			window.location = "/login";
			return;
		}
		this.props.model.openComment(line);
		evt.preventDefault();
	},

	_onCommentClose(line, evt) {
		if (CurrentUser === null) {
			window.location = "/login";
			return;
		}
		this.props.model.closeComment(line);
		evt.preventDefault();
	},

	_onCommentEdit(line, comment, newBody, evt) {
		if (comment.get("Body") === newBody) return;

		if (typeof this.props.onCommentEdit === "function") {
			this.props.onCommentEdit(this.props.model, line, comment, newBody, evt);
		}
	},

	_onCommentDelete(line, comment, evt) {
		if (typeof this.props.onCommentDelete === "function") {
			this.props.onCommentDelete(this.props.model, line, comment, evt);
		}
	},

	_onCommentSubmit(line, body, evt) {
		if (typeof this.props.onCommentSubmit === "function") {
			this.props.onCommentSubmit(this.props.model, line, body, evt);
		}
	},

	/**
	 * @description Returns the JSX needed to display a line's comment thread,
	 * and optionally its comment submit form.
	 * @param {CodeLineModel} line - The line to which the thread belongs to.
	 * @param {Array<CommentModel>} comments - The comments to be displayed.
	 * @param {bool} withForm - Whether to attach the comment submit form.
	 * @return {JSX} JSX for the thread and/or form.
	 */
	_commentThread(line, comments, withForm) {
		var items = [
			comments.map((comment, i) => (
				<div className="comment-container" key={`${line.cid}-${comment.cid}`}>
					<CodeReviewComment
						comment={comment}
						onEdit={this._onCommentEdit.bind(this, line, comment)}
						onDelete={this._onCommentDelete.bind(this, line, comment)} />

					{/*
						If this is the last comment, and the comment submit form is closed,
						we display a button to add a note.
					*/}
					{(i === comments.length - 1) && !withForm ? (
						<div className="comment-add-note">
							<a className="btn btn-default" onClick={this._onCommentOpen.bind(this, line)}>
								<span className="octicon octicon-plus"></span> Add a note
							</a>
						</div>
					) : null}
				</div>
			)),
		];

		if (withForm) {
			items.push((
				<CodeReviewComment
					draftForm={true}
					onCancel={this._onCommentClose.bind(this, line)}
					onSubmit={this._onCommentSubmit.bind(this, line)} />
			));
		}

		return <tr><td colSpan={3}>{items}</td></tr>;
	},

	render() {
		// hasChanges is true if this hunk has changes. We can recognise that
		// a hunk has changes by checking if it has both additions and deletions.
		var hasChanges = this.state.NewLines && this.state.OrigLines;

		// These will be true if we need to display one or both buttons that
		// retrieve more context lines at the top or bottom of this hunk.
		var hasExpandTop = hasChanges && this.state.expandTop && this.state.OrigStartLine > 1;
		var hasExpandBottom = hasChanges && this.state.expandBottom;

		// hasHeader will be true if we want to display header of he hunk.
		// This is used to simulate the merge of two hunks by hiding any
		// separators. This happens when the user expands context from
		// a hunk into another.
		var hasHeader = this.state.header === true;

		return (
			<table className="line-numbered-code theme-default file-diff-hunk">
				<tbody>
					{hasHeader ? (
						<tr className="line hunk-header">
							<td className="line-number">...</td>
							<td className="line-number">...</td>
							<td className="line-content">
								@@ -{this.state.OrigStartLine},{this.state.OrigLines} +{this.state.NewStartLine},{this.state.NewLines} @@ {this.state.Section}
							</td>
						</tr>
					) : null}

					{hasExpandTop ? (
						<tr>
							<td className="expand expand-top"
								colSpan={3}
								onClick={this._onHunkExpand.bind(this, true)}>
								<i className="fa fa-icon fa-caret-up" />
							</td>
						</tr>
					) : null}

					{this.state.Lines.map(line => {
						// withForm will be true if we have to show an open comment form
						// on this line.
						var withForm = Array.isArray(this.state.commentsAt) && this.state.commentsAt.indexOf(line.cid) > -1;

						var retLine = [
							<CodeLine {...this.props}
								onComment={this._onCommentOpen}
								key={line.cid}
								lineNumbers={false}
								model={line}
								allowComments={this.props.allowComments && !withForm} />,
						];

						// if a review collection is linked, comments contains the
						// inline review comments for this line (if any).
						var comments = (this.props.reviews ? this.props.reviews.getLineComments(line) : []);

						if (comments.length || withForm) {
							retLine.push(this._commentThread(line, comments, withForm));
						}

						return retLine;
					})}

					{hasExpandBottom ? (
						<tr>
							<td className="expand expand-bottom"
								colSpan={3}
								onClick={this._onHunkExpand.bind(this, false)}>
								<i className="fa fa-icon fa-caret-down" />
							</td>
						</tr>
					) : null}
				</tbody>
			</table>
		);
	},
});

module.exports = HunkView;
