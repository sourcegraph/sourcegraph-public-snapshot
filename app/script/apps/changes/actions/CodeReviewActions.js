var globals = require("../../../globals");
var AppDispatcher = require("../../../dispatchers/AppDispatcher");
var CodeReviewStore = require("../stores/CodeReviewStore");
var CodeUtil = require("../../../util/CodeUtil");
var CodeReviewServerActions = require("./CodeReviewServerActions");
var router = require("../../../routing/router");
var CurrentUser = require("../../../CurrentUser");

/**
 * @description Action creator discarded when the page loads with pre-loaded data attached
 * from the server.
 * @param {Object} data - Raw code review data received from server.
 * @returns {void}
 */
module.exports.loadData = function(data) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.CR_LOAD_DATA,
		data: data,
	});
};

/**
 * @description Dispatches the action that a token is focused. It also initiates
 * requests to fetch data for the popover.
 * @param {CodeTokenModel} token - The token that was hovered.
 * @param {Event} evt - The (mouseover) event.
 * @param {FileDiffModel} fileDiff - The FileDiff that the token is part of.
 * @returns {void}
 */
module.exports.focusToken = function focusToken(token, evt, fileDiff) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.CR_FOCUS_TOKEN,
		file: fileDiff,
		token: token,
		event: evt,
	});

	CodeUtil
		.fetchPopover(token.get("url")[0])
		.then(
			CodeReviewServerActions.receivedPopover,
			CodeReviewServerActions.receivedPopoverFailed
		);
};

/**
 * @description Dispatches the action that a request to change the title has been
 * submitted. It also initiates a request to persist the data to the server.
 * @param {Object} changeset - sourcegraph.Changeset
 * @param {string} newTitle - The new title
 * @returns {void}
 */
module.exports.submitTitle = function(changeset, newTitle) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.CR_UPDATE_TITLE,
	});

	CodeUtil
		.updateChangesetStatus(changeset.DeltaSpec.Base.URI, changeset.ID, {Title: newTitle})
		.then(
			CodeReviewServerActions.statusUpdated,
			CodeReviewServerActions.statusUpdateFailed
		);
};

/**
 * @description Aborts any exisiting popover request and dispatches the token
 * blur action.
 * @param {CodeTokenModel} token - The token that was hovered.
 * @param {Event} evt - The (mouseover) event.
 * @param {FileDiffModel} fileDiff - The FileDiff that the token is part of.
 * @returns {void}
 */
module.exports.blurTokens = function blurTokens(token, evt, fileDiff) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.CR_BLUR_TOKENS,
		file: fileDiff,
	});

	CodeUtil.abortPopoverXhr();
};

/**
 * @description Dispatches an action notifying that a review is being submitted
 * and initiates a request to submit that review.
 * @param {string} body - The text body of the review.
 * @returns {void}
 */
module.exports.submitReview = function submitReview(body) {
	var changeset = CodeReviewStore.get("Changeset"),
		repo = changeset.DeltaSpec.Base.URI,
		drafts = CodeReviewStore.get("reviews").drafts.toJSON();

	AppDispatcher.handleViewAction({
		type: globals.Actions.CR_SUBMIT_REVIEW,
		body: body,
	});

	CodeUtil
		.submitReview(repo, changeset.ID, body, drafts)
		.then(
			CodeReviewServerActions.submitReviewSuccess,
			CodeReviewServerActions.submitReviewFail
		);
};

/**
 * @description Initiates a request for additional context to a given hunk.
 * @param {HunkModel} hunk - The hunk to get context for.
 * @param {bool} isDirectionUp - Whether to receive context on the upper side
 * of the hunk or the lower side.
 * @param {Event} evt - The (click) event.
 * @returns {void}
 */
