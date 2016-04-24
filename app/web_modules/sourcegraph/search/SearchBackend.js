// @flow weak

import * as SearchActions from "sourcegraph/search/SearchActions";
import SearchStore from "sourcegraph/search/SearchStore";
import Dispatcher from "sourcegraph/Dispatcher";
import {RESULTS_LIMIT} from "sourcegraph/search/GlobalSearch";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import {trackPromise} from "sourcegraph/app/status";

const SearchBackend = {
	fetch: defaultFetch,

	__onDispatch(action) {
		switch (action.constructor) {

		case SearchActions.WantResults:
			{
				if (!action.query || action.query === "") {
					break;
				}
				let results = SearchStore.results.get(action.query);
				if (results === null) {
					trackPromise(
						SearchBackend.fetch(`/.api/global-search?Query=${encodeURIComponent(action.query)}&Limit=${RESULTS_LIMIT}`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => {
								Dispatcher.Stores.dispatch(new SearchActions.ResultsFetched(action.query, data));
							})
					);
				}
				break;
			}
		}
	},
};

Dispatcher.Backends.register(SearchBackend.__onDispatch);

export default SearchBackend;
