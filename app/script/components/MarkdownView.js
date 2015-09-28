var React = require("react");
var marked = require("marked");

var MarkdownView = React.createClass({
	propTypes: {
		content: React.PropTypes.string.isRequired,
	},

	getDefaultProps() {
		return {
			content: "",
		};
	},

	render() {
		return <div className="markdown-view" dangerouslySetInnerHTML={{__html: marked(this.props.content, {sanitize: true})}} />;
	},
});

module.exports = MarkdownView;