module.exports.expandHunk = function expandHunk(hunk, isDirectionUp, evt) {
	var delta = CodeReviewStore.get("Delta");
	var fileDiff = hunk.get("Parent");
	var url = router.fileURL(delta.HeadRepo.URI, delta.Head.Rev, fileDiff.get("NewName"));
	var index = hunk.index();
	var startLine, endLine;
	var hunks = fileDiff.get("Hunks");

	if (isDirectionUp) {
		endLine = hunk.get("NewStartLine") - 1;
		startLine = endLine - globals.HunkExpandLines < 1 ? 1 : endLine - globals.HunkExpandLines;

		// don't overflow into previous hunk
		if (hunks.length > 1) {
			var prevHunk = hunks.at(index - 1);
			if (prevHunk && prevHunk.get("NewStartLine") + prevHunk.get("NewLines") - 1 >= startLine) {
				startLine = prevHunk.get("NewStartLine") + prevHunk.get("NewLines");
			}
		}
	} else {
		startLine = hunk.get("NewStartLine") + hunk.get("NewLines");
		endLine = startLine + globals.HunkExpandLines;

		// don't overflow into previous hunk
		if (index < hunks.length-1) {
			var nextHunk = hunks.at(index + 1);
			if (nextHunk && nextHunk.get("NewStartLine") <= endLine) {
				endLine = nextHunk.get("NewStartLine") - 1;
			}
		}
	}

	CodeUtil.fetchFile(url, startLine, endLine).then(
		data => CodeReviewServerActions.receivedHunkContext(hunk, isDirectionUp, data),
		CodeReviewServerActions.receivedHunkContextFailed
	);
};

/**
 * @description Dispatches the action that a token was selected in the view. It
 * also initiates the requests for receiving popup information and examples.
 * @param {CodeTokenModel} token - The token that was hovered.
 * @param {Event} evt - The (mouseover) event.
 * @param {FileDiffModel} fileDiff - The FileDiff that the token is part of.
 * @returns {void}
 */
module.exports.selectToken = function selectToken(token, evt, fileDiff) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.CR_SELECT_TOKEN,
		token: token,
		event: evt,
	});

	var url = token.get("url")[0];

	CodeUtil
		.fetchPopup(url)
		.then(
			CodeReviewServerActions.receivedPopup,
			CodeReviewServerActions.receivedPopupFailed
		);

	CodeUtil
		.fetchExample(url, 1)
		.then(
			CodeReviewServerActions.receivedExample,
			CodeReviewServerActions.receivedExampleFailed
		);
};

/**
 * @description Dispatches the action that an example was requested. It also
 * initiates the requests for receiving this example.
 * @param {string} url - DefKey of the definition to fetch the example for.
 * @param {number} page - The page number offset for the example.
 * @returns {void}
 */
module.exports.selectExample = function selectExample(url, page) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.DIFF_FETCH_EXAMPLE,
		params: page,
	});

	CodeUtil
		.fetchExample(url, page)
		.then(
			CodeReviewServerActions.receivedExample,
			CodeReviewServerActions.receivedExampleFailed
		);
};

/**
 * @description Action called when a user selects a file in the differential.
 * Based on whether the differential is over threshold (surpressed) or not,
 * the browser is either redirect to a single-view of the file or scrolled down
 * to bring the file into view.
 * @param {FileDiffModel} fd - The file to be navigated to.
 * @param {Event} evt - The (click) event.
 * @returns {void}
 */
module.exports.selectFile = function selectFile(fd, evt) {
	var overThreshold = CodeReviewStore.get("changes").get("overThreshold");

	if (!overThreshold) {
		evt.preventDefault();
		AppDispatcher.handleViewAction({
			type: globals.Actions.CR_SELECT_FILE,
			file: fd,
		});
	}
};

/**
 * @description Changes the status of a changeset. It dispatches the action that
 * a status change was requested and initiates the request to perform this operation.
 * @param {globals.ChangesetStatus} status - The new status.
 * @param {Event} evt - The (click) event.
 * @returns {void}
 */
module.exports.changeChangesetStatus = function(status, evt) {
	if (CurrentUser === null) {
		window.location = "/login";
		return;
	}

	AppDispatcher.handleViewAction({
		type: globals.Actions.CR_CHANGE_STATUS,
		status: status,
	});

	var id = CodeReviewStore.get("Changeset").ID,
		repo = CodeReviewStore.get("Changeset").DeltaSpec.Base.URI,
		updateOp = {};

	switch (status) {
	case globals.ChangesetStatus.OPEN:
		updateOp = {Open: true};
		break;

	case globals.ChangesetStatus.CLOSED:
		updateOp = {Close: true};
		break;

	case globals.ChangesetStatus.MERGED:
		updateOp = {Merge: true};
		break;

	default:
		CodeReviewServerActions.statusUpdateFailed("Invalid op");
		return;
	}

	CodeUtil
		.updateChangesetStatus(repo, id, updateOp)
		.then(
			CodeReviewServerActions.statusUpdated,
			CodeReviewServerActions.statusUpdateFailed
		);
};

