// @flow

import type {Def} from "sourcegraph/def";

export class WantResults {
	query: string;
	repos: ?Array<string>;
	notRepos: ?Array<string>;

	constructor(query: string, repos: ?Array<string>, notRepos: ?Array<string>) {
		this.query = query;
		this.repos = repos;
		this.notRepos = notRepos;
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
