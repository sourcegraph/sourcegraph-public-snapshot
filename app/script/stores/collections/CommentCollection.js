var Backbone = require("backbone");
var CommentModel = require("../models/CommentModel");

var CommentCollection = Backbone.Collection.extend({
	model: CommentModel,
});

module.exports = CommentCollection;