module.exports.mergeChangeset = function(opt, evt) {
	var id = CodeReviewStore.get("Changeset").ID,
		repo = CodeReviewStore.get("Changeset").DeltaSpec.Base.URI;

	AppDispatcher.handleViewAction({
		type: globals.Actions.CR_MERGE,
	});

	CodeUtil
		.mergeChangeset(repo, id, opt)
		.then(
			CodeReviewServerActions.mergeSuccess,
			CodeReviewServerActions.mergeFailed
		);
};

/**
 * @description Dispatches the action that the user requested the closing of the
 * popup
 * @returns {void}
 */
module.exports.closePopup = function() {
	AppDispatcher.handleViewAction({
		type: globals.Actions.CR_DESELECT_TOKENS,
	});
};

/**
 * @description Dispatches the action and requests for updating a comment.
 * @param {FileDiffModel} fd - The file diff where this comment is located.
 * @param {HunkModel} hunk - The hunk where the comment was edited.
 * @param {CodeLineModel} line - The line where the comment was edit.
 * @param {Backbone.Model} comment - The comment model (with the old body).
 * @param {string} newBody - The new text body of the comment.
 * @param {Event} evt - The (click) event that triggered this action.
 * @returns {void}
 */
module.exports.updateDraft = function(fd, hunk, line, comment, newBody, evt) {
	if (comment.isDraft()) {
		AppDispatcher.handleViewAction({
			type: globals.Actions.CR_UPDATE_DRAFT,
			comment: comment,
			newBody: newBody,
		});
	}
};

/**
 * @description Dispatches the action and requests for a comment to be deleted.
 * @param {FileDiffModel} fd - The file diff where this comment is located.
 * @param {HunkModel} hunk - The hunk where the comment was edited.
 * @param {CodeLineModel} line - The line where the comment was edit.
 * @param {Backbone.Model} comment - The comment model (with the old body).
 * @param {Event} evt - The (click) event that triggered this action.
 * @returns {void}
 */
module.exports.deleteDraft = function(fd, hunk, line, comment, evt) {
	if (comment.isDraft() && confirm("Are you sure you want to delete this draft?")) {
		AppDispatcher.handleViewAction({
			type: globals.Actions.CR_DELETE_DRAFT,
			comment: comment,
			hunk: hunk,
		});
	}
};

/**
 * @description Dispatches an action that indicates that the user has requested
 * to save a draft.
 * @param {FileDiffModel} fd - The file diff where this draft will be.
 * @param {HunkModel} hunk - The hunk where this draft is placed.
 * @param {CodeLineModel} line - The line where the comment was placed.
 * @param {string} body - The text body of the comment.
 * @param {Event} evt - The (click) event when the draft was submitted.
 * @returns {void}
 */
module.exports.saveDraft = function(fd, hunk, line, body, evt) {
	var comment = {
		repo: CodeReviewStore.get("Changeset").DeltaSpec.Base.URI,
		changesetId: CodeReviewStore.get("Changeset").ID,
		User: null,
		Body: body,
		Filename: line.get("lineNumberHead") ? fd.getHeadFilename() : fd.getBaseFilename(),
		CommitID: line.get("lineNumberHead") ? fd.get("PostImage") : fd.get("PreImage"),
		LineNumber: line.get("lineNumberHead") || line.get("lineNumberBase"),
		CreatedAt: new Date(),
		Draft: true,
	};

	AppDispatcher.handleViewAction({
		type: globals.Actions.CR_SAVE_DRAFT,
		draft: comment,
		hunk: hunk,
		fileDiff: fd,
		line: line,
	});
};

/**
 * @description Triggers an action which requests that the passed comment model
 * to be displayed in its context.
 * @param {CommentModel} comment - The model of the comment to show.
 * @param {Event} event - The event that triggered the action.
 * @returns {void}
 */
module.exports.showComment = function(comment, event) {
	AppDispatcher.handleViewAction({
		type: globals.Actions.CR_SHOW_COMMENT,
		comment: comment,
	});
};
