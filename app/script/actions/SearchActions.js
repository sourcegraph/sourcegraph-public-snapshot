var globals = require("../globals");
var SearchUtil = require("../util/SearchUtil");
var AppDispatcher = require("../dispatchers/AppDispatcher");

module.exports.selectSearchType = (searchType) => {
	AppDispatcher.handleViewAction({
		type: globals.Actions.SEARCH_SELECT_TYPE,
		searchType: searchType,
	});
};

module.exports.searchRepo = (query, repo) => {
	AppDispatcher.handleViewAction({
		type: globals.Actions.SEARCH_SUBMIT,
		query: query,
		repo: repo,
	});
	module.exports.searchRepoTokens(query, repo, 1);
	module.exports.searchRepoText(query, repo, 1);
};

module.exports.searchRepoTokens = (query, repo, page) => {
	AppDispatcher.dispatchAsync(SearchUtil.fetchTokenResults(query, repo.URI, page), {
		started: globals.Actions.SEARCH_TOKENS_SUBMIT,
		success: globals.Actions.SEARCH_TOKENS_RECEIVED_RESULTS,
		failure: null,
	});
};

module.exports.searchRepoText = (query, repo, page) => {
	AppDispatcher.dispatchAsync(SearchUtil.fetchTextResults(query, repo.URI, page), {
		started: globals.Actions.SEARCH_TEXT_SUBMIT,
		success: globals.Actions.SEARCH_TEXT_RECEIVED_RESULTS,
		failure: null,
	});
};
