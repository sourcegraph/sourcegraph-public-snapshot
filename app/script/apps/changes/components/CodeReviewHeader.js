var React = require("react");

var RepoBuildIndicator = require("../../../components/RepoBuildIndicator");
var CurrentUser = require("../../../CurrentUser");
var $ = require("jquery");

var CodeReviewHeader = React.createClass({

	propTypes: {
		// The object holding information about this changeset. Maps to backend's
		// sourcegraph.Changeset object.
		changeset: React.PropTypes.object.isRequired,

		// Delta object that holds detailed information about the revisions in this
		// changeset. Maps to sourcegraph.Delta.
		delta: React.PropTypes.object.isRequired,

		// Array of commits in this changeset.
		commits: React.PropTypes.object.isRequired,

		// onSubmitTitle is called after the title has been edited and submitted.
		onSubmitTitle: React.PropTypes.func,
	},

	getInitialState() {
		return {
			editing: false,
		};
	},

	componentWillUnmount() {
		this._unbindKeys();
	},

	/**
	 * @description Binds keys to the edit title input. It will cancel editing
	 * on pressing ESC and submit on Enter.
	 * @returns {void}
	 * @private
	 */
	_bindKeys() {
		var el = $(this.refs.inputTitle);

		el.keyup(e => {
			if (!this.isMounted()) return;
			switch (e.keyCode) {
			case 27: this._cancelEdit(); break;
			case 13: this._submitEdit(); break;
			}
		});
	},

	/**
	 * @description Unbinds input key events.
	 * @returns {void}
	 * @private
	 */
	_unbindKeys() {
		if (!this.isMounted()) return;
		$(this.refs.inputTitle).off("keyup");
	},

	/**
	 * @description Called when editing the title is cancelled.
	 * @returns {void}
	 * @private
	 */
	_cancelEdit() {
		this._unbindKeys();
		this.setState({editing: false});
	},

	/**
	 * @description Called when a new title is submitted. If the title is the same
	 * as the current one, no change is triggered.
	 * @returns {void}
	 * @private
	 */
	_submitEdit() {
		if (!this.isMounted()) return;

		var el = $(this.refs.inputTitle);

		if (el.val() === this.props.changeset.Title) {
			this._cancelEdit();
			return;
		}

		if (typeof this.props.onSubmitTitle === "function") {
			this.props.onSubmitTitle(this.props.changeset, el.val());
		}

		this._cancelEdit();
	},

	/**
	 * @description Triggered when the Edit icon is clicked next to the title.
	 * Displays a form to edit the title.
	 * @returns {void}
	 * @private
	 */
	_onEditClick() {
		if (!this.isMounted()) return;

		this.setState({editing: true}, () => {
			$(this.refs.inputTitle).focus();
			this._bindKeys();
		});
	},

	render() {
		return (
			<div>
				{this.state.editing ? (
					<div className="title-editing">
						<input type="text" className="input-title" ref="inputTitle" defaultValue={this.props.changeset.Title} />
						<input type="button" value="Save" className="btn-save btn btn-default" onClick={this._submitEdit} />
						<input type="button" value="Cancel" className="btn-cancel btn btn-default" onClick={this._cancelEdit} />
					</div>
				) : (
					<h1 className="changeset-title">
						<span className="changeset-id">#{this.props.changeset.ID}</span>
						{this.props.changeset.Title}
						{CurrentUser !== null && CurrentUser.Login === this.props.changeset.Author.Login ? (
							<a title="Edit" onClick={this._onEditClick} className="title-edit">
								<span className="octicon octicon-pencil"></span>
							</a>
						) : null}
					</h1>
				)}

				<div className="changeset-subtitle">
					{!this.props.changeset.ClosedAt && !this.props.changeset.Merged ? (
						<span className="changeset-status status-open selected"><span className="octicon octicon-git-pull-request"></span> OPEN</span>
					) : null}
					{this.props.changeset.ClosedAt && !this.props.changeset.Merged ? (
						<span className="changeset-status status-closed selected"><span className="octicon octicon-x"></span> CLOSED</span>
					) : null}
					<b>{this.props.changeset.Author.Login}</b> wants to merge {this.props.commits.models.length} commits from
					<div className="branch">
						{this.props.changeset.DeltaSpec.Head.Rev}
						<RepoBuildIndicator
							RepoURI={this.props.delta.HeadRepo.URI}
							Rev={this.props.delta.HeadCommit.ID}
							btnSize="btn-xs"
							Buildable={true} />
					</div>
					into
					<div className="branch">
						{this.props.changeset.DeltaSpec.Base.Rev}
						<RepoBuildIndicator
							RepoURI={this.props.delta.BaseRepo.URI}
							Rev={this.props.delta.BaseCommit.ID}
							btnSize="btn-xs" label="no"
							Buildable={true} />
					</div>

				</div>
			</div>
		);
	},
});

module.exports = CodeReviewHeader;
