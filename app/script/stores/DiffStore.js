var Backbone = require("backbone");

var AppDispatcher = require("../dispatchers/AppDispatcher");
var FluxStore = require("./FluxStore");
var FileDiffModel = require("./models/FileDiffModel");
var TokenPopoverModel = require("./models/TokenPopoverModel");
var DiffPopupModel = require("./models/DiffPopupModel");

/**
 * @description The DiffStore holds a collection of FileDiffs and manages the state
 * of the CompareView.
 */
var DiffStore = FluxStore({

	defaults: {
		/**
		 * @description Holds the popover state for the compare view.
		 * @type {TokenPopoverModel}
		 */
		popoverModel: new TokenPopoverModel(),

		/**
		 * @description Holds the popup model state.
		 * @type {TokenPopupModel}
		 */
		popupModel: new DiffPopupModel(),

		/**
		 * @description Holds the open/closed state of the file diff list.
		 * @type {bool}
		 */
		fileListClosed: false,

		proposingChange: false,
		changesetLoading: false,
	},

	/**
	 * @description Holds a collection of file diffs that are shown in the compare view.
	 * @type {Backbone.Collection|null}
	 */
	fileDiffs: null,

	dispatcher: AppDispatcher,

	actions: {
		DIFF_LOAD_DATA: "_onLoadData",
		DIFF_FOCUS_TOKEN: "_onFocusToken",
		DIFF_BLUR_TOKENS: "_onBlurTokens",
		DIFF_RECEIVED_POPOVER: "_onReceivedPopover",
		DIFF_SELECT_TOKEN: "_onSelectToken",
		DIFF_DESELECT_TOKENS: "_onClosedPopup",
		DIFF_RECEIVED_POPUP: "_onReceivedPopup",
		DIFF_FETCH_EXAMPLE: "_onFetchExample",
		DIFF_RECEIVED_EXAMPLE: "_onReceivedExample",
		DIFF_SELECT_FILE: "_onSelectFile",
		DIFF_RECEIVED_HUNK_TOP: "_onReceiveHunkTop",
		DIFF_RECEIVED_HUNK_BOTTOM: "_onReceiveHunkBottom",
		DIFF_PROPOSE_CHANGE: "_onProposeChange",
	},

	initialize() {
		this.fileDiffs = new Backbone.Collection();
	},

	_onProposeChange(data) {
		this.set("changesetLoading", true);
	},

	/**
	 * @description Triggered when the user has requested context at the top of
	 * a hunk.
	 * @param {Object} action - payload action data
	 * @returns {void}
	 * @private
	 */
	_onReceiveHunkTop(action) {
		var model = action.model,
			data = action.data;

		if (!model || typeof model.updateTop !== "function") {
			return console.error("action received invalid model");
		}

		model.updateTop(data);

		// if the hunks are "touching", merge them
		var index = model.index(),
			parent = model.get("Parent"),
			prevHunk = parent.get("Hunks").at(index - 1);

		if (prevHunk && prevHunk.get("NewStartLine") + prevHunk.get("NewLines") === data.Entry.StartLine) {
			parent.merge(index - 1, index);
		}
	},

	/**
	 * @description Triggered when the user has requested context at the bottom of
	 * a hunk.
	 * @param {Object} action - payload action data
	 * @returns {void}
	 * @private
	 */
	_onReceiveHunkBottom(action) {
		var model = action.model,
			data = action.data;

		if (!model || typeof model.updateBottom !== "function") {
			return console.error("action received invalid model");
		}

		model.updateBottom(data);

		// if the hunks are "touching", merge them
		var index = model.index(),
			parent = model.get("Parent"),
			nextHunk = parent.get("Hunks").at(index + 1);

		if (nextHunk && nextHunk.get("NewStartLine") === data.Entry.EndLine + 1) {
			parent.merge(index, index + 1);
		}
	},

	/**
	 * @description Triggered when a file is clicked in the file diff list. Depending
	 * on whether the diff is over threshold or not, the view might scroll to the file
	 * or it might open a link to view that file independently.
	 * @param {Object} action - payload action data
	 * @returns {void}
	 * @private
	 */
	_onSelectFile(action) {
		this.trigger("scrollTop", action.file.getAbsolutePosition().top);
	},

	/**
	 * @description Triggered when a new example is requested from the server.
	 * @param {Object} action - payload action data
	 * @returns {void}
	 * @private
	 */
	_onFetchExample(action) {
		this.get("popupModel").setLoading(true);
	},

	/**
	 * @description Triggered when a new example is received from the server.
	 * @param {Object} action - payload action data
	 * @returns {void}
	 * @private
	 */
	_onReceivedExample(action) {
		var pm = this.get("popupModel");
		pm.setLoading(false);
		pm.showExample(action.data, action.page);
	},

	/**
	 * @description Triggered when data to populate a popup is received from the server.
	 * @param {Object} action - payload action data
	 * @returns {void}
	 * @private
	 */
	_onReceivedPopup(action) {
		this.get("popupModel").show(action.data);
	},

	/**
	 * @description Triggered when the action for a token selection is dispatched.
	 * @param {Object} action - payload action data
	 * @returns {void}
	 * @private
	 */
	_onSelectToken(action) {
		this.get("popupModel").destroyExample();
		this.get("popupModel").set({closed: true});
		this.get("popoverModel").set({visible: false});
		this.fileDiffs.forEach(fd => fd.selectToken(action.token.get("url")[0]));
	},

	/**
	 * @description Action called when the user closes a popup.
	 * @returns {void}
	 * @private
	 */
	_onClosedPopup() {
		this.get("popupModel").destroyExample();
		this.fileDiffs.forEach(fd => fd.clearSelected());
	},

	/**
	 * @description Action called when popover data is received from the server.
	 * @param {Object} action - payload action data
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
	 * @description Action called when any token in the compare view is focused.
	 * @param {Object} action - payload action data
	 * @returns {void}
	 * @private
	 */
	_onFocusToken(action) {
		this.get("popupModel").highlightTokens(action.token.get("url")[0]);

		if (typeof action.file !== "undefined") {
			action.file.highlightToken(action.token.get("url")[0]);
		} else {
			this.fileDiffs.forEach(fd => fd.highlightToken(action.token.get("url")[0]));
		}

		this.get("popoverModel").positionAt(action.event);
	},

	/**
	 * @description Action called when any token in the compare view loses focus.
	 * @param {Object} action - payload action data
	 * @returns {void}
	 * @private
	 */
	_onBlurTokens(action) {
		this.get("popupModel").clearHighlightedTokens();

		if (typeof action.file !== "undefined") {
			action.file.blurTokens();
		} else {
			this.fileDiffs.forEach(fd => fd.blurTokens());
		}

		this.get("popoverModel").set({visible: false});
	},

	/**
	 * @description Action called when data is received to popuplate the compare view.
	 * @param {Object} action - payload action data
	 * @returns {void}
	 * @private
	 */
	_onLoadData(action) {
		this.fileDiffs.add(
			(action.data.DiffData.FileDiffs || []).map(
				fileDiff => new FileDiffModel(fileDiff, {parse: true})
			)
		);

		this.set(action.data, {silent: true}); // silent doesn't trigger view update

		if (this.get("OverThreshold")) {
			this.set("filter", {}, {silent: true});
		}

		this.set("fileDiffs", this.fileDiffs);
	},
});

module.exports = DiffStore;
