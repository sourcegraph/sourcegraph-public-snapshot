// tslint:disable: typedef ordered-imports

import * as SearchActions from "sourcegraph/search/SearchActions";
import {SearchStore} from "sourcegraph/search/SearchStore";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {RESULTS_LIMIT} from "sourcegraph/search/GlobalSearch";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";

export const SearchBackend = {
	fetch: defaultFetch,

	__onDispatch(action) {
		switch (action.constructor) {

		case SearchActions.WantResults:
			{
				let p: SearchActions.WantResultsPayload = action.p;
				let results = SearchStore.get(p.query, p.repos, p.notRepos, p.commitID || null, p.limit);
				if (results === null) {
					let limit = p.limit || RESULTS_LIMIT;

					let q = [`Query=${encodeURIComponent(p.query)}`];
					q.push(`Limit=${limit}`);
					if (p.repos) {
						p.repos.forEach((repo) => q.push(`Repos=${encodeURIComponent(repo)}`));
					}
					if (p.notRepos) {
						p.notRepos.forEach((repo) => q.push(`NotRepos=${encodeURIComponent(repo)}`));
					}
					if (p.includeRepos) {
						q.push(`IncludeRepos=${encodeURIComponent(p.includeRepos.toString())}`);
					}
					if (p.commitID) {
						q.push(`CommitID=${encodeURIComponent(p.commitID)}`);
					}
					if (p.fast) {
						q.push(`Fast=1`);
					}

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
								includeRepos: p.includeRepos,
								defs: data,
								options: data.options,
							}));
						});
				}
				break;
			}
		}
	},
};

Dispatcher.Backends.register(SearchBackend.__onDispatch);
