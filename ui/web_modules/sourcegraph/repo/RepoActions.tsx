import {Branch, Commit, Repo, RepoList, RepoResolution, SymbolInformation, Tag} from "sourcegraph/api";

export type Action =
	RepoCloning |
	RefreshVCS |
	WantBranches | FetchedBranches |
	WantCreateRepo | RepoCreated |
	WantCommit | FetchedCommit |
	WantInventory | FetchedInventory |
	WantRepo | FetchedRepo |
	WantRepos | ReposFetched |
	WantResolveRepo | RepoResolved |
	WantResolveRev | ResolvedRev |
	WantSymbols | FetchedSymbols |
	WantTags | FetchedTags;

export class WantRepo {
	repo: string;

	constructor(repo: string) {
		this.repo = repo;
	}
}

export class FetchedRepo {
	repo: string;
	repoObj: Repo;

	constructor(repo: string, repoObj: Repo) {
		this.repo = repo;
		this.repoObj = repoObj;
	}
}

export class WantRepos {
	querystring: string;
	isUserRepos?: boolean;

	constructor(querystring: string, isUserRepos?: boolean) {
		this.querystring = querystring;
		this.isUserRepos = isUserRepos;
	}
}

export class ReposFetched {
	querystring: string;
	isUserRepos: boolean;
	data: RepoList;

	constructor(querystring: string, data: RepoList, isUserRepos: boolean) {
		this.querystring = querystring;
		this.data = data;
		this.isUserRepos = isUserRepos;
	}
}

export class WantResolveRev {
	repo: string;
	rev: string;
	force: boolean;

	constructor(repo: string, rev: string, force?: boolean) {
		this.repo = repo;
		this.rev = rev;
		this.force = force || false;
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

export class RepoResolved {
	repo: string;
	resolution: RepoResolution;

	constructor(repo: string, resolution: RepoResolution) {
		this.repo = repo;
		this.resolution = resolution;
	}
}

export class WantCreateRepo {
	repo: string;
	remoteRepo: any;
	refreshVCS: boolean;

	constructor(repo: string, remoteRepo: any, refreshVCS?: boolean) {
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

export class WantSymbols {
	repo: string;
	rev: string;
	query: string;

	constructor(repo: string, rev: string, query: string) {
		this.repo = repo;
		this.rev = rev;
		this.query = query;
	}
}

export class FetchedSymbols {
	mode: string;
	repo: string;
	rev: string;
	query: string;

	symbols: SymbolInformation[];

	constructor(mode: string, repo: string, rev: string, query: string, symbols: SymbolInformation[]) {
		this.mode = mode;
		this.repo = repo;
		this.rev = rev;
		this.query = query;
		this.symbols = symbols;
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
