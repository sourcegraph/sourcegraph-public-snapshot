var React = require("react");
var TimeAgo = require("./TimeAgo");
var client = require("../client");

var RepoBuildStatus = React.createClass({

	getInitialState() {
		return {
			LastBuild: this.props.LastBuild,
		};
	},

	componentWillMount() {
		this.intervals = [];
	},

	componentDidMount() {
		// TODO(x): fix this stupid polling
		this.setInterval(this.checkBuildStatus, 15000);
		if (typeof this.state.LastBuild === "undefined") this.checkBuildStatus();
	},

	componentWillUnmount() {
		this.intervals.map(clearInterval);
	},

	setInterval() {
		this.intervals.push(Reflect.apply(setInterval, null, arguments));
	},

	checkBuildStatus() {
		client.builds(this.props.Repo.URI, this.props.Rev).success(function(resp) {
			if (resp) {
				this.setState({LastBuild: resp[0]});
			} else {
				this.setState({LastBuild: null}); // LastBuild == null means doesn't exist, undefined means don't know yet
			}
			this.setState({error: false});
		}.bind(this)).error(function(err) {
			console.error(err);
			this.setState({error: true});
		}.bind(this));
	},

	render() {
		if (this.state.error) {
			return <span className="text-danger">Error getting status</span>;
		}
		if (this.state.LastBuild) {
			var txt;
			var at;
			var cls;
			if (this.state.LastBuild.Failure) {
				txt = "build failed ";
				at = this.state.LastBuild.EndedAt;
				cls = "text-danger";
			} else if (this.state.LastBuild.Success) {
				txt = "built ";
				at = this.state.LastBuild.EndedAt;
				cls = "text-success";
			} else if (this.state.LastBuild.StartedAt && !this.state.LastBuild.EndedAt) {
				txt = "started ";
				at = this.state.LastBuild.StartedAt;
				cls = "text-info";
			} else {
				txt = "queued ";
				at = this.state.LastBuild.CreatedAt;
				cls = "text-warning";
			}
			return (
				<a href={"/" + this.props.Repo.URI + "/.builds/" + this.state.LastBuild.BID}>
					<span className={cls}><span className="commit-id">{this.state.LastBuild.CommitID.slice(0, 6)}</span> {txt} <TimeAgo time={at} /></span>
				</a>
			);
		}
		if (typeof this.state.LastBuild === "undefined") {
			return <span className="text-muted">Loading...</span>;
		}
		return <span className="text-warning">Never built</span>;
	},
});

module.exports = RepoBuildStatus;
