// @flow

import type {Def} from "sourcegraph/def";

export class WantResults {
	query: string;
	repos: ?Array<string>;
	notRepos: ?Array<string>;
	limit: ?number;
	prefixMatch: ?bool;
	includeRepos: ?bool;

	constructor(query: string, repos: ?Array<string>, notRepos: ?Array<string>, limit: ?number, prefixMatch: ?bool, includeRepos: ?bool) {
		this.query = query;
		this.repos = repos;
		this.notRepos = notRepos;
		this.limit = limit;
		this.prefixMatch = prefixMatch;
		this.includeRepos = includeRepos;
	}
}

export class ResultsFetched {
	query: string;
	repos: ?Array<string>;
	notRepos: ?Array<string>;
	limit: ?number;
	prefixMatch: ?bool;
	includeRepos: ?bool;
	defs: Array<Def>;

	constructor(query: string, repos: ?Array<string>, notRepos: ?Array<string>, limit: ?number, prefixMatch: ?bool, includeRepos: ?bool, defs: Array<Def>) {
		this.query = query;
		this.limit = limit;
		this.repos = repos;
		this.notRepos = notRepos;
		this.prefixMatch = prefixMatch;
		this.includeRepos = includeRepos;
		this.defs = defs;
	}
}
