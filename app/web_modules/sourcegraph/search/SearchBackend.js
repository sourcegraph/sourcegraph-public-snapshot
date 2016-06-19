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
				let p: SearchActions.WantResultsPayload = action.p;
				let results = SearchStore.get(p.query, p.repos, p.notRepos, p.commitID, p.limit);
				// let results = SearchStore.get(p.query, p.repos, p.notRepos, p.commitID, p.limit, p.prefixMatch, p.includeRepos);
				if (results === null) {
					let limit = p.limit || RESULTS_LIMIT;

					let q = [`Query=${encodeURIComponent(p.query)}`];
					q.push(`Limit=${limit}`);
					if (p.repos) {
						q.push(`Repos=${encodeURIComponent(p.repos.toString())}`);
					}
					if (p.notRepos) {
						q.push(`NotRepos=${encodeURIComponent(p.notRepos.toString())}`);
					}
					if (p.prefixMatch) {
						q.push(`PrefixMatch=${encodeURIComponent(p.prefixMatch.toString())}`);
					}
					if (p.includeRepos) {
						q.push(`IncludeRepos=${encodeURIComponent(p.includeRepos.toString())}`);
					}
					if (p.commitID) {
						q.push(`CommitID=${encodeURIComponent(p.commitID)}`);
					}

					trackPromise(
						SearchBackend.fetch(`/.api/global-search?${q.join("&")}`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => {
								Dispatcher.Stores.dispatch(new SearchActions.ResultsFetched({
									query: p.query,
									repos: p.repos,
									notRepos: p.notRepos,
									commitID: p.commitID,
									limit: p.limit,
									prefixMatch: p.prefixMatch,
									includeRepos: p.includeRepos,
									defs: data,
									options: data.options,
								}));
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
