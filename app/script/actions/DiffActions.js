var AppDispatcher = require("../dispatchers/AppDispatcher");
var DiffServerActions = require("./DiffServerActions");
var DiffStore = require("../stores/DiffStore");
var CodeUtil = require("../util/CodeUtil");
var globals = require("../globals");
var router = require("../routing/router");

/**
 * @description Action creator discarded when the page loads with pre-loaded data attached
 * from the server.
 * @param {Object} data - Raw DeltaFiles data received from server.
 * @returns {void}
 */
module.exports.loadData = function(data) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.DIFF_LOAD_DATA,
		data: data,
	});
};

/**
 * @description Action creator discarded when the user requests context above a hunk.
 * @param {HunkModel} hunk - The model of the hunk to give context to.
 * @returns {void}
 */
module.exports.expandHunkUp = function(hunk) {
	// TODO(slimsag): severe duplication with CodeReviewActions.js : expandHunk
	var endLine = hunk.get("NewStartLine") - 1,
		startLine = endLine - globals.HunkExpandLines < 1 ? 1 : endLine - globals.HunkExpandLines,
		repoRev = DiffStore.get("RepoRevSpec"),
		fileDiff = hunk.get("Parent"),
		deltaSpec = DiffStore.get("DeltaSpec"),
		url = router.fileURL(repoRev.URI, deltaSpec.Head.Rev, fileDiff.get("NewName")),
		hunks = fileDiff.get("Hunks");

	// don't overflow into previous hunk
	var index = hunk.index();
	if (index > 0 && hunks.length > 1) {
		var prevHunk = hunks.at(index - 1);
		if (prevHunk && prevHunk.get("NewStartLine") + prevHunk.get("NewLines") - 1 >= startLine) {
			startLine = prevHunk.get("NewStartLine") + prevHunk.get("NewLines");
		}
	}
	endLine -= 1;

	CodeUtil.fetchFile(url, startLine, endLine).then(
		data => DiffServerActions.receivedHunkTop(hunk, data),
		DiffServerActions.failedReceiveExpansion
	);
	return {startLine: startLine, endLine: endLine};
};

/**
 * @description Action creator discarded when the user requests context below a hunk.
 * @param {HunkModel} hunk - The model of the hunk to give context to.
 * @returns {void}
 */
module.exports.expandHunkDown = function(hunk) {
	// TODO(slimsag): severe duplication with CodeReviewActions.js : expandHunk
	var firstLine = hunk.get("NewStartLine"),
		newLines = hunk.get("NewLines");

	var startLine = firstLine + newLines,
		endLine = startLine + globals.HunkExpandLines,
		repoRev = DiffStore.get("RepoRevSpec"),
		deltaSpec = DiffStore.get("DeltaSpec"),
		fileDiff = hunk.get("Parent"),
		url = router.fileURL(repoRev.URI, deltaSpec.Head.Rev, fileDiff.get("NewName"));

	// don't overflow into next hunk
	var index = hunk.index(),
		nextHunk = fileDiff.get("Hunks").at(index + 1);

	if (nextHunk && nextHunk.get("NewStartLine") <= endLine) {
		endLine = nextHunk.get("NewStartLine") - 1;
	}

	CodeUtil.fetchFile(url, startLine, endLine).then(
		data => DiffServerActions.receivedHunkBottom(hunk, data),
		DiffServerActions.failedReceiveExpansion
	);
};

/**
 * @description Starts the request to create a new changeset.
 * @param {string} repo - URL of repository the changeset belongs to.
 * @param {Object} changeSet - Changeset properties (sourcegraph.Changeset)
 * @returns {void}
 */
module.exports.proposeChange = function(repo, changeSet) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.DIFF_PROPOSE_CHANGE,
	});

	CodeUtil
		.createChangeset(repo, changeSet)
		.then(
			DiffServerActions.receivedChangesetCreate,
			DiffServerActions.receivedChangesetCreateFailed
		);
};

/**
 * @description Action creator discarded when the user selects a file in the list.
 * @param {FileDiffModel} fileDiff - Model of the selected diff.
 * @param {Event} evt - Browser event. Generally click.
 * @returns {void}
 */
module.exports.selectFile = function(fileDiff, evt) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.DIFF_SELECT_FILE,
		file: fileDiff,
	});

	evt.preventDefault();
};

/**
 * @description Action creator discarded when the user focuses a token in the compare view.
 * @param {CodeTokenModel} token - Focused token.
 * @param {Event} evt - Browser event. Generally click.
 * @param {FileDiffModel} fileDiff - Model of the containing diff.
 * @returns {void}
 */
module.exports.focusToken = function(token, evt, fileDiff) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.DIFF_FOCUS_TOKEN,
		file: fileDiff,
		token: token,
		event: evt,
	});

	CodeUtil
		.fetchPopover(token.get("url")[0])
		.then(
			DiffServerActions.receivedPopover,
			DiffServerActions.receivedPopoverFailed
		);
};

/**
 * @description Action creator discarded when the user selects (clicks) a token in the compare view.
 * @param {CodeTokenModel} token - Focused token.
 * @param {Event} evt - Browser event. Generally click.
 * @returns {void}
 */
module.exports.selectToken = function(token, evt) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.DIFF_SELECT_TOKEN,
		token: token,
		event: evt,
	});

	var url = token.get("url")[0];

	CodeUtil
		.fetchPopup(url)
		.then(
			DiffServerActions.receivedPopup,
			DiffServerActions.receivedPopupFailed
		);

	CodeUtil
		.fetchExample(url, 1)
		.then(
			DiffServerActions.receivedExample,
			DiffServerActions.receivedExampleFailed
		);
};

/**
 * @description Action creator discarded when the user requests a new example.
 * @param {string} url - URL of the definition to fetch the example for.
 * @param {number} page - The index of the example to fetch.
 * @returns {void}
 */
module.exports.selectExample = function(url, page) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.DIFF_FETCH_EXAMPLE,
		params: page,
	});

	CodeUtil
		.fetchExample(url, page)
		.then(
			DiffServerActions.receivedExample,
			DiffServerActions.receivedExampleFailed
		);
};

/**
 * @description Action discarded when the popup is closed.
 * @returns {void}
 */
module.exports.closePopup = function() {
	AppDispatcher.handleViewAction({
		type: globals.Actions.DIFF_DESELECT_TOKENS,
	});
};

/**
 * @description Action creator discarded when tokens are blurred in a file diff.
 * @param {FileDiffModel} fileDiff - Model of the containing diff.
 * @returns {void}
 */
module.exports.blurTokens = function(fileDiff) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.DIFF_BLUR_TOKENS,
		file: fileDiff,
	});

	CodeUtil.abortPopoverXhr();
};
