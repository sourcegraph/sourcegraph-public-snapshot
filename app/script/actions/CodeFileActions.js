var AppDispatcher = require("../dispatchers/AppDispatcher");
var CodeUtil = require("../util/CodeUtil");
var CodeFileStore = require("../stores/CodeFileStore");
var TokenPopupStore = require("../stores/TokenPopupStore");
var globals = require("../globals");
var router = require("../routing/router");
var notify = require("../components/notify");

/**
 * @description Loads the file at the given URL.
 * @param {string} url - URL of file to load
 * @returns {jQuery.jqXHR} - Promise.
 */
module.exports.selectFile = function(url) {
	return loadFile(url, globals.Actions.FETCH_FILE);
};

/**
 * @description Emaulates the file loading actions without requesting
 * data from the server. This happens when file data is served on page
 * load. It is a speed optimization.
 * @param {Object} data - File data
 * @returns {jQuery.Deferred} - promise
 */
module.exports.renderPreloaded = function(data) {
	return AppDispatcher.dispatchAsync(CodeUtil.receivedFile(JSON.parse(data.data)), {
		started: globals.Actions.FETCH_FILE,
		success: globals.Actions.RECEIVED_FILE,
		failure: globals.Actions.RECEIVED_FILE_FAILED,
	}).then(handleSuccess, handleFailure);
};

/**
 * @description Selects the passed token model (triggered by click event).
 * @param {CodeTokenModel} token - Token to select.
 * @param {Object=} opts - Additional options to be passed (such as source code
 * view for situations when there are multiple codeviews on the same page or
 * silent navigation).
 * @returns {void}
 */
module.exports.selectToken = function(token, opts) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.TOKEN_SELECT,
		token: token,
		opts: opts,
	});

	var list = token.get("url");

	if (!Array.isArray(list) || list.length < 1) {
		console.error("got definition with no URLs: ", token);
	} else if (list.length === 1) {
		showDefinition(list[0]);
	} else {
		showDefinitionList(token, list);
	}
};

/**
 * @description Dispatches the code file click event. This is used
 * in actions like hiding context menus when the user clicks outside
 * of it.
 * @param {Event} e - event
 * @returns {void}
 */
module.exports.focusCodeView = function(e) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.CODE_FILE_CLICK,
		event: e,
	});
};

/**
 * @description Switches the popup to an alternative definition. This applies
 * only when multiple definitions share the same space (ex. Scala).
 * @param {string} defKey - DefKey URL form.
 * @returns {void}
 */
module.exports.selectAlternativeDefinition = function(defKey) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.SWITCH_POPUP_DEFINITION,
		url: defKey,
	});

	showDefinition(defKey);
};

/**
 * @description Removes all token selections.
 * @returns {void}
 */
module.exports.deselectTokens = function() {
	AppDispatcher.handleViewAction({
		type: globals.Actions.TOKEN_CLEAR,
	});
};

/**
 * @description Dispatches an action that passes the highlighted token,
 * along with the highlight event (ie. mouseover)
 * @param {CodeTokenModel} token - The focused token
 * @param {Event} evt - The focus event
 * @returns {void}
 */
module.exports.focusToken = function(token, evt) {
	if (token.get("selected")) return;

	AppDispatcher.handleViewAction({
		type: globals.Actions.TOKEN_FOCUS,
		event: evt,
		token: token,
	});

	AppDispatcher.dispatchAsync(CodeUtil.fetchPopover(token.get("url")[0]), {
		started: null,
		success: globals.Actions.RECEIVED_POPOVER,
		failure: globals.Actions.RECEIVED_POPOVER_FAILED,
	});
};

/**
 * @description Dispatches an action that tokens have lost focus.
 * @returns {void}
 */
module.exports.blurTokens = function() {
	AppDispatcher.handleViewAction({
		type: globals.Actions.TOKEN_BLUR,
	});

	CodeUtil.abortPopoverXhr();
};

/**
 * @description Shows the passed definition. If the definition is in a different
 * file, that file is loaded first.
 * @param {Object} def - Definition Info object.
 * @returns {void}
 */
module.exports.navigateToDefinition = function(def) {
	dispatchIfSameFile(def.File, {
		type: globals.Actions.SHOW_DEFINITION,
		params: def,
	});
};

/**
 * @description Creates a new example request for a given definition
 * @param {string} defKey - The URL of the definition to request examples for.
 * @param {number} page - The example page.
 * @returns {void}
 */
