var Backbone = require("backbone");
var globals = require("../../globals");

/**
 * @description CodeTokenCollection holds a collection of tokens which are grouped
 * and indexed for fast and easy access.
 */
var CodeTokenCollection = Backbone.Collection.extend({
	/**
	 * @description Maps IDs to tokens. Indexing for easy and fast access.
	 * @type {Object}
	 * @private
	 */
	_byId: null,

	/**
	 * @description GUID which is incremented with each token. If source code
	 * doesn't change, each token should always hold the same ID on each visit.
	 * @private
	 */
	_guid: null,

	/**
	 * @description Selected tokens maps the highlight mode to the URL of the highlighted
	 * tokens. This is a speed optimization to avoid traversal.
	 * @type {Object}
	 * @private
	 */
	_selectedTokens: null,

	/**
	 * @description Backbone lifecycle method. Constructor.
	 * @returns {void}
	 * @private
	 */
	initialize() {
		this._selectedTokens = {};
		this._byId = {};
		this._guid = 0;
	},

	/**
	 * @description Puts a new method into the collection, indexes it and categorizes it
	 * into the group of tokens where it belongs, based on whether it is a reference or
	 * definition. This is a speed optimization.
	 * @param {CodeTokenModel|Array<CodeTokenModel>} tokenOrArray - The token or array to add.
	 * @returns {void}
	 */
	put(tokenOrArray) {
		if (Array.isArray(tokenOrArray)) {
			tokenOrArray.forEach(this._put, this);
		} else if (typeof tokenOrArray === "object") {
			this._put(tokenOrArray);
		}
	},

	_put(token) {
		var urls = token.get("url");

		urls.forEach(url => {
			if (!this.get(url)) {
				this.add({id: url, def: [], refs: []});
			}

			if (token.get("type") === globals.TokenType.DEF) {
				var defs = this.get(url).get("def");
				this.get(url).set("def", defs.concat(token));
			} else {
				var refs = this.get(url).get("refs");
				this.get(url).set("refs", refs.concat(token));
			}
		});

		token.tid = this._guid++;
		this._byId[token.tid] = token;
	},

	/**
	 * @description Gets the definitions for the token identified by the passed URL. If
	 * the definition is not available in the current collection, an empty array is returned.
	 * @param {string} url - URL of token to find definitions for.
	 * @returns {Array} Array of definitions for the URL given.
	 */
	getDefinition(url) {
		var tokens = this.get(url);
		return tokens ? tokens.get("def") : null;
	},

	/**
	 * @description Sets all the tokens with the passed URL to highlighted.
	 * @param {string} url - Definition URL.
	 * @returns {void}
	 */
	highlight(url) { this._focusTokens(url, "highlighted"); },

	/**
	 * @description Clears all highlighted tokens.
	 * @returns {void}
	 */
	clearHighlighted() { this._blurTokens("highlighted"); },

	/**
	 * @description Sets all the tokens with the passed URL to selected.
	 * @param {string} url - Definition URL.
	 * @returns {void}
	 */
	select(url) { this._focusTokens(url, "selected"); },

	/**
	 * @description Clears all selected tokens.
	 * @returns {void}
	 */
	clearSelected() { this._blurTokens("selected"); },

	/**
	 * @description Returns the token having the given ID.
	 * @param {number} id - Token ID.
	 * @returns {void}
	 */
	byId(id) { return this._byId[id]; },

	/**
	 * @description Highlights the tokens that point to the given URL/Def.
	 * This is an expensive operation.
	 * @param {string} url - Def URL of token to highlight
	 * @param {string} mode - Highlight mode ("highlighted", "selected")
	 * @returns {void}
	 * @private
	 */
	_focusTokens(url, mode) {
		if (url === this._selectedTokens[mode]) return;

		this._blurTokens(mode);

		var tokens = this.get(url);
		if (tokens) {
			tokens.get("def").forEach(d => d.set(mode, true));
			tokens.get("refs").forEach(ref => ref.set(mode, true));
		}

		this._selectedTokens[mode] = url;
	},

	/**
	 * @description Clears the tokens highlighted with the passed mode, if any.
	 * @param {string} mode - Highlight mode to clear.
	 * @returns {void}
	 * @private
	 */
	_blurTokens(mode) {
		if (this._selectedTokens[mode]) {
			var tokens = this.get(this._selectedTokens[mode]);
			if (tokens) {
				tokens.get("def").forEach(d => d.set(mode, false));
				tokens.get("refs").forEach(ref => ref.set(mode, false));
			}

			this._selectedTokens[mode] = null;
		}
	},
});

module.exports = CodeTokenCollection;
