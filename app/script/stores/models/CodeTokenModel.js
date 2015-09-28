var Backbone = require("backbone");
var globals = require("../../globals");

var _guid = 0;

var CodeTokenModel = Backbone.Model.extend({

	defaults: {
		/**
		 * @description The syntax highlighting type for this token.
		 * @type {string}
		 */
		syntax: "",

		/**
		 * @description Holds and extra classes that should be attached to this token.
		 * @type {string}
		 */
		extraClass: "",

		/**
		 * @description The HTML (encoded) value of the token.
		 * @type {string}
		 */
		html: "",

		/**
		 * @description The token type.
		 * @type {globals.TokenType}
		 */
		type: 0,

		/**
		 * @description The definition URL.
		 * @type {string}
		 */
		url: [],

		/**
		 * @description The line that the token belongs to. This is filled in
		 * by the CodeModel.
		 * @type {CodeLineModel}
		 */
		line: null,
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

	/**
	 * @description Backbone.Model method. Parses property and returns new one
	 * on set. See http://backbonejs.org/#Model-parse
	 * @param {Object} token - Raw token data as sent by server.
	 * @returns {Object} Parsed token to be saved as model attributes.
	 */
	parse(token) {
		var tokenType;
		if (!token.Class) {
			tokenType = globals.TokenType.STRING;
		} else if (token.hasOwnProperty("URL")) {
			tokenType = token.IsDef ? globals.TokenType.DEF : globals.TokenType.REF;
		} else {
			tokenType = globals.TokenType.SPAN;
		}

		return {
			syntax: token.Class,
			extraClass: token.ExtraClasses||"",
			html: typeof token === "string" ? token : token.Label,
			type: tokenType,
			url: token.URL,
			id: ++_guid,
		};
	},

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

module.exports = CodeTokenModel;
