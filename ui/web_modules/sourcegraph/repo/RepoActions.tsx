import {Repo, RepoList, SymbolInformation} from "sourcegraph/api";

export type Action =
	RepoCloning |
	WantRepo | FetchedRepo |
	WantRepos | ReposFetched |
	WantSymbols | FetchedSymbols;

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

export class RepoCloning {
	repo: string;
	isCloning: boolean;

	constructor(repo: string, isCloning: boolean) {
		this.repo = repo;
		this.isCloning = isCloning;
	}
}

export class WantSymbols {
	languages: string[];
	repo: string;
	rev: string;
	query: string;

	constructor(languages: string[], repo: string, rev: string, query: string) {
		this.languages = languages;
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
