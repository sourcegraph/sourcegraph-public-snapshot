// @flow

import type {Def} from "sourcegraph/def";

export class WantResults {
	query: string;
	repos: ?Array<string>;
	notRepos: ?Array<string>;
	limit: ?number;

	constructor(query: string, repos: ?Array<string>, notRepos: ?Array<string>, limit: ?number) {
		this.query = query;
		this.repos = repos;
		this.notRepos = notRepos;
		this.limit = limit;
	}
}

export class ResultsFetched {
	query: string;
	repos: ?Array<string>;
	notRepos: ?Array<string>;
	limit: ?number;
	defs: Array<Def>;

	constructor(query: string, repos: ?Array<string>, notRepos: ?Array<string>, limit: ?number, defs: Array<Def>) {
		this.query = query;
		this.limit = limit;
		this.repos = repos;
		this.notRepos = notRepos;
		this.defs = defs;
	}
}
