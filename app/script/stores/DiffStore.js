var Backbone = require("backbone");

var AppDispatcher = require("../dispatchers/AppDispatcher");
var FluxStore = require("./FluxStore");
var FileDiffModel = require("./models/FileDiffModel");

/**
 * @description The DiffStore holds a collection of FileDiffs and manages the state
 * of the CompareView.
 */
var DiffStore = FluxStore({

	defaults: {
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
	 * @description Triggered when a file is clicked in the file diff list.
	 * @param {Object} action - payload action data
	 * @returns {void}
	 * @private
	 */
	_onSelectFile(action) {
		this.trigger("scrollTop", action.file.getAbsolutePosition().top);
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

		this.set("fileDiffs", this.fileDiffs);
	},
});

module.exports = DiffStore;
