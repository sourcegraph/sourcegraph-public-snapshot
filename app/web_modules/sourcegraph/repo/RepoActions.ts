import {Branch, Tag} from "sourcegraph/repo/vcs";

export class WantRepo {
	repo: string;

	constructor(repo: string) {
		this.repo = repo;
	}
}

export interface Repo {}; // incomplete

export class FetchedRepo {
	repo: string;
	repoObj: Repo;

	constructor(repo: string, repoObj: Repo) {
		this.repo = repo;
		this.repoObj = repoObj;
	}
}

export class WantResolveRev {
	repo: string;
	rev: string;
	force: boolean;

	constructor(repo: string, rev: string, force: boolean) {
		this.repo = repo;
		this.rev = rev;
		this.force = force;
	}
}

export class ResolvedRev {
	repo: string;
	rev: string;
	commitID: string;

	constructor(repo: string, rev: string, commitID: string) {
		this.repo = repo;
		this.rev = rev;
		this.commitID = commitID;
	}
}

export class WantCommit {
	repo: string;
	rev: string;

	constructor(repo: string, rev: string) {
		this.repo = repo;
		this.rev = rev;
	}
}

export interface Commit {}; // incomplete

export class FetchedCommit {
	repo: string;
	rev: string;
	commit: Commit;

	constructor(repo: string, rev: string, commit: Commit) {
		this.repo = repo;
		this.rev = rev;
		this.commit = commit;
	}
}

export class WantInventory {
	repo: string;
	commitID: string;

	constructor(repo: string, commitID: string) {
		this.repo = repo;
		this.commitID = commitID;
	}
}

export interface Inventory {}; // incomplete

export class FetchedInventory {
	repo: string;
	commitID: string;
	inventory: Inventory;

	constructor(repo: string, commitID: string, inventory: Inventory) {
		this.repo = repo;
		this.commitID = commitID;
		this.inventory = inventory;
	}
}

export class RepoCloning {
	repo: string;
	isCloning: boolean;

	constructor(repo: string, isCloning: boolean) {
		this.repo = repo;
		this.isCloning = isCloning;
	}
}

export class WantResolveRepo {
	repo: string;

	constructor(repo: string) {
		this.repo = repo;
	}
}

export interface Resolution {}; // incomplete

export class RepoResolved {
	repo: string;
	resolution: Resolution;

	constructor(repo: string, resolution: Resolution) {
		this.repo = repo;
		this.resolution = resolution;
	}
}

export class WantCreateRepo {
	repo: string;
	remoteRepo: string;
	refreshVCS: boolean;

	constructor(repo: string, remoteRepo: string, refreshVCS?: boolean) {
		this.repo = repo;
		this.remoteRepo = remoteRepo;
		// Settings this option to true will cause the newly create repo to be
		// automatically cloned.
		this.refreshVCS = refreshVCS || false;
	}
}

export class RepoCreated {
	repo: string;
	repoObj: Repo;

	constructor(repo: string, repoObj: Repo) {
		this.repo = repo;
		this.repoObj = repoObj;
	}
}

export class WantBranches {
	repo: string;

	constructor(repo: string) {
		this.repo = repo;
	}
}

export class FetchedBranches {
	repo: string;
	branches: Branch[];

	constructor(repo: string, branches: Branch[]) {
		this.repo = repo;
		this.branches = branches;
	}
}

export class WantTags {
	repo: string;

	constructor(repo: string) {
		this.repo = repo;
	}
}

export class FetchedTags {
	repo: string;
	tags: Tag[];

	constructor(repo: string, tags: Tag[]) {
		this.repo = repo;
		this.tags = tags;
	}
}

export class RefreshVCS {
	repo: string;

	constructor(repo: string) {
		this.repo = repo;
	}
}
