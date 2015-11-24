var React = require("react");
var globals = require("../../../globals");
var classnames = require("classnames");

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

		// onMerge is the function called when the user triggers an automatic
		// merge.
		onMerge: React.PropTypes.func,

		merging: React.PropTypes.bool,
	},

	/**
	 * @description Triggered when the user clicks on the merge button.
	 * @param {Event} evt - The (click) event.
	 * @returns {void}
	 * @private
	 */
	_merge(evt) {
		if (this.props.changeset.ClosedAt || this.props.changeset.Merged) return;

		// TODO(renfred) use better UI element instead of native browser popup.
		if (!confirm(`Merge branch ${this.props.changeset.DeltaSpec.Head.Rev} into ${this.props.changeset.DeltaSpec.Base.Rev}?`)) return;

		evt.currentTarget.blur(); // Unfocus merge button.
		var opt = {
			squash: this.refs.squashOption.checked,
		};
		this.props.onMerge(opt, evt);
	},

	/**
	 * @description Triggered when the user clicks on an inactive changeset status.
	 * @param {Event} evt - The (click) event.
	 * @returns {void}
	 * @private
	 */
	_changeStatus(evt) {
		var status = this.refs.statusOptions.value;

		if (typeof this.props.onStatusChange === "function") {
			this.props.onStatusChange(status, evt);
		}
	},

	render() {
		var activeValue = globals.ChangesetStatus.OPEN;

		if (this.props.changeset.ClosedAt) {
			activeValue = !this.props.changeset.Merged ? globals.ChangesetStatus.CLOSED : globals.ChangesetStatus.CLOSED;
		}

		var mergeIcon = this.props.merging ? <i className="fa fa-circle-o-notch fa-spin"></i> :
			<span className="octicon octicon-git-merge"></span>;
		var mergeButtonClasses = classnames({
			"btn": true,
			"btn-secondary": true,
			"disabled": this.props.merging || this.props.changeset.ClosedAt,
		});

		return (
			<div className="changeset-control-panel">
				<div className="status-form">
					<div className="panel-label">Status</div>
					<select ref="statusOptions" className="status-options" value={activeValue} onChange={this._changeStatus}>
						<option value={globals.ChangesetStatus.OPEN}>Open</option>
						<option value={globals.ChangesetStatus.CLOSED}>Closed</option>
					</select>
				</div>
				{!this.props.changeset.Merged &&
					<div className="merge-form">
						<div className="panel-label">Automatic Merge</div>
						<label><input ref="squashOption" type="checkbox"/> Squash commits on merge</label>
						<button className={mergeButtonClasses} onClick={this._merge}>
							{mergeIcon}
							<span> Merge changeset</span>
						</button>
					</div>
				}
			</div>
		);
	},
});

module.exports = CodeReviewControlPanel;