module.exports.selectExample = function(defKey, page) {
	var file = CodeFileStore.get("file");
	var repoURI = (file && file.RepoRev) ? file.RepoRev.URI : null;

	AppDispatcher.dispatchAsync(CodeUtil.fetchExample(defKey, page, repoURI), {
		started: globals.Actions.FETCH_EXAMPLE,
		success: globals.Actions.RECEIVED_EXAMPLE,
		failure: globals.Actions.RECEIVED_EXAMPLE_FAILED,
	});
};

/**
 * @description Highlights the line range in the given file and optionally
 * highlights the passed definition.
 * @param {CodeFileStore.File} file - The file where the snippet is.
 * @param {number} startLine - The starting line for the highlight
 * @param {number} endLine - The end line for highlighting.
 * @param {string=} defKey - The URL of the definition to highlight.
 * @returns {void}
 */
module.exports.changeState = function(file, startLine, endLine, defKey) {
	dispatchIfSameFile(file, {
		type: globals.Actions.SHOW_SNIPPET,
		params: {
			file: file,
			startLine: startLine,
			endLine: endLine,
			defUrl: defKey,
		},
	});
};

/**
 * @description Triggers a line selection action based on the line number and the status
 * of the Shift Key paramter, which signifies whether this should mark the end of the line
 * selection or the start of i t.
 * @param {number} lineNumber - The clicked line number.
 * @param {bool} shiftKey - Whether shift was pressed or not.
 * @returns {void}
 */
module.exports.selectLines = function(lineNumber, shiftKey) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.LINE_SELECT,
		lineNumber: lineNumber,
		shiftKey: shiftKey,
	});
};

/**
 * @description Shows the discussion create form inside the code view's
 * pop-up.
 * @returns {void}
 */
module.exports.createDiscussion = function() {
	AppDispatcher.handleViewAction({
		type: globals.Actions.POPUP_CREATE_DISCUSSION,
	});
};

/**
 * @description Shows the discussion create form inside the code view's
 * pop-up.
 * @param {number} discussionId - Discussion ID.
 * @param {string} comment - Comment text body.
 * @returns {void}
 */
module.exports.submitDiscussionComment = function(discussionId, comment) {
	var defKey = TokenPopupStore.get("URL");
	AppDispatcher.dispatchAsync(CodeUtil.submitDiscussionComment(defKey, discussionId, comment), {
		started: globals.Actions.DISCUSSION_COMMENT,
		success: globals.Actions.DISCUSSION_COMMENT_SUCCESS,
		failure: globals.Actions.DISCUSSION_COMMENT_FAILED,
	});
};

/**
 * @description Submits a new discussion having title and body.
 * @param {string} title - Discussion title.
 * @param {string} body - Discussion body.
 * @returns {void}
 */
module.exports.submitDiscussion = function(title, body) {
	var defKey = TokenPopupStore.get("URL");
	AppDispatcher.dispatchAsync(CodeUtil.submitDiscussion(defKey, title, body), {
		started: globals.Actions.SUBMIT_DISCUSSION,
		success: globals.Actions.SUBMIT_DISCUSSION_SUCCESS,
		failure: globals.Actions.SUBMIT_DISCUSSION_FAILED,
	});
};

/**
 * @description Triggers the action to open the discussions with
 * the passed model for the given defKey.
 * @param {string} defKey - DefKey URL form
 * @param {Object} dsc - Discussion model.
 * @returns {void}
 */
module.exports.openDiscussion = function(defKey, dsc) {
	AppDispatcher.dispatchAsync(CodeUtil.fetchDiscussion(defKey, dsc), {
		started: globals.Actions.FETCH_DISCUSSION,
		success: globals.Actions.RECEIVED_DISCUSSION,
		failure: globals.Actions.RECEIVED_DISCUSSION_FAILED,
	});
};

/**
 * @description Triggers the action to reset the popup to its default
 * page.
 * @returns {void}
 */
module.exports.showPopupPageDefault = function() {
	AppDispatcher.handleViewAction({
		type: globals.Actions.POPUP_SHOW_DEFAULT_VIEW,
	});
};

/**
 * @description Triggers the action to change the page of the popup
 * so that it lists all discussions.
 * @param {string} defKey - DefKey URL form.
 * @returns {void}
 */
module.exports.showPopupPageDiscussionList = function(defKey) {
	AppDispatcher.dispatchAsync(CodeUtil.fetchDiscussionList(defKey), {
		started: globals.Actions.FETCH_DISCUSSIONS,
		success: globals.Actions.RECEIVED_DISCUSSIONS,
		failure: globals.Actions.RECEIVED_DISCUSSIONS_FAILED,
	});
};

