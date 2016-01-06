var React			= require("react"),
	Clipboard	= require("clipboard");

var RepoCloneBox = React.createClass({

	propTypes: {
		SSHCloneURL: React.PropTypes.string,
		HTTPCloneURL: React.PropTypes.string,
	},

	getInitialState() {
		return {
			type: "HTTP",
		};
	},

	componentDidMount() {
		this.cipboard = new Clipboard(".clone-url-wrap .clone-copy");
	},

	componentWillUnmount() {
		if (this.clipboard) {
			this.clipboard.destroy();
		}
	},

	_toggleType(type, sshAvailable) {
		if (sshAvailable) {
			this.setState({
				type: type,
			});
		}
	},

	render() {
		var url 		 = this.props.HTTPCloneURL,
			nextType	 = this.state.type === "SSH" ? "HTTP" : "SSH";

		var sshAvailable = this.props.SSHCloneURL.length !== 0;

		if (this.state.type === "SSH") {
			url = this.props.SSHCloneURL;
		}

		return (
			<div className="clone-url-wrap input-group input-group-sm pull-right">
				<div className="input-group-btn">
					<button className="btn btn-primary clone-url-toggle clone-type"
						onClick={this._toggleType.bind(this, nextType, sshAvailable)}
						title={sshAvailable ? null : "SSH unavailable"}
						data-tooltip={sshAvailable ? null : "true"}
						data-placement={sshAvailable ? null : "bottom"}>
							{this.state.type + (url.indexOf("https://") > -1 ? "S" : "")}
					</button>
				</div>

				<span id="clone-url-value" className="form-control">{url}</span>

				<div className="input-group-btn">
					<button className="btn btn-neutral clone-copy" data-clipboard-target="#clone-url-value">
						<span className="octicon octicon-clippy"></span>
					</button>
				</div>
			</div>
		);
	},
});

module.exports = RepoCloneBox;
