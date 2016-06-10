// @flow

import type {Error} from "sourcegraph/Error";
import type {DeltaFiles} from "sourcegraph/delta";

export class WantFiles {
	baseRepo: number;
	baseRev: string;
	headRepo: number;
	headRev: string;

	constructor(baseRepo: number, baseRev: string, headRepo: number, headRev: string) {
		this.baseRepo = baseRepo;
		this.baseRev = baseRev;
		this.headRepo = headRepo;
		this.headRev = headRev;
	}
}

export class FetchedFiles {
	baseRepo: number;
	baseRev: string;
	headRepo: number;
	headRev: string;
	data: DeltaFiles | Error;

	constructor(baseRepo: number, baseRev: string, headRepo: number, headRev: string, data: DeltaFiles | Error) {
		this.baseRepo = baseRepo;
		this.baseRev = baseRev;
		this.headRepo = headRepo;
		this.headRev = headRev;
		this.data = data;
	}
}
