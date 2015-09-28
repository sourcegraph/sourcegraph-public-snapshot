var Backbone = require("backbone");

var ContextMenuModel = Backbone.Model.extend({
	defaults: {
		options: [],
		position: {
			top: 0,
			left: 0,
		},
		closed: true,
	},

	addOption(opt) {
		var opts = this.get("options");
		if (!opt.hasOwnProperty("data") || !opt.hasOwnProperty("label")) {
			return console.error("ContextMenuModel: Specified addition of option without 'label' and 'data'");
		}
		this.set("options", opts.concat(opt));
	},
});

module.exports = ContextMenuModel;
