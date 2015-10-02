var globals = require("../globals");
var AppDispatcher = require("../dispatchers/AppDispatcher");

var SearchResultsStore = {
	state: {
		query: null,
		textSearch: null,
		tokenSearch: null,
	},

	onChange: new Event("SearchResultsStoreChange"),

	dispatchToken: AppDispatcher.register((payload) => {
		switch (payload.action.type) {
		case globals.Actions.SEARCH_SUBMIT:
			SearchResultsStore.state.query = payload.action.query;
			SearchResultsStore.state.textSearch = null;
			SearchResultsStore.state.tokenSearch = null;
			break;
		case globals.Actions.SEARCH_RECEIVED_TOKEN_RESULTS:
			SearchResultsStore.state.tokenSearch = payload.action.data;
			break;
		}
		window.dispatchEvent(SearchResultsStore.onChange);
	}),
};

module.exports = SearchResultsStore;
