export class WantCommit {
	constructor(repo, rev, path) {
		this.repo = repo;
		this.rev = rev;
		this.path = path;
	}
}

export class CommitFetched {
	constructor(repo, rev, path, commit) {
		this.repo = repo;
		this.rev = rev;
		this.path = path;
		this.commit = commit;
	}
}
