var keyMirror = require("keymirror");

module.exports = {
	// Number of lines to expand on hunk
	HunkExpandLines: 20,

	Actions: keyMirror({
		// CodeFile Actions
		REDIRECT: null,

		// Compare View Actions
		DIFF_LOAD_DATA: null,
		DIFF_SELECT_FILE: null,
		DIFF_RECEIVED_HUNK_TOP: null,
		DIFF_RECEIVED_HUNK_BOTTOM: null,
		DIFF_PROPOSE_CHANGE: null,
		DIFF_PROPOSE_CHANGE_SUCCESS: null,

		// Code Review Actions
		CR_LOAD_DATA: null,
		CR_RECEIVED_CHANGES: null,
		CR_SELECT_FILE: null,
		CR_RECEIVED_HUNK_CONTEXT: null,
		CR_SAVE_DRAFT: null,
		CR_UPDATE_DRAFT: null,
		CR_UPDATE_TITLE: null,
		CR_DELETE_DRAFT: null,
		CR_SUBMIT_REVIEW: null,
		CR_RECEIVED_CHANGED_STATUS: null,
		CR_SUBMIT_REVIEW_SUCCESS: null,
		CR_SUBMIT_REVIEW_FAIL: null,
		CR_SHOW_COMMENT: null,
		CR_MERGE: null,
		CR_MERGE_SUCCESS: null,
		CR_MERGE_FAIL: null,
		CR_LGTM_CHANGE_SUCCESS: null,
		CR_LGTM_CHANGE_FAIL: null,
		CR_SUBMIT_DESCRIPTION_SUCCESS: null,
		CR_SUBMIT_DESCRIPTION_FAIL: null,
		CR_ADD_REVIEWER_SUCCESS: null,
		CR_ADD_REVIEWER_FAIL: null,
		CR_REMOVE_REVIEWER_SUCCESS: null,
		CR_REMOVE_REVIEWER_FAIL: null,
	}),

	ChangesetStatus: keyMirror({
		CLOSED: null,
		OPEN: null,
		MERGED: null,
	}),

	Features: typeof window !== "undefined" && window._featureToggles,

	CsrfToken: typeof window !== "undefined" && window._csrfToken,
};
