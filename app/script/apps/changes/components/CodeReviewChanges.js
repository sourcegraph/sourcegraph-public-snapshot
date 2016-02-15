var React = require("react");
var $ = require("jquery");

var FileDiff = require("../../../components/FileDiffView");
var DiffFileList = require("../../../components/DiffFileList");

/**
 * @description This component holds the view of the tabs that shows the differential
 * between the base and head revision.
 */
var CodeReviewChanges = React.createClass({

	propTypes: {
		// Function is called when the expand hunk is pressed in either direction.
		// It will call the function using parameters: hunk, direction and event.
		onExpandHunk: React.PropTypes.func,

		// Function triggered when a file is clicked in the list. It receives as
		// parameters: FileDiff (Backbone.Model) and Event.
		onFileClick: React.PropTypes.func,
	},

	getInitialState() {
		return {
			changes: this.props.model.attributes,
		};
	},

	componentDidMount() {
		this.props.model.on("scrollTop", this._updateScrollPosition, this);
		this.props.model.on("add remove change", this._updateChangesState, this);
	},

	componentWillUnmount() {
		this.props.model.off("scrollTop", this._updateScrollPosition, this);
		this.props.model.off("add remove change", this._updateChangesState, this);
	},

	/**
	 * @description Callback for when the model triggers a change of scroll position.
	 * It scroll the page to the approximate vertical offset (in pixels) given by x.
	 * @param {number} x - Vertical offset in pixels to scroll page.
	 * @returns {void}
	 * @private
	 */
	_updateScrollPosition(x) {
		$("html, body").animate({scrollTop: x - 130}, 400, "linear");
	},

	/**
	 * @description Triggered when the model bound to this component changes.
	 * @returns {void}
	 * @private
	 */
	_updateChangesState() {
		this.setState({changes: this.props.model.attributes});
	},

	/**
	 * @description Triggered when the review collection changes.
	 * @returns {void}
	 * @private
	 */
	_updateReviewsState() {
		this.setState({reviews: this.props.reviews.attributes});
	},

	render() {
		if (this.state.changes.fileDiffs === null) return null;

		return (
			<div className="changeset-changes">
				{this.state.changes.overThreshold &&
					<table className="over-threshold-warning">
						<tbody>
							<tr>
								<td className="icon">
									<i className="fa fa-icon fa-warning" />
								</td>
								<td className="text">
									The requested diff is larger than usual and is surpressed. We recommend viewing it on a file-by-file basis.
									To do this, click on any of the files below. <br />
									<b>Tip:</b> You may also view groups of files by using just a prefix of the paths you wish to see.
								</td>
							</tr>
						</tbody>
					</table>
				}

				<DiffFileList {...this.props}
					model={this.state.changes.fileDiffs}
					onFileClick={this.props.onFileClick}
					stats={this.state.changes.stats} />

				{!this.state.changes.overThreshold ? (
					this.state.changes.fileDiffs.map(fd => (
						<FileDiff {...this.props}
							allowComments={true}
							key={fd.cid}
							Delta={this.state.changes.delta}
							urlBase={this.props.urlBase}
							model={fd} />
					))
				) : null}
			</div>
		);
	},
});

module.exports = CodeReviewChanges;
