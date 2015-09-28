var React = require("react");
var router = require("../routing/router");
var client = require("../client");
var classNames = require("classnames");

var DefDelta = React.createClass({
	getInitialState() {
		return {};
	},
	componentDidMount() {
		client.listExamples(this.props.defDelta.Base || this.props.defDelta.Head).success(function(resp) {
			if (resp) this.setState({examples: resp});
		}.bind(this)).error(function(err) {
			console.error(err);
			this.setState({error: true});
		}.bind(this));
	},
	render() {
		var dd = this.props.defDelta;
		var def = dd.Base || dd.Head;
		var df = def.FmtStrings;

		var classes = classNames({
			"delta-list-item": true,
			"list-group-item": true,
			"added": !dd.Base && dd.Head,
			"changed": dd.Base && dd.Head,
			"deleted": dd.Base && !dd.Head,
		});
		var iconClasses = classNames({
			"octicon": true,
			"octicon-diff-added": !dd.Base && dd.Head,
			"octicon-diff-modified": dd.Base && dd.Head,
			"octicon-diff-removed": dd.Base && !dd.Head,
		});

		var refs = this.state.examples ? <div className="pull-right"><a href={router.defExamplesURL(def.Repo, def.CommitID, def.UnitType, def.Unit, def.Path)}>{this.state.examples.length} refs</a></div> : "";

		return (
			<li className={classes} key={def.Path}>
				{refs}
				<div className="pull-left"><i className={iconClasses}></i></div>
				<div className="media-body">
				<code>{df.DefKeyword} <a className="defn-popover" href={router.defURL(def.Repo, def.CommitID, def.UnitType, def.Unit, def.Path)}><span className="name">{df.Name.DepQualified}</span>{df.NameAndTypeSeparator}{df.Type.DepQualified}</a></code>
			</div>
			</li>
		);
	},
});
module.exports = DefDelta;
