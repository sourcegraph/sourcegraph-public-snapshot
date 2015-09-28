var React = require("react");
var DeltaAuthor = require("./DeltaAuthor");
var DeltaClient = require("./DeltaClient");

var DeltaImpact = React.createClass({
	getInitialState() {
		return {};
	},
	renderDeltaAuthor(da) {
		return <DeltaAuthor deltaAuthor={da}/>;
	},
	renderDeltaClient(dc) {
		return <DeltaClient deltaClient={dc}/>;
	},
	render() {
		var deltaAuthors = this.props.deltaAffectedAuthors || [];
		var deltaClients = this.props.deltaAffectedClients || [];
		return (
			<section className="delta-impact">
			<ul className="delta-authors list-group media-list">
			{deltaAuthors.map(this.renderDeltaAuthor)}
			</ul>
			<ul className="delta-clients list-group media-list">
			{deltaClients.map(this.renderDeltaClient)}
			</ul>
			</section>
		);
	},
});
module.exports = DeltaImpact;
