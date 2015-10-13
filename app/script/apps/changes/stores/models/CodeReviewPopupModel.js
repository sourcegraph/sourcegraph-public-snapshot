var Backbone = require("backbone");
var ExamplesModel = require("./ExamplesModel");

/**
 * @description Manages the state of the popup displayed on top of the CompareView.
 */
module.exports = Backbone.Model.extend({

	defaults: {
		/**
		 * @description The visibility of the popup.
		 * @type {bool}
		 */
		closed: true,

		/**
		 * @description The URL of the definition shown in the popup.
		 * @type {string}
		 */
		URL: null,

		/**
		 * @description The revision in which this definition is located.
		 * @type {string}
		 */
		Rev: "",

		/**
		 * @description The file that holds the definition shown.
		 * @type {string}
		 */
		File: "",

		/**
		 * @description The model of the example shown in the popup.
		 * @type {ExamplesModel}
		 */
		examplesModel: new ExamplesModel(),
	},

	/**
	 * @description Sets the popup loading state.
	 * @param {bool} val - true or false.
	 * @returns {void}
	 */
	setLoading(val) {
		this.get("examplesModel").set("loading", val);
	},

	/**
	 * @description Shows an example received from the server.
	 * @param {Object} data - Raw example data.
	 * @param {number} page - New page number.
	 * @returns {void}
	 */
	showExample(data, page) {
		this.get("examplesModel").showExample(data, page);
	},

	/**
	 * @description Updates the data in the popup.
	 * @param {Object} data - Raw popup data.
	 * @returns {void}
	 */
	show(data) {
		this.set(data, {silent: true});
		this.set({
			error: false,
			closed: false,
		});
	},

	/**
	 * @description Terminates the current example and resets the model.
	 * @returns {void}
	 */
	destroyExample() {
		this.get("examplesModel").set("example", undefined);
	},

	/**
	 * @description Highlights the specified tokens in the example model of this popup.
	 * @param {string} url - The ID of the ref.
	 * @returns {void}
	 */
	highlightTokens(url) {
		var cm = this.get("examplesModel").get("codeModel");
		if (cm.tokens) cm.highlightToken(url);
	},

	/**
	 * @description Clears all highlighted tokens.
	 * @returns {void}
	 */
	clearHighlightedTokens() {
		var cm = this.get("examplesModel").get("codeModel");
		if (cm.tokens) cm.clearHighlightedTokens();
	},
});
