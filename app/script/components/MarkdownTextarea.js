var React = require("react");
var ReactDOM = require("react-dom");
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
			bodyMarkdown: $(ReactDOM.findDOMNode(this)).find(".raw-body").val(),
		});
	},

	_pasteHandler(event) {
		// Get the file from clipboard event.
		var items = event.clipboardData.items;
		if (items.length === 0) {
			return;
		}
		var item = items[0];
		if (item.kind !== "file") {
			return;
		}
		if (item.type !== "image/png") {
			return;
		}
		var file = item.getAsFile();

		// Upload the file.
		var req = new XMLHttpRequest();
		req.open("POST", "/ui/.usercontent");
		req.setRequestHeader("Content-Type", "image/png");
		req.responseType = "json";
		req.onload = () => {
			var upload = req.response;
			if (upload.Error !== undefined) {
				console.log(upload.Error);
				return;
			}

			// Insert the file into textarea.
			var url = `/usercontent/${upload.Name}`;
			this.insertText(`![Image](${url})`);
		};
		req.send(file);
	},

	/**
	 * @description Gets or sets the value of the textarea.
	 * @param {string=} str - (Optional) If provided, this value will be set in
	 * the text area.
	 * @returns {string|undefined} If a new value is set, undefined is returned.
	 */
	value(str) {
		var txt = $(ReactDOM.findDOMNode(this)).find(".raw-body");
		if (typeof str === "string") {
			return txt.val(str);
		}
		return txt.val();
	},

	/**
	 * @description Inserts a string into the textarea.
	 * @param {string} inserted - Value to insert into the text area.
	 * @returns {void}
	 */
	insertText(inserted) {
		var txt = $(ReactDOM.findDOMNode(this)).find(".raw-body");
		var value = txt.val();
		var start = txt[0].selectionStart;
		var end = txt[0].selectionEnd;
		txt.val(value.substring(0, start) + inserted + value.substring(end));
		txt[0].selectionStart = start + inserted.length;
		txt[0].selectionEnd = start + inserted.length;
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
						<textarea className="raw-body" placeholder={this.props.placeholder} defaultValue={this.props.defaultValue} onPaste={this._pasteHandler} />
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
