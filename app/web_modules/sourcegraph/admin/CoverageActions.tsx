// tslint:disable

export class WantCoverage {
	constructor() {}
}

export class CoverageFetched {
	data: any;

	constructor(data) {
		this.data = data;
	}
}
