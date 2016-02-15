var AppDispatcher = require("../dispatchers/AppDispatcher");
var globals = require("../globals");
var notify = require("../components/notify");
var router = require("../routing/router");

module.exports.receivedChangesetCreate = function(data) {
	AppDispatcher.handleServerAction({
		type: globals.Actions.DIFF_PROPOSE_CHANGE_SUCCESS,
		data: data,
	});

	window.location = router.changesetURL(data.Repo, data.ID);
};

module.exports.receivedChangesetCreateFailed = function(data) {
	notify.error("Failed to create changeset");
};

/**
 * @description Action creator discarded when context is received for the top of a hunk.
 * @param {HunkModel} model - The model of the hunk the context belongs too.
 * @param {Object} data - The entry containing the lines of code in the context.
 * @returns {void}
 */
module.exports.receivedHunkTop = function(model, data) {
	if (data.hasOwnProperty("Error")) {
		console.error(data.Error);
		return notify.error("Could not retrieve context.");
	}

	AppDispatcher.handleServerAction({
		type: globals.Actions.DIFF_RECEIVED_HUNK_TOP,
		model: model,
		data: data,
	});
};

/**
 * @description Action creator discarded when context is received for the bottom of a hunk.
 * @param {HunkModel} model - The model of the hunk the context belongs too.
 * @param {Object} data - The entry containing the lines of code in the context.
 * @returns {void}
 */
module.exports.receivedHunkBottom = function(model, data) {
	if (data.hasOwnProperty("Error")) {
		console.error(data.Error);
		return notify.error("Could not retrieve context.");
	}

	AppDispatcher.handleServerAction({
		type: globals.Actions.DIFF_RECEIVED_HUNK_BOTTOM,
		model: model,
		data: data,
	});
};

module.exports.failedReceiveExpansion = function() {
	// noop TODO(gbbr): return line count from VCSStore on error
};
