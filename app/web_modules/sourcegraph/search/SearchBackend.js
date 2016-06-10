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

				let results = SearchStore.results.get(action.query, action.repos, action.notRepos, action.limit, action.prefixMatch, action.includeRepos);
				if (results === null) {
					let limit = action.limit || RESULTS_LIMIT;

					let q = [`Query=${encodeURIComponent(action.query)}`];
					q.push(`Limit=${limit}`);
					if (action.repos) {
						q.push(`Repos=${encodeURIComponent(action.repos)}`);
					}
					if (action.notRepos) {
						q.push(`NotRepos=${encodeURIComponent(action.notRepos)}`);
					}
					if (action.prefixMatch) {
						q.push(`PrefixMatch=${encodeURIComponent(action.prefixMatch)}`);
					}
					if (action.includeRepos) {
						q.push(`IncludeRepos=${encodeURIComponent(action.includeRepos)}`);
					}
					trackPromise(
						SearchBackend.fetch(`/.api/global-search?${q.join("&")}`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => {
								Dispatcher.Stores.dispatch(new SearchActions.ResultsFetched(action.query, action.repos, action.notRepos, action.limit, action.prefixMatch, action.includeRepos, data));
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
