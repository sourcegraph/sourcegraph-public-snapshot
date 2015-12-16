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
			autoFocus: false,
		};
	},

	getInitialState() {
		return {
			activeTab: "edit",
			markdownBody: "",
			isDragging: false,
		};
	},

	componentWillMount() {
		this._resetDrag();
	},

	componentDidMount() {
		this._bindListeners();
	},

	componentWillUnmount() {
		this._unbindListeners();
	},

	_unbindListeners() {
		ReactDOM.findDOMNode(this).removeEventListener("dragenter", this._handleDrag);
		ReactDOM.findDOMNode(this).removeEventListener("dragleave", this._handleDrag);
		ReactDOM.findDOMNode(this).removeEventListener("drop", this._handleDrop);
	},

	_bindListeners() {
		ReactDOM.findDOMNode(this).addEventListener("dragenter", this._handleDrag);
		ReactDOM.findDOMNode(this).addEventListener("dragleave", this._handleDrag);
		ReactDOM.findDOMNode(this).addEventListener("drop", this._handleDrop);
	},

	_handleDrag(e) {
		this._numDrags += e.type === "dragenter" ? 1 : -1;
		var isDragging = this.state.isDragging;

		if (this._numDrags === 1) {
			isDragging = true;
		} else if (this._numDrags === 0) {
			isDragging = false;
		}

		this.setState({isDragging: isDragging});
	},

	_resetDrag() {
		this._numDrags = 0;
		this.setState({isDragging: false});
	},

	_handleDrop(e) {
		e.preventDefault();
		this._resetDrag();

		var items = event.dataTransfer ? event.dataTransfer.items : [];

		for (var i = 0; i < items.length; i++) {
			// Since `items` is read only we have to pass an addition arg
			this.uploadFile(items[i], "drop");
		}
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
		this.uploadFile(item, "clipboard");
	},

	uploadFile(item, source) {
		if (item.kind !== "file") {
			console.log(`rejecting file upload because "${item.kind}" is not a currently supported item.kind`);
			return;
		}

		if (!item.type.match("image.*")) {
			console.log(`rejecting file upload because "${item.type}" is not a currently supported item.type`);
			return;
		}

		$.ajax({
			url: "/.ui/.usercontent",
			method: "POST",
			contentType: item.type,
			headers: {
				"X-Csrf-Token": window._csrfToken,
			},
			accepts: "json",
			data: item.getAsFile(),
			processData: false,
			success: (upload) => {
				if (upload.Error !== undefined) {
					console.log(upload.Error);
					return;
				}

				// Insert the file into textarea.
				var url = `/usercontent/${upload.Name}`;
				this.insertText(`![Image](${url})`, source === "drop");
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
	 * @param {bool}   append   - Whether to append, or insert at cursor pos
	 * @returns {void}
	 */
	insertText(inserted, append) {
		var txt = $(ReactDOM.findDOMNode(this)).find(".raw-body");
		var value = txt.val();

		if (append === true) {
			txt.val(value + inserted);
			return;
		}

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
			"has-drag-item": this.state.isDragging,
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
						<textarea className="raw-body" placeholder={this.props.placeholder} defaultValue={this.props.defaultValue} onPaste={this._pasteHandler} autoFocus={this.props.autoFocus} />
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
