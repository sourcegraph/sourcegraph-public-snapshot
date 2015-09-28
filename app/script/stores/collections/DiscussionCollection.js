var Backbone = require("backbone");
var DiscussionModel = require("../models/DiscussionModel");

/**
 * @description DiscussionCollection holds a collection of discussions.
 */
var DiscussionCollection = Backbone.Collection.extend({
	model: DiscussionModel,
});

module.exports = DiscussionCollection;
