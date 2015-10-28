var Backbone = require("backbone");

var HunkModel = require("../models/HunkModel");

var FileDiffModel = Backbone.Model.extend({
	/**
	 * @description Node holds a link to this view's DOM node.
	 * This property is IMMUTABLE and private. DOM manipulation
	 * should not occur via this. The only purpose it serves is
	 * to allow access to the element's (X, Y) position on the
	 * page and to obtain the width and height for scrolling the
	 * codeview.
	 * @type {jQuery}
	 * @private
	 */
	__node: null,

	/**
	 * @description Retrieves the absolute position of this element as an object
	 * containing top and left pixel offsets.
	 * @returns {Object} - Object containing "top" and "left" absolute position.
	 */
	getAbsolutePosition() {
		return this.__node.offset();
	},

	/**
	 * @description Retrieves the relative position of this element as an object
	 * containing top and left pixel offsets.
	 * @returns {Object} - Object containing "top" and "left" relative position.
	 */
	getRelativePosition() {
		return this.__node.position();
	},

	initialize(args, opts) {
		if (!opts.parse) {
			console.warn("FileDiffModel should take parse: true if handed data from server");
		}
	},

	/**
	 * @description Highlights all tokens identified by the given URL in this file diff.
	 * @param {string} url - Token ID
	 * @returns {void}
	 */
	highlightToken(url) {
		this.get("Hunks").forEach(hunk => hunk.tokens.highlight(url));
	},

	/**
	 * @description Blurs all tokens in this file diff.
	 * @returns {void}
	 */
	blurTokens() {
		this.get("Hunks").forEach(hunk => hunk.tokens.clearHighlighted());
	},

	/**
	 * @description Selects all tokens identified by the given URL in this file diff.
	 * @param {string} url - Token ID
	 * @returns {void}
	 */
	selectToken(url) {
		this.get("Hunks").forEach(hunk => hunk.tokens.select(url));
	},

	/**
	 * @description Blurs all tokens in this file diff.
	 * @returns {void}
	 */
	clearSelected() {
		this.get("Hunks").forEach(hunk => hunk.tokens.clearSelected());
	},

	/**
	 * @description Merges the hunks at the given indexes by removing their headers and
	 * their context expansion controls.
	 * @param {number} indexDst - Index of hunk to merge into.
	 * @param {number} indexSrc - Index of hunk to merge from.
	 * @returns {void}
	 */
	merge(indexDst, indexSrc) {
		var hunks = this.get("Hunks");
		hunks.at(indexDst).set({expandBottom: false});
		hunks.at(indexSrc).set({header: false, expandTop: false});
	},

	/**
	 * @description Parses server data into the model. Backbone lifecycle method.
	 * @param {Object} fileDiff - raw server data.
	 * @returns {void}
	 */
	parse(fileDiff) {
		return {
			Stats: fileDiff.Stats,
			OrigName: fileDiff.OrigName,
			NewName: fileDiff.NewName,
			PreImage: fileDiff.PreImage,
			PostImage: fileDiff.PostImage,
			Hunks: new Backbone.Collection(
				(fileDiff.FileDiffHunks || []).map((hunk, i) => {
					hunk.parent = this;
					return new HunkModel(hunk, {parse: true});
				})
			),
		};
	},

	getHeadFilename() {
		return this.get("NewName") === "/dev/null" ? null : this.get("NewName");
	},

	getBaseFilename() {
		return this.get("OrigName") === "/dev/null" ? null : this.get("OrigName");
	},
});

module.exports = FileDiffModel;
