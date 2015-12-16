var React = require("react");

var UserSSHKey = React.createClass({
	propTypes: {
		SSHKey: React.PropTypes.shape({
			Id: React.PropTypes.number.isRequired,
			Key: React.PropTypes.string.isRequired,
			Name: React.PropTypes.string.isRequired,
		}),
	},

	onClick(e) {
		this.props.onDelete(this.props.SSHKey);
	},

	render() {
		var k = this.props.SSHKey;
		return (
			<div className="list-group-item">
					<a className="remove octicon octicon-x" onClick={this.onClick} style={{position: "absolute", right: "15px"}}></a>
					<h5>{k.Name}</h5>
					<p style={{wordWrap: "break-word"}}>
						{k.Key}
					</p>
			</div>
		);
	},
});

module.exports = UserSSHKey;
