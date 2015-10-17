var globals = require("../globals");
var AppDispatcher = require("../dispatchers/AppDispatcher");

var SearchResultsStore = {
	state: {
		query: null,
		repo: null,
		currentSearchType: globals.SearchType.TOKEN,
		textSearch: {
			data: null,
			loading: false,
			error: null,
		},
		tokenSearch: {
			data: null,
			loading: false,
			error: null,
		},
	},

	onChange: new Event("SearchResultsStoreChange"),

	dispatchToken: AppDispatcher.register((payload) => {
		switch (payload.action.type) {
		case globals.Actions.SEARCH_SUBMIT:
			SearchResultsStore.state.query = payload.action.query;
			SearchResultsStore.state.repo = payload.action.repo;
			SearchResultsStore.state.tokenSearch.data = null;
			SearchResultsStore.state.textSearch.data = null;
			break;
		case globals.Actions.SEARCH_SELECT_TYPE:
			SearchResultsStore.state.currentSearchType = payload.action.searchType;
			break;
		case globals.Actions.SEARCH_TOKENS_SUBMIT:
			SearchResultsStore.state.tokenSearch.loading = true;
			break;
		case globals.Actions.SEARCH_TOKENS_RECEIVED_RESULTS:
			SearchResultsStore.state.tokenSearch.loading = false;
			SearchResultsStore.state.tokenSearch.data = payload.action.data;
			break;
		case globals.Actions.SEARCH_TOKENS_FAILURE:
			SearchResultsStore.state.tokenSearch.loading = false;
			SearchResultsStore.state.tokenSearch.error = payload.action.data;
			break;
		case globals.Actions.SEARCH_TEXT_SUBMIT:
			SearchResultsStore.state.textSearch.loading = true;
			break;
		case globals.Actions.SEARCH_TEXT_RECEIVED_RESULTS:
			SearchResultsStore.state.textSearch.loading = false;
			SearchResultsStore.state.textSearch.data = payload.action.data;
			break;
		case globals.Actions.SEARCH_TEXT_FAILURE:
			SearchResultsStore.state.textSearch.loading = false;
			SearchResultsStore.state.textSearch.error = payload.action.data;
			break;
		}
		window.dispatchEvent(SearchResultsStore.onChange);
	}),
};

module.exports = SearchResultsStore;
