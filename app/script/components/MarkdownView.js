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
		$.ajax({
			method: "POST",
			url: "/.markdown",
			headers: {
				"X-CSRF-Token": globals.CsrfToken,
			},
			data: this.props.content,
		}).done(function(data) {
			this.setState({html: data});
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
