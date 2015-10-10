var globals = require("../globals");
var AppDispatcher = require("../dispatchers/AppDispatcher");

var SearchResultsStore = {
	state: {
		query: null,
		repo: null,
		textSearch: null,
		textSearchLoading: false,
		tokenSearch: null,
		tokenSearchLoading: false,
	},

	onChange: new Event("SearchResultsStoreChange"),

	dispatchToken: AppDispatcher.register((payload) => {
		switch (payload.action.type) {
		case globals.Actions.SEARCH_SUBMIT:
			SearchResultsStore.state.query = payload.action.query;
			SearchResultsStore.state.repo = payload.action.repo;
			SearchResultsStore.state.tokenSearch = null;
			SearchResultsStore.state.textSearch = null;
			break;
		case globals.Actions.SEARCH_TOKENS_SUBMIT:
			SearchResultsStore.state.tokenSearchLoading = true;
			break;
		case globals.Actions.SEARCH_TOKENS_RECEIVED_RESULTS:
			SearchResultsStore.state.tokenSearchLoading = false;
			SearchResultsStore.state.tokenSearch = payload.action.data;
			break;
		}
		window.dispatchEvent(SearchResultsStore.onChange);
	}),
};

module.exports = SearchResultsStore;
