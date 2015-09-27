var Backbone = require("backbone");
var CommentCollection = require("../collections/CommentCollection");
var CurrentUser = require("../../CurrentUser");

var ReviewCollection = Backbone.Collection.extend({

	/**
	 * @description Holds the localStorage key where review information is stored.
	 * It is composed of a combination of the repo's path and the changeset ID.
	 * @type {string}
	 * @private
	 */
	_storageKey: null,

	/**
	 * @description Holds a collection of drafts that are currently on this page.
	 * @type {Backbone.Collection|null}
	 */
	drafts: null,

	/**
	 * @description Holds a collection of all the comments from all the reviews
	 * that are on this changeset.
	 * @type {Backbone.Collection|null}
	 */
	comments: null,

	initialize(models, opts) {
		this._storageKey = `review-draft/${opts.repo}/${opts.changesetId}${CurrentUser !== null ? `/${CurrentUser.Login}` : ""}`;

		this.comments = new CommentCollection();
		this.comments.on("add remove", () => this.trigger("change"));

		this.drafts = new CommentCollection();
		this.drafts.on("add remove", () => this.trigger("change"));

		this.add(models);
		this.addDraft(this._getStorage().drafts, true);
	},

	/**
	 * @description Adds a draft (or an array of drafts) to the store. If the
	 * second parameter is true, it will also update the localStorage to persist
	 * these drafts. The second parameter may be false in cases when we are adding
	 * drafts to the store, but we do not want them to be added to localStorage
	 * becuase that is already the source.
	 * @param {Array<Object>|Object} itemOrArray - Draft or array of drafts.
	 * @param {bool} silent - If true, drafts will not be persisted to local storage.
	 * @returns {void}
	 */
	addDraft(itemOrArray, silent) {
		var arr = Array.isArray(itemOrArray) ? itemOrArray : [itemOrArray];

		if (silent !== true) {
			var data = this._getStorage();
			// Assign UIDs
			arr.forEach((item, i) => item.id = data.guid + i + 1);
			// Save to storage and update GUID
			localStorage.setItem(this._storageKey, JSON.stringify({
				drafts: data.drafts.concat(arr),
				guid: data.guid + arr.length,
			}));
		}

		this.drafts.add(arr);
	},

	/**
	 * @description Updates the comment with the new body.
	 * @param {CommentModel} comment - The comment to update.
	 * @param {string} newBody - The text of the new body.
	 * @returns {void}
	 */
	updateDraft(comment, newBody) {
		var storage = this._getStorage(),
			all = new CommentCollection(storage.drafts),
			item = all.get(comment.get("id"));

		if (typeof item === "undefined") return;

		[comment, item].forEach(c => c.set({
			Body: newBody,
			editingComment: false,
		}));

		localStorage.setItem(this._storageKey, JSON.stringify({
			drafts: all.toJSON(),
			guid: storage.guid,
		}));
	},

	/**
	 * @description Updates the comment with the new body.
	 * @param {CommentModel} comment - The comment to update.
	 * @returns {void}
	 */
	deleteDraft(comment) {
		var storage = this._getStorage(),
			all = new CommentCollection(storage.drafts),
			id = comment.get("id"),
			item = all.get(id);

		if (typeof item === "undefined") return;

		all.remove(id);

		localStorage.setItem(this._storageKey, JSON.stringify({
			drafts: all.toJSON(),
			guid: storage.guid,
		}));

		this.drafts.remove(id);
	},

	/**
	 * @description Resets the collection and removes all drafts from the local
	 * storage. It also triggers a change event.
	 * @returns {void}
	 */
	clearDrafts() {
		localStorage.removeItem(this._storageKey);
		this.drafts.reset();
		this.trigger("change");
	},

	/**
	 * @description Returns all comments for the given line. This is used by the
	 * view to retrieve comments.
	 * @param {CodeLineModel} line - The line for which to return the comments.
	 * @returns {void}
	 */
	getLineComments(line) {
		var fd = line.get("fileDiff");

		if (this.comments.length === 0 && this.drafts.length === 0 || !fd) return [];

		var filterFn = comment => {
			var isValid = (fd.get("PostImage") === comment.get("CommitID") && fd.getHeadFilename() === comment.get("Filename")) ||
				(fd.get("PreImage") === comment.get("CommitID") && fd.getBaseFilename() === comment.get("Filename"));

			if (isValid) {
				comment.set("atHead", fd.get("PostImage") === comment.get("CommitID"));
			}
			return isValid;
		};

		var fileComments = this.comments.filter(filterFn),
			fileDrafts = this.drafts.filter(filterFn),
			all = fileComments.concat(fileDrafts);

		var result = all.filter(c => {
			var isOnHead = c.get("atHead") && line.get("lineNumberHead") === c.get("LineNumber");
			var isOnBase = !c.get("atHead") && line.get("lineNumberBase") === c.get("LineNumber");
			return isOnHead || isOnBase;
		});

		return result;
	},

	/**
	 * @description Returns all drafts that are in local storage.
	 * @returns {Array<Object>} Array of objects containing draft attributes.
	 * @private
	 */
	_getStorage() {
		var data = JSON.parse(localStorage.getItem(this._storageKey));
		return data && Array.isArray(data.drafts) ? data : {guid: 0, drafts: []};
	},

	/**
	 * @description Overrides the default add behavior by converting review comments into a collection
	 * and adding an extra property on each comment linking to the parent review, calling super at the end.
	 * @param {Backbone.Model|Array<Backbone.Model>} modelOrArray - The model (or array of models) to add.
	 * @returns {void}
	 */
	add(modelOrArray) {
		(Array.isArray(modelOrArray) ? modelOrArray : [modelOrArray]).forEach(model => {
			if (!model.Comments || !model.Comments.length) return;

			var modelCollection = new Backbone.Collection(model.Comments.map(comment => {
				comment.parent = model;
				return this.comments.add(comment, {silent: true});
			}));

			model.Comments = modelCollection;
		});

		Reflect.apply(Backbone.Collection.prototype.add, this, [modelOrArray]);
	},
});

module.exports = ReviewCollection;
