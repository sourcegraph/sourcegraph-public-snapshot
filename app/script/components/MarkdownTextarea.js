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
		autoFocus: React.PropTypes.bool,
	},

	getDefaultProps() {
		return {
			placeholder: "",
			defaultValue: "",
			autoFocus: "false",
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

	// TODO(dmitri): Create/find a good way to propogate backend CLI flag value (appconf.Flags.DisableUserContent)
	//               to this frontend component and disable _pasteHandler if that flag is true. For now, it will just
	//               create a harmless error in console when you try to paste when user content is disabled.
	_pasteHandler(event) {
		// Get the file from clipboard event.
		var items = event.clipboardData.items;
		if (items.length === 0) {
			return;
		}
		var item = items[0];
		if (item.kind !== "file") {
			console.log(`rejecting file upload because "${item.kind}" is not a currently supported item.kind`);
			return;
		}
		if (item.type !== "image/png") {
			console.log(`rejecting file upload because "${item.type}" is not a currently supported item.type`);
			return;
		}
		var file = item.getAsFile();

		// Upload the file.
		$.ajax({
			url: "/ui/.usercontent",
			method: "POST",
			contentType: "image/png",
			accepts: "json",
			data: file,
			processData: false,
			success: (upload) => {
				if (upload.Error !== undefined) {
					console.log(upload.Error);
					return;
				}

				// Insert the file into textarea.
				var url = `/usercontent/${upload.Name}`;
				this.insertText(`![Image](${url})`);
			},
			error: (upload, error) => {
				console.log(error);
			},
		});
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
						<textarea className="raw-body" placeholder={this.props.placeholder} defaultValue={this.props.defaultValue} onPaste={this._pasteHandler} autoFocus={this.props.autoFocus === "true" ? "true" : null} />
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
