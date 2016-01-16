var React = require("react");
var $ = require("jquery");
var router = require("../routing/router");

var DiffStore = require("../stores/DiffStore");
var DiffActions = require("../actions/DiffActions");
var CurrentUser = require("../CurrentUser");

var DiffPopup = require("./DiffPopupView");
var DiffFileList = require("./DiffFileList");
var DiffProposeChangeForm = require("./DiffProposeChangeForm");
var FileDiff = require("./FileDiffView");
var TokenPopover = require("./TokenPopoverView");
var RepoBuildIndicator = require("./RepoBuildIndicator");
var RepoRevSwitcher = require("./RepoRevSwitcher");

/*
 * @description CompareView displays a series of file diffs.
 */
var CompareView = React.createClass({

	propTypes: {
		// Data may hold a string representation of the JSON object that is used
		// to popuplate this component. This will be attached server-side as a
		// preloading optimization.
		data: React.PropTypes.string,
	},

	getInitialState() {
		return DiffStore.attributes;
	},

	componentDidMount() {
		DiffStore.on("change", () => this.replaceState(DiffStore.attributes));
		DiffStore.on("scrollTop", X => $("html, body").animate({scrollTop: X - 130}, 400, "linear"));
		if (this.props.data !== null) DiffActions.loadData(JSON.parse(this.props.data));
	},

	componentWillUnmount() {
		DiffStore.off("change");
		DiffStore.off("scrollTop");
	},

	_onExpandHunk(hunk, isDirectionUp, evt) {
		if (isDirectionUp) {
			DiffActions.expandHunkUp(hunk, evt);
		} else {
			DiffActions.expandHunkDown(hunk, evt);
		}
	},

	/*
	 * @description Callback when tokens in the compare view are focused.
	 * @param {object} fd - FileDiff that contains the token
	 * @param {CodeTokenModel} token - Focused token
	 * @param {Event} evt - action event
	 * @private
	 */
	_onFileDiffTokenFocus(fd, token, evt) {
		DiffActions.focusToken(token, evt, fd);
	},

	/*
	 * @description Redirects the user to a the page that has the selected revision as
	 * the new base. This method is a callback of using the revision switcher dropdown.
	 * @param {string} repoSpec - Repository URL
	 * @param {string} rev - Reivision
	 * @private
	 */
	_changeBaseBranch(repoSpec, rev) {
		window.location = router.compareURL(repoSpec, rev, this.state.DeltaSpec.Head.Rev);
	},

	/*
	 * @description Redirects the user to a the page that has the selected revision as
	 * the new head. This method is a callback of using the revision switcher dropdown.
	 * @param {string} repoSpec - Repository URL
	 * @param {string} rev - Reivision
	 * @private
	 */
	_changeHeadBranch(repoSpec, rev) {
		window.location = router.compareURL(repoSpec, this.state.DeltaSpec.Base.Rev, rev);
	},

	_openProposeChangeForm() {
		this.setState({proposingChange: true});
	},

	_onFileClick(fd, evt) {
		var baseName = fd.get("OrigName");
		baseName = baseName === "/dev/null" ? undefined : baseName;
		var headName = fd.get("NewName");
		headName = headName === "/dev/null" ? undefined : headName;
		var delta = this.state.DeltaSpec;
		var url = router.compareURL(this.state.RepoRevSpec.URI, delta.Base.Rev, delta.Head.Rev, headName||baseName);

		if (this.state.OverThreshold) {
			window.location = url;
			return;
		}

		DiffActions.selectFile(fd, evt);
	},

	render() {
		if (typeof this.state.fileDiffs === "undefined") return null;

		var displayedDiffs = typeof this.state.filter === "object" ?
			this.state.fileDiffs.where(this.state.filter) : this.state.fileDiffs;

		return (
			<div className="compare-view">
				{this.props.revisionHeader === "yes" ? (
					<header>
						<div className="compare-icon octicon octicon-git-compare" />

						<RepoRevSwitcher
							repoSpec={this.state.DeltaSpec.Base.RepoSpec.URI}
							rev={this.state.DeltaSpec.Base.Rev}
							onBranchSelect={this._changeBaseBranch}
							label="base:" />

						<RepoBuildIndicator
							RepoURI={this.state.DiffData.Delta.Base.RepoSpec.URI}
							CommitID={this.state.DiffData.Delta.BaseCommit.ID}
							btnSize="btn-xs" />

						<span className="separator">...</span>

						<RepoRevSwitcher
							repoSpec={this.state.DeltaSpec.Head.RepoSpec.URI}
							rev={this.state.DeltaSpec.Head.Rev}
							onBranchSelect={this._changeHeadBranch}
							label="head:" />

						<RepoBuildIndicator
							RepoURI={this.state.DiffData.Delta.Head.RepoSpec.URI}
							CommitID={this.state.DiffData.Delta.HeadCommit.ID}
							btnSize="btn-xs" />

						{CurrentUser !== null && this.state.fileDiffs.length && !this.state.proposingChange ? (
							<a href="#" className="btn btn-sgblue pull-right" onClick={this._openProposeChangeForm}>
								<span>Propose this change</span>
							</a>
						) : null}

						{this.state.DeltaSpec.Base.CommitID !== this.state.DiffData.Delta.BaseCommit.ID ? (
							<span className="pull-right warning">
								&nbsp;merge-base is {this.state.DiffData.Delta.BaseCommit.ID.substring(0, 7)}
								<i className="fa fa-icon fa-warning merge-base-warning" />
							</span>
						) : null}
					</header>
				) : null}

				{this.state.proposingChange ? (
					<DiffProposeChangeForm
						deltaSpec={this.state.DeltaSpec}
						loading={this.state.changesetLoading}
						closed={!this.state.proposingChange}
						onCancel={()=>this.setState({proposingChange: false})} />
				) : null}

				{this.state.OverThreshold &&
					<table className="over-threshold-warning">
						<tbody>
							<td className="icon">
								<i className="fa fa-icon fa-warning" />
							</td>
							<td className="text">
								The requested diff is larger than usual and is surpressed. We recommend viewing it on a file-by-file basis.
								To do this, click on any of the files below. <br /><b>Tip:</b> You may also view smaller groups of files using the <span className="backtick">filter</span> query parameter.
							</td>
						</tbody>
					</table>
				}

				<DiffFileList
					model={this.state.fileDiffs}
					stats={this.state.DiffData.Stats}
					onFileClick={this._onFileClick} />

				<TokenPopover model={this.state.popoverModel} />

				<DiffPopup
					model={this.state.popupModel}
					onClose={DiffActions.closePopup} />

				{!this.state.OverThreshold ? displayedDiffs.map(fd => (
					<FileDiff
						key={fd.cid}
						Delta={this.state.DiffData.Delta}
						model={fd}
						onTokenFocus={this._onFileDiffTokenFocus.bind(this, fd)}
						onTokenBlur={DiffActions.blurTokens.bind(this, fd)}
						onTokenClick={DiffActions.selectToken}
						onExpandHunk={this._onExpandHunk} />
				)) : null}
			</div>
		);
	},
});

module.exports = CompareView;
