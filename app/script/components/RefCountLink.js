var React = require("react");
var router = require("../routing/router");
var client = require("../client");

var RefCountLink = React.createClass({
	getInitialState() {
		return {};
	},
	componentDidMount() {
		client.listExamples(this.props.defSpec).success(function(resp) {
			if (resp) this.setState({refs: resp});
		}.bind(this)).error(function(err) {
			console.error(err);
			this.setState({error: true});
		}.bind(this));
	},
	render() {
		var def = this.props.defSpec;
		if (!this.state.refs) {
			return null;
		}
		return (
			<a href={router.defExamplesURL(def.Repo, def.CommitID, def.UnitType, def.Unit, def.Path)}>{this.state.refs.length} ref{this.state.refs.length === 1 ? "" : "s"}</a>
		);
	},
});
exports.RefCountLink = RefCountLink;
