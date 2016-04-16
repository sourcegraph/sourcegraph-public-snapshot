// @flow

import type {Def} from "sourcegraph/def";

export class WantResults {
	query: string;

	constructor(query: string) {
		this.query = query;
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
