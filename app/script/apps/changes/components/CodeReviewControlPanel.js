var React = require("react");
var globals = require("../../../globals");
var classnames = require("classnames");
var Tooltip = require("../../../../web_modules/sourcegraph/util/Tooltip").default;

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

		// onLGTMChange is the function called when the user changes their LGTM
		// status on the changeset.
		onLGTMChange: React.PropTypes.func,

		// onAddReviewer is the function called when a user tries to add a reviewer
		// to the changeset. Provided as a parameter is the username to be added.
		onAddReviewer: React.PropTypes.func,

		// onRemoveReviewer is the function called when a user tries to remove a
		// reviewer from the changeset. Provided as a parameter is the reviewer's
		// User object.
		onRemoveReviewer: React.PropTypes.func,
	},

	getInitialState: function() {
		return {showAddPersonMenu: false};
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

	_toggleAddPersonMenu(evt) {
		this.setState({showAddPersonMenu: !this.state.showAddPersonMenu});
	},

	_onEnterReviewer(evt) {
		if (evt.keyCode !== 13) {
			return;
		}
		if(this.props.onAddReviewer) {
			this.props.onAddReviewer(evt.target.value);
			this.setState({showAddPersonMenu: false});
		}
	},

	_toggleLGTM(evt) {
		if(this.props.onLGTMChange) {
			this.props.onLGTMChange(evt.currentTarget.checked);
		}
	},

	_removeReviewer(reviewer, evt) {
		if(this.props.onRemoveReviewer) {
			this.props.onRemoveReviewer(reviewer);
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

		var reviewers = [
			{Login: "slimsag", AvatarURL: "https://image.freepik.com/free-photo/cat-face-viewed-from-front_21220647.jpg"},
			{Login: "richard", AvatarURL: "http://media4.popsugar-assets.com/files/2014/09/19/978/n/1922507/4bc5042ee37fa1f9_thumb_temp_cover_file13465311411161397.xxxlarge/i/Funny-Cat-Costumes.jpg"},
			{Login: "renfred", AvatarURL: "http://hdwallpapersfit.com/wp-content/uploads/2015/03/white-cat-face_wallpaper.jpg"},
			{Login: "john"}
		];

		var addPersonMenu = null;
		if (this.state.showAddPersonMenu) {
			addPersonMenu = (
				<div className="add-person-menu" >
					<input type="text" placeholder="Username" onKeyDown={this._onEnterReviewer} autoFocus="true"></input>
				</div>
			);
		}

		return (
			<div className="changeset-control-panel">
				<div className="status-form">
					<div className="panel-label">Status</div>
					<select ref="statusOptions" className="status-options" value={activeValue} onChange={this._changeStatus}>
						<option value={globals.ChangesetStatus.OPEN}>Open</option>
						<option value={globals.ChangesetStatus.CLOSED}>Closed</option>
					</select>
				</div>

				<hr/>

				<div className="review-form">
					<div className="panel-label">Reviewers</div>
					<div className="people">
						{reviewers.map(function(reviewer){
							var classes = null;
							var styles = {};
							if (reviewer.AvatarURL) {
								styles["backgroundImage"] = "url(" + reviewer.AvatarURL + ")";
								classes = classnames({"reviewer": true, "avatar": true});
							} else {
								classes = classnames({"reviewer": true, "octicon": true, "octicon-person": true});
							}

							return (
								<span className={classes} key={reviewer.Login} style={styles}>
									<div className="remove">
										<button className="btn btn-secondary" onClick={this._removeReviewer.bind(this, reviewer)}>
											<span className="octicon octicon-x"></span>
										</button>
									</div>

									<Tooltip><b>{reviewer.Login}</b></Tooltip>
								</span>
							);
						}, this)}

						<button className="add-person btn btn-secondary" data-toggle="button" style={{position: "relative"}} onClick={this._toggleAddPersonMenu}>
							<Tooltip><b>add a reviewer</b></Tooltip>
							<span className="octicon octicon-plus"></span>
						</button>

						{addPersonMenu}
					</div>
					<label className="btn btn-secondary active">
						<input type="checkbox" autoComplete="off" onClick={this._toggleLGTM}></input> LGTM
					</label>
				</div>

				<hr/>

				<div className="merge-form">
					<div className="panel-label">Automatic Merge</div>
					<label><input ref="squashOption" type="checkbox"/> Squash commits on merge</label>
					<button className={mergeButtonClasses} onClick={this._merge}>
						{mergeIcon}
						<span> Merge changeset</span>
					</button>
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
