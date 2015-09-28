var Backbone = require("backbone");

/*
 * @description CodeLineModel holds information about a line, including the tokens
 * that are contained within it.
 */
var CodeLineModel = Backbone.Model.extend({

	defaults: {
		/**
		 * @description The start of this line as a byte offset.
		 * @type {number}
		 */
		start: 0,

		/**
		 * @description The end of this line as a byte offset.
		 * @type {number}
		 */
		end: 0,

		/**
		 * @description The line number.
		 * @type {number}
		 */
		number: 0,

		/**
		 * @description The tokens contained on this line.
		 * @type {CodeTokenCollection}
		 */
		tokens: [],
	},

	/**
	 * @description Node holds a link to this view's DOM node.
	 * This property is IMMUTABLE and private. DOM manipulation
	 * should not occur via this. The only purpose it serves is
	 * to allow access to the element's (X, Y) position on the
	 * page and to obtain the width and height.
	 * @type {jQuery}
	 */
	__node: null,

	/**
	 * @description The ID attribute of the line is it's number in the snippet
	 * that it belongs too.
	 */
	idAttribute: "number",

	/**
	 * @description Retrieves the absolute position of this element as an object
	 * containing top and left pixel offsets.
	 * @returns {Object} - Object containing "top" and "left"
	 */
	getAbsolutePosition() {
		return this.__node.offset();
	},

	/**
	 * @description Retrieves the relative position of this element as an object
	 * containing top and left pixel offsets.
	 * @returns {Object} - Object containing "top" and "left"
	 */
	getRelativePosition() {
		return this.__node.position();
	},
});

module.exports = CodeLineModel;
