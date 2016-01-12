var Backbone = require("backbone");
var globals = require("../../globals");

var CodeTokenCollection = require("../collections/CodeTokenCollection");
var CodeTokenModel = require("../models/CodeTokenModel");
var CodeLineModel = require("../models/CodeLineModel");

var HunkModel = Backbone.Model.extend({

	defaults: {
		/**
		 * @description Sets the visibility of the hunk header.
		 * @type {bool}
		 */
		header: true,

		/**
		 * @description Sets the visibility of the control that expands context
		 * at the top of the hunk.
		 * @type {bool}
		 */
		expandTop: true,

		/**
		 * @description Sets the visibility of the control that expands context
		 * at the bottom of the hunk.
		 * @type {bool}
		 */
		expandBottom: true,

		/**
		 * @description Holds the cid's of lines that have an open comment form.
		 * @type {Array<string>}
		 */
		commentsAt: [],
	},

	/**
	 * @description Holds all tokens that are refs (or defs) shown in this hunk's code.
	 * @type {CodeTokenCollection|null}
	 */
	tokens: null,

	initialize(args, opts) {
		if (typeof opts === "undefined" || !opts.parse) {
			console.warn("FileDiffModel should take parse: true if handed data from server");
		}
	},

	/**
	 * @description Returns the index of this hunk within its parent FileDiff.
	 * @returns {number} Index
	 */
	index() {
		return this.get("Parent").get("Hunks").indexOf(this);
	},

	/**
	 * @description Registers all tokens within the collection of this hunk and returns an array
	 * of CodeTokenModels.
	 * @param {Array<Object>} tokens - Raw token data
	 * @returns {Array<CodeTokenModel>} Array of tokens.
	 * @private
	 */
	_registerTokens(tokens) {
		return Array.isArray(tokens) && tokens.length ? tokens.map(token => {
			var t = new CodeTokenModel(token, {parse: true});
			var type = t.get("type");
			if (type === globals.TokenType.REF || type === globals.TokenType.DEF) {
				this.tokens.put(t);
			}
			return t;
		}) : [];
	},

	/**
	 * @description Creates an array of CodeLineModels to be presented as context for a hunk.
	 * @param {Array<Object>} lines - Raw line data
	 * @param {number} baseStart - Starting line number for base revision.
	 * @param {number} headStart - Starting line number for head revision.
	 * @returns {Array<CodeLineModel>} Array of lines.
	 * @private
	 */
	_newContextLines(lines, baseStart, headStart) {
		return lines.map((line, i) =>
			new CodeLineModel({
				tokens: this._registerTokens(line.Tokens),
				prefix: " ",
				extraClass: "gray",
				lineNumberBase: baseStart + i,
				lineNumberHead: headStart + i,
				contextLine: true,
			})
		);
	},

	/**
	 * @description Updates the top of the hunk with new context received from the server.
	 * @param {Object} data - Raw server data containing entry snippet of code for context.
	 * @returns {void}
	 */
	updateTop(data) {
		this.validateHunkExpansionTreeEntry(data);

		var totalLines = data.Entry.EndLine - data.Entry.StartLine,
			origStartLine = this.get("OrigStartLine") - totalLines - 1,
			newStartLine = this.get("NewStartLine") - totalLines - 1;

		this.set({
			LinePrefixes: this.get("LinePrefixes") + Array(totalLines + 1).join(" "),
			OrigStartLine: origStartLine,
			NewStartLine: newStartLine,
			Section: "",
			NewLines: this.get("NewLines") + totalLines + 1,
			OrigLines: this.get("OrigLines") + totalLines + 1,
			Lines: this._newContextLines(data.Entry.SourceCode.Lines, origStartLine, newStartLine).concat(this.get("Lines")),
		});
	},

	/**
	 * @description Updates the bottom of the hunk with new context received from the server.
	 * @param {Object} data - Raw server data containing entry snippet of code for context.
	 * @returns {void}
	 */
	updateBottom(data) {
		this.validateHunkExpansionTreeEntry(data);

		var totalLines = data.Entry.EndLine - data.Entry.StartLine,
			origStartLine = this.get("OrigStartLine") + this.get("OrigLines"),
			newStartLine = this.get("NewStartLine") + this.get("NewLines");

		this.set({
			LinePrefixes: this.get("LinePrefixes") + Array(totalLines + 1).join(" "),
			NewLines: this.get("NewLines") + totalLines + 1,
			OrigLines: this.get("OrigLines") + totalLines + 1,
			Lines: this.get("Lines").concat(this._newContextLines(data.Entry.SourceCode.Lines, origStartLine, newStartLine)),
		});
	},

	/**
	 * @description Validates that a sourcegraph.TreeEntry has an Entry, Entry.EndLine, and Entry.StartLine fields.
	 * @param {Object} data - The tree entry object for validation.
	 * @returns {void}
	 */
	validateHunkExpansionTreeEntry(data) {
		if (typeof data.Entry === "undefined") {
			console.error("hunk expansion has failed: data.Entry is undefined; Please report this issue immediately!");
		}
		if (typeof data.Entry.EndLine === "undefined") {
			console.error("hunk expansion has failed: data.Entry.EndLine is undefined; Please report this issue immediately!");
		}
		if (typeof data.Entry.StartLine === "undefined") {
			console.error("hunk expansion has failed: data.Entry.StartLine is undefined; Please report this issue immediately!");
		}
	},

	/**
	 * @description Backbone lifecycle method. Parses raw hunk data and populates the model.
	 * @param {Object} hunk - Raw hunk data.
	 * @returns {void}
	 */
	parse(hunk) {
		// sanity checks
		if (!hunk.BodySource) return null;
		if (hunk.LinePrefixes.length > hunk.BodySource.Lines.length) {
			console.error("LinePrefixes > BodySource.Lines in:", hunk);
			throw new Error("LinePrefixes entries length different from Body LOC");
		}

		this.tokens = new CodeTokenCollection();
		this.lines = null; // TODO(gbbr): Line collection

		var baseLineCount = hunk.OrigStartLine;
		var headLineCount = hunk.NewStartLine;
		var currentLabelBase = baseLineCount||"";
		var currentLabelHead = headLineCount||"";

		this.lines = hunk.LinePrefixes.split("").map((p, i) => {
			switch (p) {
			case "+":
				if (i > 0) {
					currentLabelHead = ++headLineCount;
					currentLabelBase = "";
				}
				break;

			case "-":
				if (i > 0) {
					currentLabelHead = "";
					currentLabelBase = ++baseLineCount;
				}
				break;

			case " ":
				if (i > 0) {
					currentLabelHead = ++headLineCount;
					currentLabelBase = ++baseLineCount;
				}
				break;

			default:
				throw new Error("Invalid line prefix in data.");
			}

			return new CodeLineModel({
				tokens: this._registerTokens(hunk.BodySource.Lines[i].Tokens),
				prefix: hunk.LinePrefixes[i],
				lineNumberBase: currentLabelBase,
				lineNumberHead: currentLabelHead,
				fileDiff: hunk.parent,
			});
		});

		return {
			LinePrefixes: hunk.LinePrefixes,
			OrigStartLine: hunk.OrigStartLine,
			OrigLines: hunk.OrigLines,
			NewStartLine: hunk.NewStartLine,
			NewLines: hunk.NewLines,
			Section: hunk.Section,
			Parent: hunk.parent,
			Lines: this.lines,
		};
	},

	openComment(line) {
		if (Array.isArray(this.get("commentsAt")) && this.get("commentsAt").indexOf(line.cid) > -1) {
			return;
		}

		this.set({commentsAt: (this.get("commentsAt") || []).concat([line.cid])});
	},

	closeComment(line) {
		var c = this.get("commentsAt") || [];
		this.set({
			// the below operation creates a new array with the element removed without altering
			// the state object.
			commentsAt: c.slice(0, c.indexOf(line.cid)).concat(c.slice(c.indexOf(line.cid)+1)),
		});
	},
});

module.exports = HunkModel;
