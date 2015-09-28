var globals = require("../globals");
var AppDispatcher = require("../dispatchers/AppDispatcher");
var CodeUtil = require("../util/CodeUtil");
var CodeFileStore = require("../stores/CodeFileStore");

/**
 * @description Selects the passed token model.
 * @param {CodeTokenModel} token - Model of token to show.
 * @returns {void}
 */
module.exports.selectToken = function(token) {
	AppDispatcher.handleRouterAction({
		type: globals.Actions.TOKEN_SELECT,
		token: token,
	});

	CodeUtil.abortPopoverXhr();

	var defKey = token.get("url")[0];
	AppDispatcher.dispatchAsync(CodeUtil.fetchPopup(defKey), {
		started: null,
		success: globals.Actions.RECEIVED_POPUP,
		failure: globals.Actions.RECEIVED_POPUP_FAILED,
	}).then(data => loadExampleAndDiscussions(defKey, data));
};

/**
 * @description Loads the file at the given URL
 * @param {string} url - URL of file to load
 * @returns {void}
 */
module.exports.loadFile = function(url) {
	AppDispatcher.handleRouterAction({
		type: globals.Actions.FETCH_FILE,
		url: url,
	});

	return CodeUtil
		.fetchFile(url)
		.then(onReceiveFile, onReceiveFileError);
};

function onReceiveFile(data) {
	AppDispatcher.handleServerAction({
		type: globals.Actions.RECEIVED_FILE,
		data: data,
	});

	if (data.Definition) {
		AppDispatcher.handleServerAction({
			type: globals.Actions.RECEIVED_POPUP,
			data: data.Definition,
		});

		loadExampleAndDiscussions(data.Definition.URL, data.Definition);
	}
}

function loadExampleAndDiscussions(defKey, data) {
	var file = CodeFileStore.get("file");
	var repoURI = (file && file.RepoRev) ? file.RepoRev.URI : null;

	AppDispatcher.dispatchAsync(CodeUtil.fetchExample(defKey, 1, repoURI), {
		started: globals.Actions.FETCH_EXAMPLE,
		success: globals.Actions.RECEIVED_EXAMPLE,
		failure: globals.Actions.RECEIVED_EXAMPLE_FAILED,
	});

	AppDispatcher.dispatchAsync(CodeUtil.fetchTopDiscussions(defKey), {
		started: null,
		success: globals.Actions.RECEIVED_TOP_DISCUSSIONS,
		failure: globals.Actions.RECEIVED_TOP_DISCUSSIONS_FAILED,
	});
}

function onReceiveFileError(data) {
	if (data.hasOwnProperty("RedirectTo")) {
		AppDispatcher.redirectTo(data.RedirectTo);
	}
}

/**
 * @description Shows the definition contained in the box info object.
 * @param {Object} def - sourcegraph.Def
 * @returns {void}
 */
module.exports.navigateToDefinition = function(def) {
	var dispatch = function() {
		AppDispatcher.handleRouterAction({
			type: globals.Actions.SHOW_DEFINITION,
			params: def,
		});
	};

	if (CodeFileStore.isSameFile(def.File)) return dispatch();
	module.exports.loadFile(def.URL);
};
