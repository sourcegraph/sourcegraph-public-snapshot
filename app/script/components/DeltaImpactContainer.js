var React = require("react");
var client = require("../client");
var DeltaImpact = require("./DeltaImpact");

var DeltaImpactContainer = React.createClass({
	getInitialState() {
		return {deltaAffectedAuthorsByUnit: {}, deltaAffectedClientsByUnit: {}};
	},
	componentDidMount() {
		client.deltaListUnits(this.props.deltaRouteVars).success(function(resp) {
			if (resp) this.setState({units: resp || []});
			this.state.units.forEach(function(unitDelta) {
				var unit = unitDelta.Base || unitDelta.Head;

				client.listAffectedAuthors(this.props.deltaRouteVars, {UnitType: unit.Type, Unit: unit.Name}).success(function(resp2) {
					var deltaAffectedAuthorsByUnit = this.state.deltaAffectedAuthorsByUnit;
					deltaAffectedAuthorsByUnit[`${unit.Type}:${unit.Name}`] = resp2 || [];
					if (resp2) this.setState({deltaAffectedAuthorsByUnit: deltaAffectedAuthorsByUnit});
				}.bind(this)).error(function(err) {
					console.error(err);
					this.setState({error: true});
				}.bind(this));

				client.listAffectedClients(this.props.deltaRouteVars, {UnitType: unit.Type, Unit: unit.Name}).success(function(resp2) {
					var deltaAffectedClientsByUnit = this.state.deltaAffectedClientsByUnit;
					deltaAffectedClientsByUnit[`${unit.Type}:${unit.Name}`] = resp2 || [];
					if (resp2) this.setState({deltaAffectedClientsByUnit: deltaAffectedClientsByUnit});
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
		var authors = this.state.deltaAffectedAuthorsByUnit[`${unit.Type}:${unit.Name}`];
		var clients = this.state.deltaAffectedClientsByUnit[`${unit.Type}:${unit.Name}`];
		return (
			<tr className="delta-item-counts" key={unit.Name}>
			<td className="scope">
			{unit.Name}
			</td>
			<td className="diffstat-cell">
			{authors ? authors.length : "?"} authors, {clients ? clients.length : "?"} users
			</td>
			</tr>
		);
	},
	renderUnitImpact(unitDelta) {
		var unit = unitDelta.Base || unitDelta.Head;
		return (
			<div className="delta-impact-by-unit panel panel-main" key={unit.Name}>
			<div className="panel-heading">
			{unit.Name}
			</div>
			<DeltaImpact deltaRouteVars={this.props.deltaRouteVars}
				deltaSpec={this.props.deltaSpec}
				deltaAffectedAuthors={this.state.deltaAffectedAuthorsByUnit[`${unit.Type}:${unit.Name}`]}
				deltaAffectedClients={this.state.deltaAffectedClientsByUnit[`${unit.Type}:${unit.Name}`]}
				deltaFilter={{UnitType: unit.Type, Unit: unit.Name}}/>
			</div>
		);
	},
	render() {
		return (
			<div>
			<p className="diff-summary"><i className="icon octicon octicon-diff"></i> Showing affected users and authors</p>
			<table className="delta-counts">
			{(this.state.units || []).map(this.renderUnitCount)}
			</table>
			{(this.state.units || []).map(this.renderUnitImpact)}
			</div>
		);
	},
});

module.exports = DeltaImpactContainer;
