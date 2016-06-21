// @flow

import type {Def} from "sourcegraph/def";

export type WantResultsPayload = {
	query: string;
	repos?: ?Array<string>;
	notRepos?: ?Array<string>;
	commitID?: string;
	limit: number;
	prefixMatch?: bool;
	includeRepos?: bool;
	fast?: bool;
}

export class WantResults {
	p: WantResultsPayload;

	constructor(params: WantResultsPayload) {
		this.p = params;
	}
}

export type ResultsFetchedPayload = {
	query: string;
	repos?: ?Array<string>;
	notRepos?: ?Array<string>;
	commitID?: string;
	limit?: number;
	prefixMatch?: bool;
	includeRepos?: bool;
	defs: Array<Def>;
	options: Array<Object>;
}

export class ResultsFetched {
	p: ResultsFetchedPayload;

	constructor(params: ResultsFetchedPayload) {
		this.p = params;
	}
}
