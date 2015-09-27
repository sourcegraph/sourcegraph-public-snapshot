var React = require("react");
var client = require("../client");
var DeltaDefs = require("./DeltaDefs");

var DeltaDefsContainer = React.createClass({
	getInitialState() {
		return {deltaDefsByUnit: {}};
	},
	componentDidMount() {
		client.deltaListUnits(this.props.deltaRouteVars).success(function(resp) {
			if (resp) this.setState({units: resp || []});
			this.state.units.forEach(function(unitDelta) {
				var unit = unitDelta.Base || unitDelta.Head;
				client.deltaListDefs(this.props.deltaRouteVars, {UnitType: unit.Type, Unit: unit.Name}).success(function(resp2) {
					var deltaDefsByUnit = this.state.deltaDefsByUnit || [];
					deltaDefsByUnit[`${unit.Type}:${unit.Name}`] = resp2.Defs ? resp2 : {Defs: []};
					if (resp2) this.setState({deltaDefsByUnit: deltaDefsByUnit});
				}.bind(this)).error(function(err) {
					console.error(err);
					this.setState({error: true});
				}.bind(this));
			}.bind(this));
		}.bind(this)).error(function(err) {
			console.error(err);
			this.setState({error: true});
		}.bind(this));
	},
	renderUnitCount(unitDelta) {
		var unit = unitDelta.Base || unitDelta.Head;
		var dds = this.state.deltaDefsByUnit[`${unit.Type}:${unit.Name}`];
		return (
			<tr className="delta-item-counts">
				<td className="scope">
				{unit.Name}
				</td>
				<td className="diffstat-cell">
				{dds ? dds.Defs.length : ""}
				</td>
			</tr>
		);
	},
	renderUnitDefs(unitDelta) {
		var unit = unitDelta.Base || unitDelta.Head;
		return (
			<div className="delta-defs-by-unit panel panel-main">
			<div className="panel-heading">
				{unit.Name}
			</div>
			<DeltaDefs deltaRouteVars={this.props.deltaRouteVars}
				deltaSpec={this.props.deltaSpec}
				deltaDefs={this.state.deltaDefsByUnit[`${unit.Type}:${unit.Name}`]}
				deltaFilter={{UnitType: unit.Type, Unit: unit.Name}}/>
			</div>
		);
	},
	render() {
		return (
			<div>
			<p className="diff-summary"><i className="icon octicon octicon-diff"></i> Showing changes to API</p>
			<table className="delta-counts">
			{(this.state.units || []).map(this.renderUnitCount)}
			</table>
			{(this.state.units || []).map(this.renderUnitDefs)}
			</div>
		);
	},
});
module.exports = DeltaDefsContainer;
