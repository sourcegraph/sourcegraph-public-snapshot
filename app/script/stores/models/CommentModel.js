var Backbone = require("backbone");

var CommentModel = Backbone.Model.extend({
	defaults: {
		editingComment: false,
	},

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

	isDraft() {
		return Boolean(this.get("Draft"));
	},

	/**
	 * @description Retrieves the absolute position of this element as an object
	 * containing top and left pixel offsets.
	 * @returns {Object} An object containing keys "top" and "left" absolute position.
	 */
	getAbsolutePosition() {
		return this.__node.offset();
	},

	/**
	 * @description Retrieves the relative position of this element as an object
	 * containing top and left pixel offsets.
	 * @returns {Object} An object containing keys "top" and "left" relative position.
	 */
	getRelativePosition() {
		return this.__node.position();
	},
});

module.exports = CommentModel;
