var router = require("../../routing/router");
var Backbone = require("backbone");
var CodeModel = require("./CodeModel");

module.exports = Backbone.Model.extend({

	defaults: {
		codeModel: new CodeModel(),
	},

	/**
	 * @description Loads the passed example data into the store and CodeModel.
	 * @param {Object} data - Valid ExampleList data.
	 * @param {number} page - Page offset for this example data.
	 * @returns {void}
	 */
	showExample(data, page) {
		if (data && data.length) {
			var ex = data[0],
				cm = this.get("codeModel");

			if (!Array.isArray(data) || !data.length || ex.Error) {
				return this.set("error", true);
			}

			var defUrl = router.defURL(ex.DefRepo, (ex.Rev||ex.CommitID), ex.DefUnitType, ex.DefUnit, ex.DefPath);

			cm.load(ex);
			cm.selectToken(defUrl);

			this.set({
				example: ex,
				page: page,
			});
		} else if (page === 1) {
			this.set({example: null});
		}

		this.set("lastPage", !(data && data.length));
	},
});
