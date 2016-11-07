export type Action =
	WantCommit |
	CommitFetched;

export class WantCommit {
	repo: string;
	rev: string;
	path: string;

	constructor(repo: string, rev: string, path: string) {
		this.repo = repo;
		this.rev = rev;
		this.path = path;
	}
}

export class CommitFetched {
	repo: string;
	rev: string;
	path: string;
	commit: string;

	constructor(repo: string, rev: string, path: string, commit: string) {
		this.repo = repo;
		this.rev = rev;
		this.path = path;
		this.commit = commit;
	}
}
