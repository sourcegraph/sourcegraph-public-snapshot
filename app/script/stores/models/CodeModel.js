var Backbone = require("backbone");
var globals = require("../../globals");

var CodeLineModel = require("./CodeLineModel");
var CodeLineCollection = require("../collections/CodeLineCollection");

var CodeTokenModel = require("./CodeTokenModel");
var CodeTokenCollection = require("../collections/CodeTokenCollection");

/**
 * @description CodeModel holds information necessary to display a CodeView.
 * To initialize the CodeModel, a valid entry (generally TreeEntry) must be passed
 * to its "load" function. A valid entry contains at minimum the sourcegraph.SourceCode
 * structure and optionally a StartLine.
 */
var CodeModel = Backbone.Model.extend({

	defaults: {
		/**
		 * @description Lines is the property used by the view to render itself.
		 * It is expected to contain a valid CodeLineCollection.
		 * @type {CodeLineCollection}
		 */
		lines: new CodeLineCollection([]),
	},

	/**
	 * @description Holds a collection of all the tokens contained in this model.
	 * @type {CodeTokenCollection}
	 */
	tokens: null,

	/**
	 * @description Loads the passed entry into the model by replacing the
	 * current state. The entry is expected to have at minimum a SourceCode key
	 * and optionally a StartLine (ie. sourcegraph.TreeEntry or sourcegraph.DefExample)
	 * @param {Object} entry - the entry to load
	 * @returns {void}
	 */
	load(entry) {
		var startLine = entry.StartLine || 1;
		this.tokens = new CodeTokenCollection();
		this.set("lines", new CodeLineCollection(
			entry.SourceCode.Lines.map(line => {
				var lineModel = new CodeLineModel({
					start: line.StartByte,
					end: line.EndByte,
					number: startLine++,
				});

				lineModel.set({
					tokens: Array.isArray(line.Tokens) && line.Tokens.length ? this._registerTokens(line.Tokens, lineModel) : [],
				}, {silent: true});

				return lineModel;
			})
		));
	},

	/**
	 * @description Registers the array of passed tokens with the global token collection
	 * and additionally returns them as a valid CodeTokenCollection.
	 * @param {Array} tokens - Array of tokens as sent via sourcegraph.SourceCode
	 * @param {CodeLineModel} lineModel - The line that the tokens belong to (so we can
	 * link each token back to its owner.
	 * @returns {CodeTokenCollection} - Collection of tokens.
	 * @private
	 */
	_registerTokens(tokens, lineModel) {
		return tokens.map(token => {
			var tokenModel = new CodeTokenModel(token, {parse: true});
			var type = tokenModel.get("type");
			if (type === globals.TokenType.REF || type === globals.TokenType.DEF) {
				tokenModel.set({line: lineModel}, {silent: true});
				this.tokens.put(tokenModel);
			}
			return tokenModel;
		});
	},

	/**
	 * @description Returns the token that is the definition with the passed URL, if available
	 * in this CodeModel.
	 * @param {string} url - Definition URL
	 * @returns {CodeTokenModel} - Token model.
	 */
	getDefinition(url) {
		return this.tokens.getDefinition(url);
	},

	/**
	 * @description Retrieves an array of highlighted lines, if any. Otherwise returns null
	 * @returns {Array<CodeLineCollection>|null} - Array of lines, if any.
	 */
	getHighlightedLines() {
		return this.get("lines").getHighlighted();
	},

	/**
	 * @description Highlights the passed line range and returns all highlighted lines.
	 * @param {number} start - Start line.
	 * @param {number} end - End line.
	 * @returns {Array<CodeLineModel>} - Array of highlighted lines, if any.
	 */
	highlightLineRange(start, end) {
		return this.get("lines").highlightRange(start, end);
	},

	/**
	 * @description Highlights the passed byte range and returns all highlighted lines.
	 * @param {number} start - Start byte.
	 * @param {number} end - End byte.
	 * @returns {Array<CodeLineModel>} - Array of highlighted lines, if any.
	 */
	highlightByteRange(start, end) {
		return this.get("lines").highlightByteRange(start, end);
	},

	/**
	 * @description Clears all highlighted lines.
	 * @returns {void}
	 */
	clearHighlightedLines() {
		return this.get("lines").clearHighlighted();
	},

	/**
	 * @description Highlights all the tokens that have the passed definition URL.
	 * @param {string} url - The definition URL for the tokens to be highlighted.
	 * @returns {void}
	 */
	highlightToken(url) {
		this.tokens.highlight(url);
	},

	/**
	 * @description Clears all token highlights.
	 * @returns {void}
	 */
	clearHighlightedTokens() {
		this.tokens.clearHighlighted();
	},

	/**
	 * @description Selects all the tokens that have the passed definition URL.
	 * @param {string} url - The definition URL for the tokens to be selected.
	 * @returns {void}
	 */
	selectToken(url) {
		this.tokens.select(url);
	},

	/**
	 * @description Clears all selected tokens.
	 * @returns {void}
	 */
	clearSelectedTokens() {
		this.tokens.clearSelected();
	},

	destroy() {
		console.info("TODO(gbbr): Clean-up");
	},
});

module.exports = CodeModel;
