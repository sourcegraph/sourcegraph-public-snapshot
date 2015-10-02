var React = require("react");

var TokenSearchResult = React.createClass({
	render() {
		var doc;
		if (this.props.result.Def.DocHTML) {
			// This HTML should be sanitized in ui/search.go
			doc = <p dangerouslySetInnerHTML={this.props.result.Def.DocHTML}></p>;
		}

		return (
			<div className="token-search-result">
				<hr/>
				<a href={this.props.result.URL}>
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
