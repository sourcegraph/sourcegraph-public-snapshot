import * as SearchActions from "sourcegraph/search/SearchActions";
import SearchResultsStore from "sourcegraph/search/SearchResultsStore";
import Dispatcher from "sourcegraph/Dispatcher";
import defaultXhr from "xhr";

const SearchBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
		case SearchActions.WantResults:
			{
				let result = SearchResultsStore.results.get(action.repo, action.rev, action.query, action.type, action.page);
				if (result === null) {
					let uri;
					let revPart = action.rev ? `@${action.rev}` : "";
					if (action.type === "tokens" || action.type === "text")	{
						uri = `/.ui/${action.repo}${revPart}/.search/${action.type}?q=${action.query}&PerPage=${action.perPage}&Page=${action.page}`;
					} else {
						uri = `/${action.repo}${revPart}/.search/${action.type}?q=${action.query}&PerPage=${action.perPage}&Page=${action.page}`;
					}
					SearchBackend.xhr({
						uri: uri,
						json: {},
					}, function(err, resp, body) {
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.dispatch(new SearchActions.ResultsFetched(action.repo, action.rev, action.query, action.type, action.page, body));
					});
				}
				break;
			}
		}
	},
};

Dispatcher.register(SearchBackend.__onDispatch);

export default SearchBackend;
