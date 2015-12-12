var React = require("react");
var client = require("../client");

var RepoBuildIndicator = React.createClass({

	propTypes: {
		// SuccessReload will cause the page to reload when the build becomes
		// successful. The option will be enabled if the prop is set, no matter
		// its value.
		SuccessReload: React.PropTypes.string,

		// RepoURI represents the URI of the repository that we are checking
		// build data for.
		RepoURI: React.PropTypes.string,

		// Rev sets the revision for which we are checking build information.
		Rev: React.PropTypes.string,

		// Buildable is whether or not the RepoBuildIndicator will let the
		// user trigger a build if a build does not exist.
		Buildable: React.PropTypes.bool,
	},

	getDefaultProps() {
		return {
			tooltipPosition: "top",
			Buildable: false,
		};
	},

	getInitialState() {
		return {
			LastBuild: this.props.LastBuild,
			status: this._getBuildStatus(this.props.LastBuild),
		};
	},

	componentDidMount() {
		if (this.state.status === this.BuildStatus.UNKNOWN) {
			this.checkBuildStatus();
		}
	},

	componentWillUnmount() {
		clearInterval(this.interval);
	},

	// BuildStatus indicates the current status of the indicator.
	BuildStatus: {
		FAILURE: "FAILURE",
		BUILT: "BUILT",
		STARTED: "STARTED",
		QUEUED: "QUEUED",
		NA: "NOT_AVAILABLE",
		ERROR: "ERROR",
		UNKNOWN: "UNKNOWN",
	},

	// getBuildStatus returns the status appropriate for the given build data.
	_getBuildStatus(buildData) {
		if (typeof buildData === "undefined") {
			return this.BuildStatus.UNKNOWN;
		}
		if (Array.isArray(buildData) && buildData.length === 0 || buildData === null) {
			return this.BuildStatus.NA;
		}
		if (buildData.Failure) {
			return this.BuildStatus.FAILURE;
		}
		if (buildData.Success) {
			return this.BuildStatus.BUILT;
		}
		if (buildData.StartedAt && !buildData.EndedAt) {
			return this.BuildStatus.STARTED;
		}
		return this.BuildStatus.QUEUED;
	},

	// PollSpeeds holds the intervals at which to poll for updates (ms).
	// Keys that are not present will cause no polling.
	PollSpeeds: {
		STARTED: 5000,
		QUEUED: 10000,
	},

	_updatePoller() {
		clearInterval(this.interval);
		var freq = this.PollSpeeds[this.state.status] || 0;
		if (freq) {
			this.interval = setInterval(this.checkBuildStatus, freq);
		}
	},

	// _updateBuild updates the component's state based on new LastBuild data.
	// If the data argument is an Array of builds, the one at index 0 is used.
	_updateBuildData(data) {
		this.setState({LastBuild: data || null, status: this._getBuildStatus(data)});
	},

	// _handleError handles network errors
	_updateBuildDataError(err) {
		this.setState({LastBuild: null, status: this.BuildStatus.ERROR});
	},

	checkBuildStatus() {
		client.builds(this.props.RepoURI, this.props.Rev, this.state.noCache)
			.then(
				data => this._updateBuildData(data && data.Builds ? data.Builds[0] : null),
				this._updateBuildDataError
			);
	},

	triggerBuild(ev) {
		this.setState({noCache: true}); // Otherwise after creating the build, API responses still show the prior state.
		client.createRepoBuild(this.props.RepoURI, this.props.Rev)
			.then(this._updateBuildData, this._updateBuildDataError);
	},

	render() {
		this._updatePoller();
		if (this.state.status === this.BuildStatus.BUILT && this.props.SuccessReload) {
			location.reload();
		}

		var txt, icon, cls;
		switch (this.state.status) {
		case this.BuildStatus.ERROR:
			return (
				<a key="indicator" className={`build-indicator btn ${this.props.btnSize} btn-danger`}>
					<i className="fa fa-exclamation-triangle"></i>
				</a>
			);

		case this.BuildStatus.UNKNOWN:
		case this.BuildStatus.NA:
			return (
				<a key="indicator"
					data-tooltip={this.props.tooltipPosition}
					title={this.props.Buildable ? "Build this version" : null}
					onClick={this.props.Buildable ? this.triggerBuild : null}
					className={`build-indicator btn ${this.props.btnSize} not-available`}>
					<i className="fa fa-circle"></i>
				</a>
			);

		case this.BuildStatus.FAILURE:
			txt = "failed";
			cls = "danger";
			icon = "fa-exclamation-circle";
			break;

		case this.BuildStatus.BUILT:
			txt = "succeeded";
			cls = "success";
			icon = "fa-check";
			break;

		case this.BuildStatus.STARTED:
			txt = "started";
			cls = "info";
			icon = "fa-circle-o-notch fa-spin";
			break;

		case this.BuildStatus.QUEUED:
			txt = "queued";
			cls = "info";
			icon = "fa-circle-o-notch";
			break;
		}
		return (
			<a key="indicator"
				className={`build-indicator btn ${this.props.btnSize} text-${cls}`}
				href={`/${this.props.RepoURI}/.builds/${this.state.LastBuild.CommitID}/${this.state.LastBuild.Attempt}`}
				data-tooltip={this.props.tooltipPosition}
				title={`Build ${txt}`}>
				<i className={`fa ${icon}`}></i>
			</a>
		);
	},
});

module.exports = RepoBuildIndicator;
