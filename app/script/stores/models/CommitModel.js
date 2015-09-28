var Backbone = require("backbone");

var CommitModel = Backbone.Model.extend({
	defaults: {
		Author: {
			Name: "",
			Email: "",
			Date: 0,
		},

		AuthorPerson: {
			AvatarURL: "",
		},

		Message: "",

		ID: "",
	},
});

module.exports = CommitModel;
