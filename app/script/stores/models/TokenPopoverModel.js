var Backbone = require("backbone");

var TokenPopoverModel = Backbone.Model.extend({
	defaults: {
		visible: false,

		body: null,

		position: {
			top: 0,
			left: 0,
		},

		offset: {
			top: 15,
			left: 15,
		},
	},

	positionAt(event) {
		var x = event.clientX, pw = 380; // popover width
		if (x > window.innerWidth-pw) x = window.innerWidth-pw;

		this.set({
			position: {
				top: event.clientY + this.get("offset").top,
				left: x + this.get("offset").left,
			},
		});
	},
});

module.exports = TokenPopoverModel;
