export class WantBranches {
	constructor(repo) {
		this.repo = repo;
	}
}

export class FetchedBranches {
	constructor(repo, branches, err) {
		this.repo = repo;
		this.branches = branches;
		this.err = err;
	}
}

export class WantTags {
	constructor(repo) {
		this.repo = repo;
	}
}

export class FetchedTags {
	constructor(repo, tags, err) {
		this.repo = repo;
		this.tags = tags;
		this.err = err;
	}
}
