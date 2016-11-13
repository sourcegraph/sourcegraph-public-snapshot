import {RepoList, SymbolInformation} from "sourcegraph/api";

export type Action =
	WantRepos | ReposFetched |
	WantSymbols | FetchedSymbols;

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
