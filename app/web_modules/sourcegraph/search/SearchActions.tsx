// tslint:disable

import {Def} from "sourcegraph/def/index";

export type WantResultsPayload = {
	query: string;
	repos: string[] | null;
	notRepos: string[] | null;
	commitID?: string;
	limit: number;
	includeRepos?: boolean;
	fast?: boolean;
}

export class WantResults {
	p: WantResultsPayload;

	constructor(params: WantResultsPayload) {
		this.p = params;
	}
}

export type ResultsFetchedPayload = {
	query: string;
	repos: string[] | null;
	notRepos: string[] | null;
	commitID?: string;
	limit?: number;
	includeRepos?: boolean;
	defs: Array<Def>;
	options: Array<Object>;
}

export class ResultsFetched {
	p: ResultsFetchedPayload;

	constructor(params: ResultsFetchedPayload) {
		this.p = params;
	}
}
