var globals = require("../globals");
var SearchUtil = require("../util/SearchUtil");
var AppDispatcher = require("../dispatchers/AppDispatcher");

module.exports.searchRepo = (query, repo) => {
	AppDispatcher.handleViewAction({
		type: globals.Actions.SEARCH_SUBMIT,
		query: query,
		repoURI: repo.URI,
	});

	AppDispatcher.dispatchAsync(SearchUtil.fetchTokenResults(query, repo.URI), {
		started: null,
		success: globals.Actions.SEARCH_RECEIVED_TOKEN_RESULTS,
		failure: null,
	});
};
