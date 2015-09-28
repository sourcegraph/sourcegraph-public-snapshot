var FluxStore = require("./FluxStore");

var CodeModel = require("./models/CodeModel");
var ContextMenuModel = require("./models/ContextMenuModel");
var AppDispatcher = require("../dispatchers/AppDispatcher");
var TokenPopoverModel = require("./models/TokenPopoverModel");

module.exports = FluxStore({

	defaults: {
		/**
		 * @description loading will be through then the code view is loading a file.
		 * @type {bool}
		 */
		loading: true,

		/**
		 * @description The CodeModel that is contained by this store.
		 * @type {CodeModel}
		 */
		codeModel: new CodeModel(),

		/**
		 * @description The model for the contents of the context menu. The
		 * context menu is shown upon clicking on multiple definitions.
		 * @type {ContextMenuModel}
		 */
		contextMenuModel: new ContextMenuModel(),

		/**
		 * @description The model for the popover that is shown when hovering
		 * tokens in the code view".
		 * @type {TokenPopoverModel}
		 */
		popoverModel: new TokenPopoverModel(),
	},

	dispatcher: AppDispatcher,

	actions: {
		FETCH_FILE: "_onFetchFile",
		RECEIVED_FILE: "_onReceive",
		TOKEN_SELECT: "_onSelectToken",
		TOKEN_CLEAR: "_onTokenClear",
		TOKEN_FOCUS: "_onFocusToken",
		TOKEN_BLUR: "_onTokenBlur",
		SHOW_DEFINITION: "_onShowSnippet",
		SHOW_SNIPPET: "_onShowSnippet",
		LINE_SELECT: "_onLineSelect",
		SWITCH_POPUP_DEFINITION: "_onSwitchPopupDefinition",
		REDIRECT: "_onRedirect",
		LOAD_CONTEXT_MENU: "_onLoadContextMenu",
		RECEIVED_MENU_OPTIONS: "_onReceivedMenuOptions",
		RECEIVED_POPOVER: "_onReceivedPopover",
		CODE_FILE_CLICK: "_onCodeFileClick",
	},

	/*
	 * @description Triggered when a request to change the popup view to a new
	 * URL happens.
	 * @private
	 */
	_onSwitchPopupDefinition() {
		this.get("contextMenuModel").set("closed", true);
	},

	/*
	 * @description Triggered when the code file is clicked to close the context
	 * menu.
	 * @param {Object} action - payload action data
	 * @private
	 */
	_onCodeFileClick(action) {
		var cmm = this.get("contextMenuModel");
		if (!cmm.get("closed")) {
			this.get("contextMenuModel").set("closed", true);
		}
	},

	/*
	 * @description Appends new received data to the context menu items.
	 * @param {Object} action - payload action data
	 * @private
	 */
	_onReceivedMenuOptions(action) {
		var options = action.data.map(def => {
			def.File = this.get("file");
			return {
				label: def.QualifiedName,
				data: def,
			};
		});

		this.get("contextMenuModel").set({
			options: options,
			closed: false,
		});
	},

	/*
	 * @description Triggered when the context menu is preparing the receive
	 * options. Positions the menu under the clicked token.
	 * @param {Object} action - payload action data
	 * @private
	 */
	_onLoadContextMenu(action) {
		var menuModel = this.get("contextMenuModel");
		var pos = action.token.getRelativePosition();
		menuModel.set({
			position: {
				top: pos.top+17,
				left: pos.left+5,
			},
		});
	},

	/**
	 * @description Sets the store into loading mode.
	 * @returns {void}
	 * @private
	 */
	_onFetchFile() {
		this.set("loading", true);
	},

	/**
	 * @description Selects the token described in the payload.
	 * @param {Object} action - Action payload.
	 * @returns {void}
	 * @private
	 */
	_onSelectToken(action) {
		this.get("popoverModel").set({visible: false});
		var cm = this.get("codeModel");
		cm.clearHighlightedLines();
		cm.selectToken(action.token.get("url")[0]);
	},

	/**
	 * @description Clears all token selections and line highlights.
	 * @param {Object} action - Action payload.
	 * @returns {void}
	 * @private
	 */
	_onTokenClear(action) {
		var cm = this.get("codeModel");
		cm.clearSelectedTokens();
		cm.clearHighlightedLines();
	},

	/**
	 * @description Updates the popover to the data received and contained in the action payload.
	 * @param {Object} action - Action payload.
	 * @returns {void}
	 * @private
	 */
	_onReceivedPopover(action) {
		this.get("popoverModel").set({
			visible: true,
			body: action.data,
		});
	},

	/**
	 * @description Highlights the token contained in the payload.
	 * @param {Object} action - Action payload.
	 * @returns {void}
	 * @private
	 */
	_onFocusToken(action) {
		this.get("codeModel").highlightToken(action.token.get("url")[0]);
		this.get("popoverModel").positionAt(action.event);
	},

	/**
	 * @description Removes all token highlights.
	 * @returns {void}
	 * @private
	 */
	_onTokenBlur() {
		this.get("codeModel").clearHighlightedTokens();
		this.get("popoverModel").set({visible: false});
	},

	/**
	 * @description Loads new data received from the server into the code model and
	 * sets the store properties.
	 * @param {Object} action - Action payload.
	 * @returns {void}
	 * @private
	 */
	_onReceive(action) {
		var data = action.data,
			def = data.Definition ? data.Definition : null;

		this.get("codeModel").load(data.Entry);
		this.set({
			file: data.EntrySpec,
			activeDef: def,
			loading: false,
			numRefs: data.Entry.SourceCode.NumRefs,
			maxRefs: data.Entry.SourceCode.TooManyRefs,
			buildInfo: data.RepoBuildInfo,
			latestCommit: data.RepoCommit,
		});
	},

	/**
	 * @description Redirects the browser to the specified URL. This generally happens
	 * if there has been a request to go to a definition that is a package, so the user
	 * will be redirected to the tree view.
	 * @param {Object} action - Payload, contains redirect URL.
	 * @returns {void}
	 * @private
	 */
	_onRedirect(action) {
		document.location = action.url;
	},

	/**
	 * @description Highlights the passed snippet of code and additionally selects the passed
	 * token.
	 * @param {Object} action - Action payload.
	 * @returns {void}
	 * @private
	 */
	_onShowSnippet(action) {
		var params = action.params;

		if (params.ByteStartPosition && params.ByteEndPosition) {
			this._showByteRange(params.ByteStartPosition, params.ByteEndPosition);
		} else if (params.startLine && params.endLine) {
			this._showLineRange(params.startLine, params.endLine);
		}

		if (params.defUrl) {
			this.get("codeModel").selectToken(params.defUrl);
		}
	},

	/**
	 * @description Triggers line selection start or end, based on the status of the shiftKey
	 * sent via the action payload.
	 * @param {Object} action - Action payload.
	 * @returns {void}
	 * @private
	 */
	_onLineSelect(action) {
		var startLine = action.lineNumber;
		var endLine = action.lineNumber;

		if (action.shiftKey) {
			// Preserve and expand previous line selection to include the clicked line.
			var highlighted = this.getHighlightedLines();
			if (Array.isArray(highlighted)) {
				if (highlighted[0].get("number") < startLine) {
					startLine = highlighted[0].get("number");
				}
				if (highlighted[highlighted.length - 1].get("number") > endLine) {
					endLine = highlighted[highlighted.length - 1].get("number");
				}
			}
		}

		this.get("codeModel").highlightLineRange(startLine, endLine);
		this.set("snippet", {
			start: startLine,
			end: endLine,
		});
	},

	/**
	 * @description Highlights the passed line range and brings it into view.
	 * @param {number} start - Start line.
	 * @param {number} end - End line.
	 * @returns {void}
	 * @private
	 */
	_showLineRange(start, end) {
		var lines = this.get("codeModel").highlightLineRange(start, end);
		if (lines.length) this._scrollIntoView(lines[0]);

		this.set("snippet", {
			start: start,
			end: end,
		});
	},

	/**
	 * @description Highlights the passed byte range and scrolls it into view.
	 * @param {number} start - Start byte.
	 * @param {number} end - End byte.
	 * @returns {void}
	 * @private
	 */
	_showByteRange(start, end) {
		var lines = this.get("codeModel").highlightByteRange(start, end);
		if (lines.length) this._scrollIntoView(lines[0]);
	},

	/**
	 * @description Brings the passed token or line into view. The CodeFileView
	 * will respond to the triggered event.
	 * @param {CodeTokenModel|CodeLineModel} lineOrToken - Line or token to scroll to.
	 * @returns {void}
	 * @private
	 */
	_scrollIntoView(lineOrToken) {
		this.trigger("scrollTop", lineOrToken);
	},

	/**
	 * @description Returns true if the passed file is the same as this one.
	 * @param {Object} file - Must contain "RepoRev" object and "Path" as keys.
	 * @returns {bool} - True or false.
	 */
	isSameFile(file) {
		var f = this.get("file"),
			isSameRepo = file.RepoRev.URI === f.RepoRev.URI,
			isSameRev = file.RepoRev.CommitID === f.RepoRev.CommitID || file.RepoRev.Rev === f.RepoRev.Rev ||
				file.RepoRev.Rev === f.RepoRev.CommitID,
			isSamePath = file.Path === f.Path,
			isSameFile = isSameRepo && isSameRev && isSamePath;

		return isSameFile;
	},

	/**
	 * @description Retrieves an array of highlighted lines, if any. Otherwise returns null
	 * @returns {Array<CodeLineCollection>|null} - Array of lines, if any.
	 */
	getHighlightedLines() {
		return this.get("codeModel").getHighlightedLines();
	},

	destroy() {
		this.get("codeModel").destroy();
		AppDispatcher.unregister(this.dispatchToken);
	},
});
