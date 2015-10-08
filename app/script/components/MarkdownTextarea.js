var React = require("react");
var $ = require("jquery");
var classNames = require("classnames");
var MarkdownView = require("./MarkdownView");

var TAB_EDIT = "edit",
	TAB_PREVIEW = "preview";

var MarkdownTextarea = React.createClass({

	propTypes: {
		placeholder: React.PropTypes.string,
		defaultValue: React.PropTypes.string,
	},

	getDefaultProps() {
		return {
			placeholder: "",
			defaultValue: "",
		};
	},

	getInitialState() {
		return {
			activeTab: "edit",
			markdownBody: "",
		};
	},

	_showEdit() {
		this.setState({activeTab: TAB_EDIT});
	},

	_showPreview() {
		this.setState({
			activeTab: TAB_PREVIEW,
			bodyMarkdown: $(React.findDOMNode(this)).find(".raw-body").val(),
		});
	},

	/**
	 * @description Gets or sets the value of the textarea.
	 * @param {string=} str - (Optional) If provided, this value will be set in
	 * the text area.
	 * @returns {string|undefined} If a new value is set, undefined is returned.
	 */
	value(str) {
		var txt = $(React.findDOMNode(this)).find(".raw-body");
		if (typeof str === "string") {
			return txt.val(str);
		}
		return txt.val();
	},

	render() {
		var cx = classNames({
			"tab-content": true,
			"show-edit": this.state.activeTab === TAB_EDIT,
			"show-preview": this.state.activeTab === TAB_PREVIEW,
		}) + (this.props.className ? ` ${this.props.className}` : "");

		return (
			<div className="markdown-textarea">
				<ul className="nav nav-tabs">
					<li className={this.state.activeTab === TAB_EDIT ? "active" : ""}>
						<a onClick={this._showEdit}><span className="octicon octicon-pencil"></span> Edit</a>
					</li>
					<li className={this.state.activeTab === TAB_PREVIEW ? "active" : ""}>
						<a onClick={this._showPreview}><span className="octicon octicon-search"></span> Preview</a>
					</li>
				</ul>

				<div className={cx}>
					<div className="tab-edit">
						<textarea className="raw-body" placeholder={this.props.placeholder} defaultValue={this.props.defaultValue} />
					</div>
					<div className="tab-preview">
						<MarkdownView content={this.state.bodyMarkdown} />
					</div>
				</div>
			</div>
		);
	},
});

module.exports = MarkdownTextarea;
