var globals = require("../globals");
var AppDispatcher = require("../dispatchers/AppDispatcher");

var SearchResultsStore = {
	state: {
		query: null,
		textResults: null,
		tokenResults: null,
	},

	onChange: new Event("SearchResultsStoreChange"),

	dispatchToken: AppDispatcher.register((payload) => {
		switch (payload.action.type) {
		case globals.Actions.SEARCH_SUBMIT:
			SearchResultsStore.state.query = payload.action.query;
			break;
		}
		window.dispatchEvent(SearchResultsStore.onChange);
	}),
};

module.exports = SearchResultsStore;
