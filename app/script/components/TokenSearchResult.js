var React = require("react");
var router = require("../routing/router");

var TokenSearchResult = React.createClass({
	render() {
		var doc;
		if (this.props.result.Def.DocHTML) {
			// This HTML should be sanitized in ui/search.go
			doc = <p dangerouslySetInnerHTML={this.props.result.Def.DocHTML}></p>;
		}

		var def = this.props.result.Def;
		var href = router.defURL(def.Repo, def.CommitID, def.UnitType, def.Unit, def.Path);

		return (
			<div className="token-search-result">
				<hr/>
				<a href={href}>
					<code>{this.props.result.Def.Kind} </code>
					{/* This HTML should be sanitized in ui/search.go */}
					<code dangerouslySetInnerHTML={this.props.result.QualifiedName}></code>
				</a>
				{doc}
			</div>
		);
	},
});

module.exports = TokenSearchResult;
