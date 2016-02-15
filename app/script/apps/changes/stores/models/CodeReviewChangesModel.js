var Backbone = require("backbone");

var FileDiffModel = require("../../../../stores/models/FileDiffModel");

/**
 * @description CodeReviewChangesModel holds the data needed to back the differential
 * view in a changeset.
 */
var CodeReviewChangesModel = Backbone.Model.extend({

	defaults: {
		/**
		 * @description Holds a collection of file diffs that are shown in the compare view.
		 * @type {Backbone.Collection|null}
		 */
		fileDiffs: null,

		/**
		 * @description Holds detailed information about the revisions of this changeset.
		 * For more info see sourcegraph.Delta (in the API).
		 * @type {Object}
		 */
		delta: null,
	},

	/**
	 * @description Loads new data into the store. It should map to sourcegraph.DeltaFiles.
	 * @param {Object} data - sourcegraph.DeltaFiles
	 * @returns {void}
	 */
	load(data) {
		var fdiffs = new Backbone.Collection(
			(data.FileDiffs || []).map(
				fileDiff => new FileDiffModel(fileDiff, {parse: true})
			)
		);

		this.set({
			delta: data.Delta,
			fileDiffs: fdiffs,
			stats: data.Stats,
			overThreshold: Boolean(data.OverThreshold),
		});
	},

	/**
	 * @description Triggers a scroll event towards the given file or line.
	 * @param {Object} fileOrLine - Any item that allows retrieving
	 * its absolute position.
	 * @returns {void}
	 */
	scrollTo(fileOrLine) {
		this.trigger("scrollTop", fileOrLine.getAbsolutePosition().top);
	},
});

module.exports = CodeReviewChangesModel;
