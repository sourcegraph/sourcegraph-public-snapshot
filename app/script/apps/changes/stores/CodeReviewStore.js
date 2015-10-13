var Backbone = require("backbone");
var notify = require("../../../components/notify");
var FluxStore = require("../../../stores/FluxStore");
var AppDispatcher = require("../../../dispatchers/AppDispatcher");
var CodeReviewChangesModel = require("./models/CodeReviewChangesModel");
var CommitCollection = require("../../../stores/collections/CommitCollection");
var ReviewCollection = require("./collections/ReviewCollection");

var CodeReviewStore = FluxStore({

	defaults: {
		/**
		 * @description Holds information about the differential of this changeset.
		 * @type {CodeReviewChangesModel}
		 */
		changes: new CodeReviewChangesModel(),

		/**
		 * @description Holds the collection of commits for this changeset.
		 * @type {CommitCollection}
		 */
		commits: new CommitCollection(),

		/**
		 * @description Holds a collection of reviews for the changeset.
		 * @type {ReviewCollection}
		 */
		reviews: null,

		/**
		 * @description Holds the state of the review form. If true, the form
		 * will be open.
		 * @type {bool}
		 */
		submittingReview: false,
	},

	dispatcher: AppDispatcher,

	actions: {
		CR_LOAD_DATA: "_onLoadData",
		CR_RECEIVED_CHANGES: "_onLoadData",
		CR_RECEIVED_POPOVER: "_onReceivePopover",
		CR_FOCUS_TOKEN: "_onFocusToken",
		CR_BLUR_TOKENS: "_onBlurTokens",
		CR_SELECT_TOKEN: "_onSelectToken",
		CR_DESELECT_TOKENS: "_onDeselectToken",
		CR_RECEIVED_HUNK_CONTEXT: "_onReceivedHunkContext",
		CR_SELECT_FILE: "_onSelectFile",
		CR_RECEIVED_POPUP: "_onReceivedPopup",
		CR_RECEIVED_EXAMPLE: "_onReceivedExample",
		CR_RECEIVED_CHANGED_STATUS: "_onReceivedStatusChange",
		CR_SAVE_DRAFT: "_onSaveDraft",
		CR_UPDATE_DRAFT: "_onUpdateDraft",
		CR_DELETE_DRAFT: "_onDeleteDraft",
		CR_SUBMIT_REVIEW: "_onSubmitReview",
		CR_SUBMIT_REVIEW_SUCCESS: "_onSubmitReviewSuccess",
		CR_SHOW_COMMENT: "_onShowComment",
	},

	/**
	 * @description Triggered when data is loaded into the store. It parses it
	 * and creates all the models needed to display the Changeset view.
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onLoadData(action) {
		var d = action.data;

		if (d.Commits) {
			this.get("commits").load(d.Commits);
		}
		if (d.Files) {
			this.get("changes").load(d.Files);
		}

		var reviews = new ReviewCollection((d.Reviews || []), {
			changesetId: d.Changeset.ID,
			repo: d.Changeset.DeltaSpec.Base.URI,
		});

		// TODO(gbbr): Make a separated collection
		var events = new Backbone.Collection(d.Events || []);

		this.set({
			Changeset: d.Changeset,
			Delta: d.Delta,
			BaseTip: d.BaseTip,
			FileFilter: d.FileFilter,
			ReviewGuidelines: d.ReviewGuidelines,
			loading: false,
			reviews: reviews,
			events: events,
		});
	},

	/**
	 * @description Triggered when new data is received for a popover.
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onReceivePopover(action) {
		this.get("changes").updatePopover(action.data);
	},

	/**
	 * @description Triggered when the action for focusing a token is dispatched.
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onFocusToken(action) {
		this.get("changes").focusToken(action.token, action.event, action.file);
	},

	/**
	 * @description Triggered when the action for bluring a token is dispatched.
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onBlurTokens(action) {
		this.get("changes").blurTokens(action.token, action.event, action.file);
	},

	/**
	 * @description Triggered when the action for selecting a token is dispatched.
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onSelectToken(action) {
		this.get("changes").selectToken(action.token, action.event, action.file);
	},

	/**
	 * @description Triggered when the action for deselecting a token is dispatched.
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onDeselectToken(action) {
		this.get("changes").deselectTokens();
	},

	/**
	 * @description Triggered when context for a hunk was received from the server.
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onReceivedHunkContext(action) {
		var hunk = action.model;
		var data = action.data;
		var index, parent;

		if (action.isTop) {
			hunk.updateTop(data);

			// if the hunks are "touching", merge them
			index = hunk.index();
			parent = hunk.get("Parent");
			var prevHunk = parent.get("Hunks").at(index - 1);

			if (prevHunk && prevHunk.get("NewStartLine") + prevHunk.get("NewLines") === data.Entry.StartLine) {
				parent.merge(index - 1, index);
			}
		} else {
			hunk.updateBottom(data);

			// if the hunks are "touching", merge them
			index = hunk.index();
			parent = hunk.get("Parent");
			var nextHunk = parent.get("Hunks").at(index + 1);

			if (nextHunk && nextHunk.get("NewStartLine") === data.Entry.EndLine + 1) {
				parent.merge(index, index + 1);
			}
		}
	},

	/**
	 * @description Triggered when a file is selected in the list of the differential.
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onSelectFile(action) {
		this.get("changes").scrollTo(action.file);
	},

	/**
	 * @description Triggered when data for a popup is received from the server.
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onReceivedPopup(action) {
		this.get("changes").showPopup(action.data);
	},

	/**
	 * @description Triggered when data from the server is received as a follow
	 * up for a status change request (Open, Close, etc)
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onReceivedStatusChange(action) {
		if (action.data.hasOwnProperty("Op")) {
			this.get("events").add(action.data, {silent: true});
			this.set("Changeset", action.data.After);
			notify.info("Changeset status updated");
		}
	},

	/**
	 * @description Triggered when a usage example is received from the server.
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onReceivedExample(action) {
		this.get("changes").showExample(action.data, action.page);
	},

	/**
	 * @description Triggered when a draft has been added.
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onSaveDraft(action) {
		this.get("reviews").addDraft(action.draft);
		action.hunk.closeComment(action.line); // triggers change in hunk model
		action.fileDiff.trigger("change");
	},

	/**
	 * @description Triggered when a draft has been edited.
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onUpdateDraft(action) {
		this.get("reviews").updateDraft(action.comment, action.newBody);
	},

	/**
	 * @description Triggered when a draft has been deleted.
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onDeleteDraft(action) {
		this.get("reviews").deleteDraft(action.comment);
		action.hunk.trigger("change");
	},

	/**
	 * @description Triggered when the user initiates submitting a review.
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onSubmitReview(action) {
		// noop
	},

	/**
	 * @description Triggered when the server has confirmed that the review
	 * has been successfully submitted.
	 * @param {Object} action - The action's payload.
	 * @returns {void}
	 * @private
	 */
	_onSubmitReviewSuccess(action) {
		this.get("reviews").clearDrafts();
		this.get("reviews").add(action.data);
		this.set("submittingReview", false);
	},

	_onShowComment(action) {
		this.trigger("scrollTop", action.comment.getAbsolutePosition().top);
	},
});

module.exports = CodeReviewStore;
