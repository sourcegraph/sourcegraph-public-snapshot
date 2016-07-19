export class WantRepo {
	constructor(repo) {
		this.repo = repo;
	}
}

export class FetchedRepo {
	constructor(repo, repoObj) {
		this.repo = repo;
		this.repoObj = repoObj;
	}
}

export class WantResolveRev {
	constructor(repo, rev, force) {
		this.repo = repo;
		this.rev = rev;
		this.force = force;
	}
}

export class ResolvedRev {
	constructor(repo, rev, commitID) {
		this.repo = repo;
		this.rev = rev;
		this.commitID = commitID;
	}
}

export class WantCommit {
	constructor(repo, rev) {
		this.repo = repo;
		this.rev = rev;
	}
}

export class FetchedCommit {
	constructor(repo, rev, commit) {
		this.repo = repo;
		this.rev = rev;
		this.commit = commit;
	}
}

export class WantInventory {
	constructor(repo, commitID) {
		this.repo = repo;
		this.commitID = commitID;
	}
}

export class FetchedInventory {
	constructor(repo, commitID, inventory) {
		this.repo = repo;
		this.commitID = commitID;
		this.inventory = inventory;
	}
}

export class RepoCloning {
	constructor(repo, isCloning) {
		this.repo = repo;
		this.isCloning = isCloning;
	}
}

export class WantResolveRepo {
	constructor(repo) {
		this.repo = repo;
	}
}

export class RepoResolved {
	constructor(repo, resolution) {
		this.repo = repo;
		this.resolution = resolution;
	}
}

export class WantCreateRepo {
	constructor(repo, remoteRepo) {
		this.repo = repo;
		this.remoteRepo = remoteRepo;
	}
}

export class RepoCreated {
	constructor(repo, repoObj) {
		this.repo = repo;
		this.repoObj = repoObj;
	}
}

export class WantBranches {
	constructor(repo) {
		this.repo = repo;
	}
}

export class FetchedBranches {
	constructor(repo, branches) {
		this.repo = repo;
		this.branches = branches;
	}
}

export class WantTags {
	constructor(repo) {
		this.repo = repo;
	}
}

export class FetchedTags {
	constructor(repo, tags) {
		this.repo = repo;
		this.tags = tags;
	}
}

export class RefreshVCS {
	constructor(repo) {
		this.repo = repo;
	}
}
