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

	_toggleType(type) {
		this.setState({
			type: type,
		});
	},

	render() {
		var url 		 = this.props.HTTPCloneURL,
			nextType	 = this.state.type === "SSH" ? "HTTP" : "SSH";

		if (this.state.type === "SSH") {
			url = this.props.SSHCloneURL;
		}

		return (
			<div className="clone-url-wrap input-group input-group-sm">
				<button className="btn btn-primary clone-copy" data-clipboard-target="#clone-url-value">
					<span className="octicon octicon-clippy"></span>
				</button>

				<button className="btn btn-default clone-type"
					onClick={this._toggleType.bind(this, nextType)}
					disabled={this.props.SSHCloneURL.length ? "false" : "true"}>
						{this.state.type + (url.indexOf("https://") > -1 ? "S" : "")}
				</button>

				<span id="clone-url-value" className="form-control">{url}</span>
			</div>
		);
	},
});

module.exports = RepoCloneBox;
