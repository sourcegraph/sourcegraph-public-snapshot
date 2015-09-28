var Backbone = require("backbone");

var FileDiffModel = require("./FileDiffModel");
var TokenPopoverModel = require("./TokenPopoverModel");
var CodeReviewPopupModel = require("./CodeReviewPopupModel");

/**
 * @description CodeReviewChangesModel holds the data needed to back the differential
 * view in a changeset.
 */
var CodeReviewChangesModel = Backbone.Model.extend({

	defaults: {
		/**
		 * @description Holds a collection of file diffs that are shown in the compare view.
		 * @type {Backbone.Collection|null}
		 */
		fileDiffs: null,

		/**
		 * @description Holds the popup model state.
		 * @type {TokenPopupModel}
		 */
		popupModel: new CodeReviewPopupModel(),

		/**
		 * @description Holds the popover state for the compare view.
		 * @type {TokenPopoverModel}
		 */
		popoverModel: new TokenPopoverModel(),

		/**
		 * @description Holds detailed information about the revisions of this changeset.
		 * For more info see sourcegraph.Delta (in the API).
		 * @type {Object}
		 */
		delta: null,
	},

	/**
	 * @description Loads new data into the store. It should map to sourcegraph.DeltaFiles.
	 * @param {Object} data - sourcegraph.DeltaFiles
	 * @returns {void}
	 */
	load(data) {
		var fdiffs = new Backbone.Collection(
			(data.FileDiffs || []).map(
				fileDiff => new FileDiffModel(fileDiff, {parse: true})
			)
		);

		this.set({
			delta: data.Delta,
			fileDiffs: fdiffs,
			stats: data.Stats,
			overThreshold: Boolean(data.OverThreshold),
		});
	},

	/**
	 * @description Updates the popover data.
	 * @param {string} data - HTML to display in popover.
	 * @returns {void}
	 */
	updatePopover(data) {
		this.get("popoverModel").set({
			visible: true,
			body: data,
		});
	},

	/**
	 * @description Clears highlighted tokens. If an array of FileDiffs is given
	 * it only clears highlights within that file.
	 * @param {CodeTokenModel} token - The token model (this is not used but is
	 * passed by the action dispatcher).
	 * @param {Event} evt - The (click) event.
	 * @param {Array<FileDiffModel>|undefined} fds - The file diffs where to blur.
	 * @returns {void}
	 */
	blurTokens(token, evt, fds) {
		this.get("popupModel").clearHighlightedTokens();

		(Array.isArray(fds) ? fds : this.get("fileDiffs")).forEach(
			fd => fd.blurTokens()
		);

		this.get("popoverModel").set({visible: false});
	},

	/**
	 * @description Highlights tokens. If an array of FileDiffs is given
	 * it only clears highlights within that file.
	 * @param {CodeTokenModel} token - The token model (this is not used but is
	 * passed by the action dispatcher).
	 * @param {Event} evt - The (click) event.
	 * @param {Array<FileDiffModel>|undefined} fds - The file diffs where to highlight.
	 * @returns {void}
	 */
	focusToken(token, evt, fds) {
		this.get("popupModel").highlightTokens(token.get("url")[0]);

		(Array.isArray(fds) ? fds : this.get("fileDiffs")).forEach(
			fd => fd.highlightToken(token.get("url")[0])
		);

		this.get("popoverModel").positionAt(evt);
	},

	/**
	 * @description Triggers a scroll event towards the given file or line.
	 * @param {Object} fileOrLineOrToken - Any item that allows retrieving
	 * its absolute position.
	 * @returns {void}
	 */
	scrollTo(fileOrLineOrToken) {
		this.trigger("scrollTop", fileOrLineOrToken.getAbsolutePosition().top);
	},

	/**
	 * @description Selects the given token.
	 * @param {CodeTokenModel} token - The token model (this is not used but is
	 * passed by the action dispatcher).
	 * @param {Event} evt - The (click) event.
	 * @param {Array<FileDiffModel>|undefined} fd - The file diffs where the event
	 * occurred.
	 * @returns {void}
	 */
	selectToken(token, evt, fd) {
		this.get("popupModel").set({closed: true});
		this.get("popupModel").destroyExample();
		this.get("popoverModel").set({visible: false});
		this.get("fileDiffs").forEach(diff => diff.selectToken(token.get("url")[0]));
	},

	/**
	 * @description Deselects all tokens in the view.
	 * @returns {void}
	 */
	deselectTokens() {
		this.get("popupModel").destroyExample();
		this.get("popupModel").set({closed: true});
		this.get("fileDiffs").forEach(fd => fd.clearSelected());
	},

	/**
	 * @description Shows the popup with the given data.
	 * @param {Object} data - Popup model data.
	 * @returns {void}
	 */
	showPopup(data) {
		this.get("popupModel").show(data);
	},

	/**
	 * @description Shows the example data at the passed page.
	 * @param {Array<Object>} data - Array of examples.
	 * @param {number} page - Page number to show.
	 * @returns {void}
	 */
	showExample(data, page) {
		var pm = this.get("popupModel");
		pm.setLoading(false);
		pm.showExample(data, page);
	},
});

module.exports = CodeReviewChangesModel;
