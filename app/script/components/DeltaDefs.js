var React = require("react");
var DefDelta = require("./DefDelta");

var DeltaDefs = React.createClass({
	getInitialState() {
		return {};
	},
	renderDefDelta(dd) {
		return <DefDelta key={(dd.Base || dd.Head).Path} defDelta={dd}/>;
	},
	render() {
		var defDeltas = this.props.deltaDefs ? this.props.deltaDefs.Defs : [];
		return (
			<ol className="delta-defs list-group">
			{defDeltas.map(this.renderDefDelta)}
			</ol>
		);
	},
});
module.exports = DeltaDefs;
