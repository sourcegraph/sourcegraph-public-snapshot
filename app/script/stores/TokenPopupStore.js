var FluxStore = require("./FluxStore");
var globals = require("../globals");

var AppDispatcher = require("../dispatchers/AppDispatcher");
var ExamplesModel = require("./models/ExamplesModel");
var DiscussionModel = require("../stores/models/DiscussionModel");
var DiscussionCollection = require("../stores/collections/DiscussionCollection");

module.exports = FluxStore({

	defaults: {
		/**
		 * @description closed, if true, will close a visible pop-up.
		 * @type {bool}
		 */
		closed: true,

		/**
		* @description type holds the type of reference shown in the pop-up.
		* Allows the view to tell if the 'Go to definition'
		* button should be presented to the user.
		* @type {globals.TokenType}
		*/
		type: globals.TokenType.DEF,

		/**
		 * @description URL is the DefKey of the shown definition, in URL form.
		 * @type {string}
		 */
		URL: null,

		/**
		 * @description examplesModel holds the model of the example that needs
		 * to be presented.
		 * @type {ExamplesModel}
		 */
		examplesModel: new ExamplesModel(),

		/**
		 * @description extraDefinitions holds an array of definitions that overlap
		 * with this one (happens in Scala with 'case classes').
		 * @type {Array<object>}
		 */
		extraDefinitions: [],

		/**
		 * @description topDiscussions holds the data for the discussions snippet.
		 * These discussions are ordered by rating.
		 * @type {DiscussionCollection}
		 */
		topDiscussions: new DiscussionCollection(),

		/**
		 * @description page is the current page visible in the pop-up. It has a 'type'
		 * property which holds the page type and an optional 'data' property which holds
		 * data specific to that page.
		 * @type {globals.PopupPages}
		 */
		page: {type: globals.PopupPages.DEFAULT},
	},

	dispatcher: AppDispatcher,

	actions: {
		TOKEN_SELECT: "_onTokenSelect",
		SWITCH_POPUP_DEFINITION: "_onSwitchDefinition",
		RECEIVED_POPUP: "_onReceivedPopup",
		FETCH_EXAMPLE: "_onFetchExample",
		RECEIVED_EXAMPLE: "_onReceivedExample",
		TOKEN_FOCUS: "_onTokenFocus",
		TOKEN_CLEAR: "_onTokenClear",
		TOKEN_BLUR: "_onTokenBlur",
		SHOW_SNIPPET: "_onShowDefinition",
		RECEIVED_DISCUSSION: "_onReceivedDiscussion",
		SUBMIT_DISCUSSION_SUCCESS: "_onCreatedDiscussion",
		RECEIVED_TOP_DISCUSSIONS: "_onReceivedTopDiscussions",
		RECEIVED_DISCUSSIONS: "_onReceivedDiscussionList",
		POPUP_CREATE_DISCUSSION: "_onPageDiscussionCreate",
		POPUP_SHOW_DEFAULT_VIEW: "_onPageDefault",
		DISCUSSION_COMMENT_SUCCESS: "_onDiscussionComment",
	},

	/**
	 * @description Sets the page to the default (token) view.
	 * @returns {void}
	 * @private
	 */
	_onPageDefault() {
		this._setPage(globals.PopupPages.DEFAULT);
	},

	/**
	 * @description Sets the page to the "Create Discussion" form.
	 * @param {object} action - Contains data & source.
	 * @returns {void}
	 * @private
	 */
	_onPageDiscussionCreate(action) {
		this._setPage(globals.PopupPages.NEW_DISCUSSION);
	},

	/**
	 * @description Sets the page to view the received discussion. This is
	 * triggered both when viewing an existing discussion or after creating
	 * a new one.
	 * @param {object} action - Payload data including discussion.
	 * @returns {void}
	 * @private
	 */
	_onReceivedDiscussion(action) {
		var model = new DiscussionModel(action.data);
		this._setPage(globals.PopupPages.VIEW_DISCUSSION, model);
	},

	/**
	 * @description Puts the newly created discussion in view and adds it to
	 * the snippet (if the snippet has less than 4 discussions).
	 * @param {object} action - Payload data including discussion.
	 * @returns {void}
	 * @private
	 */
	_onCreatedDiscussion(action) {
		var top = this.get("topDiscussions");
		if (top.length < globals.DiscussionSnippetEntries) {
			top.add(action.data);
		}

		this._onReceivedDiscussion(action);
	},

	/**
	 * @description Sets the page to view the received discussion list.
	 * @param {object} action - Contains data & source.
	 * @returns {void}
	 * @private
	 */
	_onReceivedDiscussionList(action) {
		var collection = new DiscussionCollection(action.data.Discussions);
		this._setPage(globals.PopupPages.LIST_DISCUSSION, collection);
	},

	/**
	 * @description Sets the page to the given page 'type' and passes
	 * it the data. Valid page types can be seen in globals.PopupPages.
	 * Each type expects its own data form.
	 * @param {globals.PopupPages} type - Page type
	 * @param {object} data - Any data to pass to the page.
	 * @returns {void}
	 * @private
	 */
	_setPage(type, data) {
		if (globals.PopupPages.hasOwnProperty(type)) {
			this.set("page", {
				type: type,
				data: data,
			});
		}
	},

	/**
	 * @description Updates the top discussions to the ones received.
	 * @param {object} action - Contains data & source.
	 * @returns {void}
	 * @private
	 */
	_onReceivedTopDiscussions(action) {
		this.get("topDiscussions").set(action.data.Discussions);
	},

	/**
	 * @description Called after a discussion was successfully created. Updates
	 * the model and triggers change.
	 * @param {object} action - Triggered action and data.
	 * @returns {void}
	 * @private
	 */
	_onDiscussionComment(action) {
		var pg = this.get("page");
		if (pg.data instanceof DiscussionModel) {
			pg.data.addComment(action.data);
		}
	},

	/**
	 * @description Triggered when starting to fetch example.
	 * @returns {void}
	 * @private
	 */
	_onFetchExample() {
		this.get("examplesModel").set("loading", true);
	},

	/**
	 * @description Triggered when an example has been received.
	 * @param {object} action - Contains data & source.
	 * @returns {void}
	 * @private
	 */
	_onReceivedExample(action) {
		var data = action.data;
		var em = this.get("examplesModel");
		em.set("loading", false);
		em.showExample(data.example, data.page);
	},

	/**
	 * @description This method indicates that a token was focused.
	 * @param {object} action - Contains data & source.
	 * @returns {void}
	 * @private
	 */
	_onTokenFocus(action) {
		var cm = this.get("examplesModel").get("codeModel");
		if (!cm.tokens) return;
		cm.highlightToken(action.token.get("url")[0]);
	},

	/**
	 * @description Triggered when tokens are de-selected.
	 * @returns {void}
	 * @private
	 */
	_onTokenClear() {
		this.get("examplesModel").set({example: undefined});
	},

	/**
	 * @description Triggered when a token receives the mouse-out / blur event.
	 * @returns {void}
	 * @private
	 */
	_onTokenBlur() {
		var cm = this.get("examplesModel").get("codeModel");
		if (!cm.tokens) return;
		cm.clearHighlightedTokens();
	},

	/**
	 * @description Closes the popup and updates its data to the token contained in the payload.
	 * The popup reopens when the "RECEIVED_POPUP" action type is dispatched.
	 * @param {Object} action - Action payload.
	 * @returns {void}
	 * @private
	 */
	_onTokenSelect(action) {
		this.get("examplesModel").set("example", undefined);
		this.set({
			closed: true,
			extraDefinitions: [],
			type: action.token.get("type"),
			page: {type: globals.PopupPages.DEFAULT},
		});
	},

	/**
	 * @description Sets the popup to definition mode.
	 * @returns {void}
	 * @private
	 */
	_onShowDefinition() {
		this.set({type: globals.TokenType.DEF});
	},

	/**
	 * @description Prepares the popup for displaying another definition.
	 * @param {object} action - Contains data & source.
	 * @returns {void}
	 * @private
	 */
	_onSwitchDefinition(action) {
		this.set("closed", true);
	},

	/*
	 * @description Sets the popup into loading mode to signal that the request
	 * for additional definitions is in progress.
	 * @private
	 */
	_onFetchAdditionalDefs() {
		this.set("loading", true);
	},

	/*
	 * @description Stops the loading and registers the new received definitions
	 * with the store.
	 * @param {Object} action - action payload object
	 * @private
	 */
	_onReceivedAdditionalDefs(action) {
		this.set("loading", false);
		this.set("extraDefinitions", action.data);
	},

	/**
	 * @description Updates the model to the passed data, setting closed and error to false.
	 * @param {Object} action - Action payload.
	 * @returns {void}
	 * @private
	 */
	_onReceivedPopup(action) {
		this.set(action.data, {silent: true});
		this.set({
			error: false,
			closed: false,
		});
	},
});
