var React = require("react");
var CommitModel = require("../stores/models/CommitModel");
var ModelPropWatcherMixin = require("./mixins/ModelPropWatcherMixin");
var moment = require("moment");
var router = require("../routing/router");
var MarkdownView = require("./MarkdownView");

var CodeReviewTimelineCommit = React.createClass({
	propTypes: {
		model: React.PropTypes.instanceOf(CommitModel),
	},

	mixins: [ModelPropWatcherMixin],

	render() {
		var url = `${router.repoURL(this.state.RepoURI)}/.commits/${this.state.ID}`;

		return (
			<tr className="changeset-timeline-header timeline-commit">
				<td className="changeset-timeline-icon">
					<span className="octicon octicon-git-commit"></span>
				</td>

				<td colSpan="3">
					<div className="header">
						<b>{this.state.Author.Name}</b> committed <a className="commit-id" href={url}>{this.state.ID.substring(0, 7)}</a> <span className="date">{moment(this.state.Author.Date).fromNow()}</span>
					</div>
					<MarkdownView content={this.state.Message} />
				</td>
			</tr>
		);
	},
});

module.exports = CodeReviewTimelineCommit;
