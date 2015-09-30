var globals = require("../globals");
var AppDispatcher = require("../dispatchers/AppDispatcher");

module.exports.searchRepo = (query, repo) => {
	AppDispatcher.handleViewAction({
		type: globals.Actions.SEARCH_SUBMIT,
		query: query,
		repoURI: repo.URI,
	});
};
