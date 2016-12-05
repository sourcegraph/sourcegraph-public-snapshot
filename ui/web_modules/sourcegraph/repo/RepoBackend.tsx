import * as Dispatcher from "sourcegraph/Dispatcher";
import * as lsp from "sourcegraph/editor/lsp";
import { languagesToSearchModes } from "sourcegraph/editor/modes";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import { RepoStore } from "sourcegraph/repo/RepoStore";
import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";

export const OriginGitHub = 0; // Origin.ServiceType enum value for GitHub origin

const workspaceSymbolFlights = new Set<string>();

export const RepoBackend = {
	fetch: singleflightFetch(defaultFetch),

	__onDispatch(payload: RepoActions.Action): void {
		if (payload instanceof RepoActions.WantRepos) {
			const action = payload;
			const repos = RepoStore.repos.list(action.querystring);
			if (repos === null) {
				const url = `/.api/repos?${action.querystring}`;
				RepoBackend.fetch(url)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({ Error: err }))
					.then((data) => {
						Dispatcher.Stores.dispatch(new RepoActions.ReposFetched(action.querystring, data, action.isUserRepos || false));
					});
			}
		}

		if (payload instanceof RepoActions.WantSymbols) {
			const action = payload;
			let symbols = RepoStore.symbols.list(action.languages, action.repo, action.rev, action.query);
			if (symbols.results.length > 0) {
				return;
			}
			languagesToSearchModes(action.languages).forEach(mode => {
				const url = `git:\/\/${action.repo}?${action.rev}`;
				const flightKey = `${url}@${mode}?${action.query}`;
				if (workspaceSymbolFlights.has(flightKey)) {
					return;
				}
				const method = action.prepare ? "workspace/symbol?prepare" : "workspace/symbol";
				workspaceSymbolFlights.add(flightKey);
				lsp.sendExt(url, mode, method, { query: action.query, limit: 100 })
					.then((r) => {
						let result;
						if (r === null || !r.result || !r.result.length) {
							result = [];
						} else {
							result = r.result;
						}
						Dispatcher.Stores.dispatch(
							new RepoActions.FetchedSymbols(mode, action.repo, action.rev, action.query, result)
						);
						workspaceSymbolFlights.delete(flightKey);
					});
			});
		}
	},
};

Dispatcher.Backends.register(RepoBackend.__onDispatch);
