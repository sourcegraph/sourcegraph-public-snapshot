var Backbone = require("backbone");

/**
 * @description Holds information about a discussion.
 */
var DiscussionModel = Backbone.Model.extend({

	defaults: {
		/**
		 * @type {number}
		 */
		ID: 0,

		/**
		 * @type {string}
		 */
		Title: "",

		/**
		 * @type {Date}
		 */
		CreatedAt: null,

		/**
		 * @type {sourcegraph.UserSpec}
		 */
		Author: {Login: null},

		/**
		 * @type {string}
		 */
		Description: "",

		/**
		 * @type {Array<sourcegraph.UserSpec>}
		 */
		Ratings: [],

		/**
		 * @description Holds an array of comments that have keys:
		 * ID, Author, CreatedAt and Body.
		 * @type {Array<DiscussionComment>}
		 */
		Comments: [],
	},

	/**
	 * @typedef {object} DiscussionComment
	 * @property {number} ID - Comment ID.
	 * @property {sourcegraph.UserSpec} Author - Comment author.
	 * @property {Date} CreatedAt - Comment creation date.
	 * @property {string} Body - Comment body..
	 */

	/**
	 * @description Adds a new comment to the discussion.
	 * @param {DiscussionComment} c - The comment to add.
	 * @returns {void}
	 */
	addComment(c) {
		var all = this.get("Comments");
		this.set("Comments", all.concat([c]));
	},
});

module.exports = DiscussionModel;
