// @flow

// NOTE: Actions with Flow types should go in this file. When all RepoActions have
// types, remove this file and move it back to RepoActions.js. This is so that we
// can progressively add Flow types without having to do it for all RepoActions
// at once.

export class WantRepos {
	querystring: string;

	constructor(querystring: string) {
		this.querystring = querystring;
	}
}

export class ReposFetched {
	querystring: string;
	data: any;

	constructor(querystring: string, data: any) {
		this.querystring = querystring;
		this.data = data;
	}
}
