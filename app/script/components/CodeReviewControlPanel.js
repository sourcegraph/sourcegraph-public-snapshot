var React = require("react");
var globals = require("../globals");
var $ = require("jquery");

/**
 * @description CodeReviewControlPanel holds the view that shows general information
 * about the status and state of the changeset, as well as allows controling and changing
 * it
 */
var CodeReviewControlPanel = React.createClass({

	propTypes: {
		// The object holding information about this changeset. Maps to backend's
		// sourcegraph.Changeset object.
		changeset: React.PropTypes.object.isRequired,

		// onStatusChange is the function called when the user triggers a change
		// in the status of the changeset.
		onStatusChange: React.PropTypes.func,
	},

	/**
	 * @description Triggered when the user clicks on an inactive changeset status.
	 * @param {Event} evt - The (click) event.
	 * @returns {void}
	 * @private
	 */
	_changeStatus(evt) {
		var status = $(this.getDOMNode()).find(".status-options").val();

		if (typeof this.props.onStatusChange === "function") {
			this.props.onStatusChange(status, evt);
		}
	},

	render() {
		var activeValue = globals.ChangesetStatus.OPEN;

		if (this.props.changeset.ClosedAt) {
			activeValue = !this.props.changeset.Merged ? globals.ChangesetStatus.CLOSED : globals.ChangesetStatus.CLOSED;
		}

		return (
			<div className="changeset-control-panel">
				<div className="panel-label">Status</div>
				<select className="status-options" value={activeValue} onChange={this._changeStatus}>
					<option value={globals.ChangesetStatus.OPEN}>Open</option>
					<option value={globals.ChangesetStatus.CLOSED}>Closed</option>
				</select>
			</div>
		);
	},
});

module.exports = CodeReviewControlPanel;
