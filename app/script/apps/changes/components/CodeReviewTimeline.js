var React = require("react");
var Backbone = require("backbone");
var moment = require("moment");

var TimelineCommit = require("./CodeReviewTimelineCommit");
var TimelineReview = require("./CodeReviewTimelineReview");
var TimelineEvent = require("./CodeReviewTimelineEvent");
var MarkdownView = require("../../../components/MarkdownView");

/**
 * @description CodeReviewTimeline holds the Timeline on the first tab of the
 * Code Review functionality
 */
var CodeReviewTimeline = React.createClass({

	propTypes: {
		// The collection of commits that will be shown on the timeline.
		commits: React.PropTypes.instanceOf(Backbone.Collection).isRequired,

		// The events that occurred on the changeset (open, closing, etc) to be
		// shown on the timeline.
		events: React.PropTypes.instanceOf(Backbone.Collection).isRequired,

		// The reviews that were placed on this changeset.
		reviews: React.PropTypes.instanceOf(Backbone.Collection).isRequired,

		// Information about the changeset. Maps to sourcegraph.Changeset object.
		changeset: React.PropTypes.object.isRequired,
	},

	getInitialState() {
		return {
			commits: this.props.commits.models,
			reviews: this.props.reviews,
			events: this.props.events,
		};
	},

	componentDidMount() {
		this.props.commits.on("add remove change", this._updateCommitsState, this);
		this.props.reviews.on("add remove change", this._updateReviewsState, this);
		this.props.events.on("add remove change", this._updateEventsState, this);
	},

	componentWillUnmount() {
		this.props.commits.off("add remove change", this._updateCommitsState, this);
		this.props.reviews.off("add remove change", this._updateReviewsState, this);
		this.props.events.off("add remove change", this._updateEventsState, this);
	},

	/**
	 * @description Function bound to changes on the reviews collection property.
	 * @returns {void}
	 * @private
	 */
	_updateReviewsState() {
		this.setState({reviews: this.props.reviews});
	},

	/**
	 * @description Function bound to changes on the commits collection property.
	 * @returns {void}
	 * @private
	 */
	_updateCommitsState() {
		this.setState({commits: this.props.commits.models});
	},

	/**
	 * @description Function bound to changes on the events collection property.
	 * @returns {void}
	 * @private
	 */
	_updateEventsState() {
		this.setState({events: this.props.events});
	},

	/**
	 * @description Sorts an array of events by date. These events may be commits,
	 * reviews or other events. It is assumed that they contain a valid date as
	 * a property named "CreatedAt" or "Author.Date". The function uses the built-in
	 * sort, thus altering the original object. It has no return value.
	 * @param {Array<Object>} arrayOfStuff - Events to be sorted
	 * @returns {void}
	 * @private
	 */
	_sortByDate(arrayOfStuff) {
		arrayOfStuff.sort((x, y) => {
			var ma = x.props.model;
			var mb = y.props.model;
			var a = ma.get("CreatedAt") || ma.get("Author").Date;
			var b = mb.get("CreatedAt") || mb.get("Author").Date;

			return (new Date(a).getTime()) - (new Date(b).getTime());
		});
	},

	render() {
		if (this.state.commits.length === 0) return null;

		var all = [].concat(
			this.state.commits.map(commit => <TimelineCommit key={commit.cid} model={commit} />)).concat(
			this.state.reviews.map(review => <TimelineReview key={review.cid} model={review} />)).concat(
			this.state.events.map(event => <TimelineEvent key={event.cid} model={event} />)
		);

		this._sortByDate(all);

		return (
			<div className="changeset-history">
				<table className="changeset-timeline-block changeset-description">
					<tbody>
						<tr className="changeset-timeline-header">
							<td className="changeset-timeline-icon">
								<span className="fa fa-flag"></span>
							</td>
							<td className="timeline-header-message">
								<b>{this.props.changeset.Author.Login}</b> started this changeset<span className="date">{moment(this.props.changeset.CreatedAt).fromNow()}</span>
								<br /><MarkdownView content={this.props.changeset.Description} />
							</td>
						</tr>
					</tbody>
				</table>

				<table className="changeset-timeline-block">
					<tbody>{all}</tbody>
				</table>
			</div>
		);
	},
});

module.exports = CodeReviewTimeline;
