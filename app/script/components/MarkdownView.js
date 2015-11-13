var React = require("react");
var $ = require("jquery");
var globals = require("../globals");

var MarkdownView = React.createClass({
	propTypes: {
		content: React.PropTypes.string.isRequired,
	},

	getDefaultProps() {
		return {
			content: "",
		};
	},

	getInitialState() {
		return {html: null};
	},

	componentDidMount() {
		// Render the markdown content.
		this._renderMarkdown(this.props);
	},

	componentWillReceiveProps(nextProps) {
		// Component will get new properties, so re-render the Markdown content.
		this._renderMarkdown(nextProps);
	},

	// _renderMarkdown performs an AJAX request to render the Markdown content to
	// sanitized HTML on the server.
	_renderMarkdown(props) {
		$.ajax({
			method: "POST",
			url: "/.markdown",
			headers: {
				"X-CSRF-Token": globals.CsrfToken,
			},
			data: props.content,
		}).done(function(data) {
			if (this.isMounted()) {
				this.setState({html: data});
			}
		}.bind(this));
	},

	render() {
		if (this.state.html) {
			return <div className="markdown-view" dangerouslySetInnerHTML={this.state.html} />;
		}
		return <div></div>;
	},
});

module.exports = MarkdownView;