/**
 * @description Triggers all requests that the pop-up depends on.
 * @param {string} defKey - DefKey URL form.
 * @returns {void}
 * @private
 */
function showDefinition(defKey) {
	AppDispatcher.dispatchAsync(CodeUtil.fetchPopup(defKey), {
		started: null,
		success: globals.Actions.RECEIVED_POPUP,
		failure: globals.Actions.RECEIVED_POPUP_FAILED,
	}).then(data => popupAsyncRequests(defKey));
}

/**
 * @description Triggers a context menu at the given token to show a list
 * of overlapping definitions to chose from (in Scala case classes).
 * @param {CodeTokenModel} token - the token model.
 * @param {Array<string>} list - a list of DefKey's in URL form to fetch.
 * @returns {void}
 * @private
 */
function showDefinitionList(token, list) {
	AppDispatcher.handleDependentAction({
		type: globals.Actions.LOAD_CONTEXT_MENU,
		token: token,
	});

	AppDispatcher.dispatchAsync(CodeUtil.fetchDefinitionList(list), {
		started: null,
		success: globals.Actions.RECEIVED_MENU_OPTIONS,
		failure: globals.Actions.RECEIVED_MENU_OPTIONS_FAILED,
	});
}

/**
 * @description Loads async components of the popup. Examples and
 * discussions.
 * @param {string} defKey - DefKey URL form.
 * @returns {void}
 * @private
 */
function popupAsyncRequests(defKey) {
	module.exports.selectExample(defKey, 1);

	AppDispatcher.dispatchAsync(CodeUtil.fetchTopDiscussions(defKey), {
		started: null,
		success: globals.Actions.RECEIVED_TOP_DISCUSSIONS,
		failure: globals.Actions.RECEIVED_TOP_DISCUSSIONS_FAILED,
	});
}

/**
 * @description Dispatches the payload after silently making sure the current file
 * is loaded first.
 * @param {object} file - Contains file information. At minimum RepoRev and Path
 * @param {object} payload - Payload to dispatch.
 * @returns {void}
 * @private
 */
function dispatchIfSameFile(file, payload) {
	var dispatch = () => AppDispatcher.handleViewAction(payload);

	if (CodeFileStore.isSameFile(file)) {
		return dispatch();
	}
	var url = file.URL;
	if (typeof url !== "string") {
		url = router.fileURL(file.RepoRev.URI, file.RepoRev.CommitID||file.RepoRev.Rev, file.Path);
	}
	dependentLoadFile(url).then(dispatch);
}

/**
 * @description This is a dependent action. It is used by other actions
 * as a dependency and should not be registered in the system.
 * @param {string} url - URL of file to lead.
 * @returns {jQuery.jqXHR} - Promise.
 */
function dependentLoadFile(url) {
	AppDispatcher.handleDependentAction({
		type: globals.Actions.FETCH_FILE,
		url: url,
	});

	return loadFile(url, null);
}

/**
 * @description Loads the file at the given URL.
 * @param {string} url - The URL of the file or definition file to load.
 * @param {globals.Actions|null} startActionType - The action type to dispatch when
 * the requests starts.
 * @returns {jQuery.jqXHR} - promise
 * @private
 */
function loadFile(url, startActionType) {
	return AppDispatcher.dispatchAsync(CodeUtil.fetchFile(url), {
		started: startActionType,
		success: globals.Actions.RECEIVED_FILE,
		failure: globals.Actions.RECEIVED_FILE_FAILED,
	}).then(handleSuccess, handleFailure);
}

/**
 * @description Validates the file payload to check if pop-up
 * information was supplied, in which case, the correct ations are
 * taken to display it.
 * @param {object} data - File data.
 * @returns {void}
 * @private
 */
function handleSuccess(data) {
	if (data.Definition) {
		AppDispatcher.handleServerAction({
			type: globals.Actions.RECEIVED_POPUP,
			data: data.Definition,
		});

		popupAsyncRequests(data.Definition.URL);
	}
	return data;
}

/**
 * @description If the file request fails for whatever reason, check
 * if we need to redirect the user or report the returned error.
 * @param {object} data - Reject reason.
 * @returns {void}
 * @private
 */
function handleFailure(data) {
	if (data.hasOwnProperty("RedirectTo")) {
		AppDispatcher.redirectTo(data.RedirectTo);
	} else {
		notify.error("Failed to load file");
		console.error(data);
	}
	return data;
}
