import * as SearchActions from "./SearchActions";
// import SearchStore from "./SearchStore";
import Dispatcher from "../Dispatcher";
import defaultXhr from "xhr";

const SearchBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
		case SearchActions.WantResults:
			{
				let uri = `/.ui/${action.repo}/.search/${action.type}?q=${action.query}`;
				SearchBackend.xhr({
					uri: uri,
					json: {},
				}, function(err, resp, body) {
					if (err) {
						console.error(err);
						return;
					}
					Dispatcher.dispatch(new SearchActions.ResultsFetched(action.repo, action.rev, action.type, action.page, body));
				});
				break;
			}
		}
	},
};

Dispatcher.register(SearchBackend.__onDispatch);

export default SearchBackend;
