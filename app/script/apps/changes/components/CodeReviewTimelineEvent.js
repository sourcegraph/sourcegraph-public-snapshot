var React = require("react");
var Backbone = require("backbone");
var ModelPropWatcherMixin = require("../../../components/mixins/ModelPropWatcherMixin");
var moment = require("moment");

var CodeReviewTimelineEvent = React.createClass({
	propTypes: {
		model: React.PropTypes.instanceOf(Backbone.Model),
	},

	mixins: [ModelPropWatcherMixin],

	render() {
		var op = this.state.Op;
		var login = op.Author.Login ? <b>{op.Author.Login}</b> : "A user";
		var msg;
		var icon = "octicon-pencil";

		if (op.Open) {
			msg = <span className="msg">re-opened the changeset</span>;
			icon = "octicon-issue-opened";
		} else if (op.Merged) {
			msg = <span className="msg">merged the changeset</span>;
			icon = "octicon-git-merge";
		} else if (op.Close) {
			msg = <span className="msg">closed the changeset</span>;
			icon = "octicon-x";
		} else if (op.Title && op.Title !== "") {
			msg = <span className="msg"> changed title to <i>"{op.Title}"</i></span>;
		} else if (op.Description && op.Description !== "") {
			msg = <span className="msg"> changed title to <i>"{op.Description}"</i></span>;
		}

		return (
			<tr className="changeset-timeline-header timeline-event">
				<td className="changeset-timeline-icon">
					<span className={`octicon ${icon}`}></span>
				</td>
				<td className="timeline-header-message">
					{login} {msg}
					<span className="date">{moment(this.state.CreatedAt).fromNow()}</span>
				</td>
			</tr>
		);
	},
});

module.exports = CodeReviewTimelineEvent;
