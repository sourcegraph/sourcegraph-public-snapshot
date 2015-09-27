var globals = require("../globals");
var notify = require("../components/notify");
var AppDispatcher = require("../dispatchers/AppDispatcher");

/**
 * @description Action called by server after receiving differential
 * data for a changeset.
 * @param {Object} data - sourcegraph.DeltaFiles
 * @returns {void}
 */
module.exports.receivedReviewChanges = function(data) {
	AppDispatcher.handleServerAction({
		type: globals.Actions.CR_RECEIVED_CHANGES,
		data: data,
	});
};

/**
 * @description Action called by server if the request for differential
 * data fails.
 * @param {*} data - Server response.
 * @returns {void}
 */
module.exports.receivedReviewChangesFailed = function(data) {
	notify.error(`Failed to load: ${data}`);
};

/**
 * @description Action triggered by server when receiving popopver data.
 * @param {string} data - HTML
 * @returns {void}
 */
module.exports.receivedPopover = function(data) {
	AppDispatcher.handleServerAction({
		type: globals.Actions.CR_RECEIVED_POPOVER,
		data: data,
	});
};

/**
 * @description Action called by server if the request for popover
 * data fails. Currently a no-op because we do not wish to take action.
 * @param {*} data - Server response.
 * @returns {void}
 */
module.exports.receivedPopoverFailed = function(data) {
	// noop - no need to disturb user
};

/**
 * @description Triggered when the request for context to a hunk has
 * been successful.
 * @param {HunkModel} hunk - The hunk that the context is for.
 * @param {bool} isTop - Whether the context is at the top.
 * @param {Object} data - The source code to be used for context.
 * @returns {void}
 */
module.exports.receivedHunkContext = function(hunk, isTop, data) {
	if (data.hasOwnProperty("Error")) {
		console.error(data.Error);
	}
	AppDispatcher.handleServerAction({
		type: globals.Actions.CR_RECEIVED_HUNK_CONTEXT,
		model: hunk,
		data: data,
		isTop: isTop,
	});
};

/**
 * @description Action called by server if the request for hunk context
 * data fails.
 * @returns {void}
 */
module.exports.receivedHunkContextFailed = function() {
	// TODO(gbbr) (backend) retrieve total number of lines of parent
	// file.
	// noop - currently we don't know if this means reaching an end of
	// file or an error because we don't know the total number of lines
	// in a file.
};

/**
 * @description Called when the request for popup data is successful.
 * @param {Object} data - Popup data.
 * @returns {void}
 */
module.exports.receivedPopup = function(data) {
	AppDispatcher.handleServerAction({
		type: globals.Actions.CR_RECEIVED_POPUP,
		data: data,
	});
};

/**
 * @description Action called by server if the request for popup
 * data fails.
 * @returns {void}
 */
module.exports.receivedPopupFailed = function() {
	// noop - no need to disturb user
};

module.exports.receivedExample = function(data) {
	AppDispatcher.handleServerAction({
		type: globals.Actions.CR_RECEIVED_EXAMPLE,
		data: data.data,
		page: data.page,
	});
};

/**
 * @description Action called by server if the request for example
 * data fails.
 * @returns {void}
 */
module.exports.receivedExampleFailed = function() {
	// noop - no need to disturb user
};

/**
 * @description Action called by server if the request for submitting
 * a review succeeds.
 * @param {*} data - Server response.
 * @returns {void}
 */
module.exports.submitReviewSuccess = function(data) {
	AppDispatcher.handleServerAction({
		type: globals.Actions.CR_SUBMIT_REVIEW_SUCCESS,
		data: data,
	});
};

/**
 * @description Action called by server if the request for submitting
 * a review fails.
 * @param {*} err - Server response.
 * @returns {void}
 */
module.exports.submitReviewFail = function(err) {
	notify.error("failed to submit review");
};

module.exports.statusUpdated = function(data) {
	AppDispatcher.handleServerAction({
		type: globals.Actions.CR_RECEIVED_CHANGED_STATUS,
		data: data,
	});
};

/**
 * @description Action called by server if the request for updating
 * a review's status fails.
 * @param {*} msg - Server response.
 * @returns {void}
 */
module.exports.statusUpdateFailed = function(msg) {
	if (msg.hasOwnProperty("Error")) msg = msg.Error;
	notify.error(msg);
};
