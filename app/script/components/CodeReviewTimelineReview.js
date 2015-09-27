var React = require("react");
var Backbone = require("backbone");
var ModelPropWatcherMixin = require("./mixins/ModelPropWatcherMixin");
var moment = require("moment");
var MarkdownView = require("./MarkdownView");
var CodeReviewActions = require("../actions/CodeReviewActions");

var CodeReviewTimelineReview = React.createClass({

	propTypes: {
		model: React.PropTypes.instanceOf(Backbone.Model),
	},

	mixins: [ModelPropWatcherMixin],

	_commentsByFile(comments) {
		var byFile = {};

		(comments || []).forEach(comment => {
			if (!Array.isArray(byFile[comment.get("Filename")])) byFile[comment.get("Filename")] = [];
			byFile[comment.get("Filename")].push(comment);
		});

		if (Object.keys(byFile).length === 0) return null;

		return Object.keys(byFile).map((filename, i) => {
			var commentGroup = [
				<div key={`file-${i}`} className="file-name">
					<i className="fa fa-file-text-o"></i> {filename}
				</div>,
			];

			commentGroup.push((
				<table className="comment-group" key={`comment-group-${filename}-${i}`}>
					{byFile[filename].map((comment, j) => (
						<tr className="comment" key={`review-comment-${i}-${j}`}>
							<td className="comment-line-number" onClick={CodeReviewActions.showComment.bind(this, comment)}>
								<i className="fa fa-reply"></i> {comment.get("LineNumber")}
							</td>
							<td className="comment-body">
								<div className="comment-body-inner">
									<MarkdownView content={comment.get("Body")} />
								</div>
							</td>
						</tr>
					))}
				</table>
			));

			return (
				<div key={`review-group-${filename}`} className="comment-file-group">{commentGroup}</div>
			);
		});
	},

	render() {
		return (
			<tr className="changeset-timeline-header timeline-review">
				<td className="changeset-timeline-icon">
					<span className="octicon octicon-comment"></span>
				</td>
				<td colSpan="3" className="timeline-header-message">
					<b>{this.state.Author.Login}</b> reviewed<span className="date">{moment(this.state.CreatedAt).fromNow()}</span>
					<MarkdownView content={this.state.Body} />
					{this._commentsByFile(this.state.Comments)}
				</td>
			</tr>
		);
	},
});

module.exports = CodeReviewTimelineReview;
