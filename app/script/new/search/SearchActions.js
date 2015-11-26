export class WantResults {
	constructor(repo, rev, type, page, perPage, query) {
		this.repo = repo;
		this.rev = rev;
		this.type = type;
		this.page = page;
		this.perPage = perPage;
		this.query = query;
	}
}

export class ResultsFetched {
	constructor(repo, rev, query, type, page, results) {
		this.repo = repo;
		this.rev = rev;
		this.query= query;
		this.type = type;
		this.page = page;
		this.results = results;
	}
}

export class SelectResultType {
	constructor(type) {
		this.type = type;
	}
}
