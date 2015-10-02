var React = require("react");

var TokenSearchResult = React.createClass({
	render() {
		return (
			<div className="token-search-result">
				<hr/>
				<a href={this.props.result.URL}>
					<code dangerouslySetInnerHTML={this.props.result.QualifiedName}></code>
				</a>
			</div>
		);
	},
});

module.exports = TokenSearchResult;
