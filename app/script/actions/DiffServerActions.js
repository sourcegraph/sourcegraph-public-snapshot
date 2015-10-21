var AppDispatcher = require("../dispatchers/AppDispatcher");
var globals = require("../globals");
var notify = require("../components/notify");
var router = require("../routing/router");

/**
 * @description Action creator discarded when popover data is received from the server.
 * @param {Object} data - Raw popover data.
 * @returns {void}
 */
module.exports.receivedPopover = function(data) {
	AppDispatcher.handleServerAction({
		type: globals.Actions.DIFF_RECEIVED_POPOVER,
		data: data,
	});
};

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

/**
 * @description Action creator discarded when an example is received as a server response.
 * @param {Object} msg - Example data. Contains keys 'example' and 'page'.
 * @returns {void}
 */
module.exports.receivedExample = function(msg) {
	AppDispatcher.handleServerAction({
		type: globals.Actions.DIFF_RECEIVED_EXAMPLE,
		data: msg.example,
		page: msg.page,
	});
};

/**
 * @description This action creator dispatches the action that the server request
 * for an example has failed.
 * @param {jQuery.jqXHR} xhr - Request object.
 * @param {string} status - Status
 * @returns {void}
 */
module.exports.receivedExampleFailed = function(xhr, status) {
	// Noop. Fail silently.
};

module.exports.receivedPopoverFailed = function(data) {
	// Noop. Fail silently.
};

/**
 * @description Action creator discarded when popup data is received as a server response.
 * @param {Object} data - Raw popup data.
 * @returns {void}
 */
module.exports.receivedPopup = function(data) {
	AppDispatcher.handleServerAction({
		type: globals.Actions.DIFF_RECEIVED_POPUP,
		data: data,
	});
};

module.exports.receivedPopupFailed = function(data) {
	// Noop. Fail silently.
};
