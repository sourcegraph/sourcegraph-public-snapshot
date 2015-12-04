var keyMirror = require("keymirror");

module.exports = {
	// Token types
	TokenType: keyMirror({
		STRING: null,
		SPAN: null,
		REF: null,
		DEF: null,
	}),

	// Number of lines to expand on hunk
	HunkExpandLines: 20,

	Actions: keyMirror({
		// CodeFile Actions
		FETCH_FILE: null,
		RECEIVED_FILE: null,
		RECEIVED_FILE_FAILED: null,
		TOKEN_SELECT: null,
		RECEIVED_POPUP: null,
		RECEIVED_POPUP_FAILED: null,
		FETCH_EXAMPLE: null,
		RECEIVED_EXAMPLE: null,
		RECEIVED_EXAMPLE_FAILED: null,
		TOKEN_FOCUS: null,
		RECEIVED_POPOVER: null,
		RECEIVED_POPOVER_FAILED: null,
		TOKEN_BLUR: null,
		TOKEN_CLEAR: null,
		LINE_SELECT: null,
		SHOW_SNIPPET: null,
		SHOW_DEFINITION: null,
		SWITCH_POPUP_DEFINITION: null,
		REDIRECT: null,
		LOAD_CONTEXT_MENU: null,
		RECEIVED_MENU_OPTIONS: null,
		RECEIVED_MENU_OPTIONS_FAILED: null,
		CODE_FILE_CLICK: null,
		POPUP_SHOW_DEFAULT_VIEW: null,

		POPUP_FETCH_PAGE: null,
		POPUP_SHOW_PAGE: null,
		POPUP_SHOW_PAGE_FAILED: null,

		// Compare View Actions
		DIFF_LOAD_DATA: null,
		DIFF_FOCUS_TOKEN: null,
		DIFF_BLUR_TOKENS: null,
		DIFF_RECEIVED_POPOVER: null,
		DIFF_RECEIVED_POPUP: null,
		DIFF_SELECT_TOKEN: null,
		DIFF_DESELECT_TOKENS: null,
		DIFF_FETCH_EXAMPLE: null,
		DIFF_RECEIVED_EXAMPLE: null,
		DIFF_SELECT_FILE: null,
		DIFF_RECEIVED_HUNK_TOP: null,
		DIFF_RECEIVED_HUNK_BOTTOM: null,
		DIFF_PROPOSE_CHANGE: null,
		DIFF_PROPOSE_CHANGE_SUCCESS: null,

		// Code Review Actions
		CR_LOAD_DATA: null,
		CR_RECEIVED_CHANGES: null,
		CR_FOCUS_TOKEN: null,
		CR_BLUR_TOKENS: null,
		CR_SELECT_TOKEN: null,
		CR_DESELECT_TOKENS: null,
		CR_SELECT_FILE: null,
		CR_RECEIVED_POPOVER: null,
		CR_RECEIVED_HUNK_CONTEXT: null,
		CR_RECEIVED_POPUP: null,
		CR_RECEIVED_EXAMPLE: null,
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
	}),

	ChangesetStatus: keyMirror({
		CLOSED: null,
		OPEN: null,
		MERGED: null,
	}),

	PopupPages: keyMirror({
		DEFAULT: null,
	}),

	Features: typeof window !== "undefined" && window._featureToggles,

	CsrfToken: typeof window !== "undefined" && window._csrfToken,
};
