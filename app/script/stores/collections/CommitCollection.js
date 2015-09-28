var Backbone = require("backbone");
var CommitModel = require("../models/CommitModel");

var CommitCollection = Backbone.Collection.extend({
	load(commits) {
		this.add(commits.map(commit => new CommitModel(commit)));
	},
});

module.exports = CommitCollection;
