import {Def} from "sourcegraph/def/index";

export type Action =
	WantResults |
	ResultsFetched;

export interface WantResultsPayload {
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

export interface ResultsFetchedPayload {
	query: string;
	repos: string[] | null;
	notRepos: string[] | null;
	commitID?: string;
	limit?: number;
	includeRepos?: boolean;
	defs: Def[];
	options: Object[];
}

export class ResultsFetched {
	p: ResultsFetchedPayload;

	constructor(params: ResultsFetchedPayload) {
		this.p = params;
	}
}
