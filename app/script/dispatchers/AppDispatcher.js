var Dispatcher = require("flux").Dispatcher;
var $ = require("jquery");
var globals = require("../globals");

var AppDispatcher = new Dispatcher();

/**
 * @description Dispatches the passed action having the View as a source.
 * This is generally a user action.
 * @param {Object} action - Action content.
 * @returns {void}
 */
AppDispatcher.handleViewAction = function(action) {
	this.dispatch({
		source: "VIEW_ACTION",
		action: action,
	});
};

/**
 * @description Dispatches the given action, having the server as a source.
 * This is generally used when asynchronous requests towards the server end.
 * @param {Object} action - Action content.
 * @returns {void}
 */
AppDispatcher.handleServerAction = function(action) {
	this.dispatch({
		source: "SERVER_ACTION",
		action: action,
	});
};

/**
 * @description Dispatches a dependent action. This is used when actions creators
 * need to create additional actions in order to complete their task.
 * @param {Object} action - Action content.
 * @returns {void}
 */
AppDispatcher.handleDependentAction = function(action) {
	this.dispatch({
		source: "DEPENDENT_ACTION",
		action: action,
	});
};

/**
 * @description Dispatches router actions. Used when the router needs to create
 * a state as a result of history navigation.
 * @param {Object} action - Action content.
 * @returns {void}
 */
AppDispatcher.handleRouterAction = function(action) {
	this.dispatch({
		source: "ROUTER_ACTION",
		action: action,
	});
};

/**
 * @description Dispatches a redirect action type with the URL attached.
 * @param {string} url - URL to redirect to.
 * @returns {void}
 */
AppDispatcher.redirectTo = function(url) {
	this.handleServerAction({
		type: globals.Actions.REDIRECT,
		url: url,
	});
};

/**
 * @description Dispatches given action types based on the outcome of
 * running the passed promise. The 'types' parameter must have 3 keys,
 * one of them being optional. Example 'types':
 * 	{
 * 		started: null | ActionType when promise starts,
 * 		success: ActionType if promise succeeds,
 * 		failure: ActionType if promise fails,
 * 	}
 * The method also returns the promise for chainability.
 * @param {jQuery.Deferred} promise - Promise to run.
 * @param {object} types - Action types.
 * @returns {jQuery.Deferred} - The original promise.
 */
AppDispatcher.dispatchAsync = function(promise, types) {
	if (!typesValid(types)) {
		return $.Deferred().reject("No action types supplied for this action.");
	}
	if (types.started) {
		this.handleViewAction({type: types.started});
	}
	return promise.then(
		y => {
			this.handleServerAction({type: types.success, data: y});
			return $.Deferred().resolve(y);
		},
		n => {
			this.handleServerAction({type: types.failure, data: n});
			return $.Deferred().reject(n);
		}
	);
};

function typesValid(types) {
	return ["started", "success", "failure"].every(key => {
		if (!types.hasOwnProperty(key)) {
			console.error(`Didn't provide action for "${key}" on a request`, types);
			return false;
		}
		return true;
	});
}

module.exports = AppDispatcher;
