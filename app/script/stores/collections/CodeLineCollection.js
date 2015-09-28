var Backbone = require("backbone");

var CodeLineModel = require("../models/CodeLineModel");

/**
 * @description CodeLineCollection holds a snippet of code.
 */
var CodeLineCollection = Backbone.Collection.extend({

	model: CodeLineModel,

	/**
	 * @description Stores an array of the currently highlighted lines.
	 * This is an optimization to avoid traversing the entire collection.
	 * @type {Array<CodeLineModel>}
	 * @private
	 */
	_highlightedLines: null,

	/**
	 * @description Highlight lines contained between the given byte range and returns them.
	 * @param {number} byteStart - Start byte range
	 * @param {number} byteEnd - End byte range
	 * @returns {Array<CodeLineModel>} - Array of lines.
	 */
	highlightByteRange(byteStart, byteEnd) {
		this.clearHighlighted();

		this._highlightedLines = this.filter(line => {
			var start = line.get("start");
			var	end = line.get("end");
			return ((byteStart >= start && byteStart <= end) ||
				(end <= byteEnd && start >= byteStart) ||
					(byteEnd >= start && byteEnd <= end));
		});

		this._highlightedLines.forEach(l => l.set("highlight", true));

		return this._highlightedLines;
	},

	/**
	 * @description Highlights the lines between the given line number range and returns them.
	 * @param {number} start - Start line
	 * @param {number} end - End line
	 * @returns {Array<CodeLineModel>} - Array of lines.
	 */
	highlightRange(start, end) {
		this.clearHighlighted();

		this._highlightedLines = [];
		for (var i = start; i <= end; i++) {
			var line = this.get(i);
			if (line) {
				line.set("highlight", true);
				this._highlightedLines.push(line);
			}
		}

		return this._highlightedLines;
	},

	/**
	 * @description Retrieves an array of highlighted lines, if any. Otherwise returns null.
	 * @returns {Array<CodeLineCollection>|null} - Array of highlighted lines, if any.
	 */
	getHighlighted() {
		return this._highlightedLines;
	},

	/**
	 * @description Clears highlighted lines, if any.
	 * @returns {void}
	 */
	clearHighlighted() {
		if (!Array.isArray(this._highlightedLines)) {
			return;
		}

		this._highlightedLines.forEach(l => l.set("highlight", false));
		this._highlightedLines = null;
	},
});

module.exports = CodeLineCollection;
