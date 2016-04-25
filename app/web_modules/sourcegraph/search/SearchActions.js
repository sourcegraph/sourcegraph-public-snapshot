// @flow

import type {Def} from "sourcegraph/def";

const DEFAULT_LIMIT = 20;

export class WantResults {
	query: string;
	repos: ?Array<string>;
	notRepos: ?Array<string>;
	limit: number;

	constructor(query: string, repos: ?Array<string>, notRepos: ?Array<string>, limit: ?number) {
		this.query = query;
		this.repos = repos;
		this.notRepos = notRepos;
		this.limit = limit;
	}
}

export class ResultsFetched {
	query: string;
	defs: Array<Def>;
	eventName: string;

	constructor(query: string, defs: Array<Def>) {
		this.query = query;
		this.defs = defs;
		this.eventName = "ResultsFetched";
	}
}
