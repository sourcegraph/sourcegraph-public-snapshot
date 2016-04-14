// @flow

import type {Def, Ref} from "sourcegraph/def";

export class WantDef {
	repo: string;
	rev: ?string;
	def: string;

	constructor(repo: string, rev: ?string, def: string) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
	}
}

export class DefFetched {
	repo: string;
	rev: ?string;
	def: string;
	defObj: Def;
	eventName: string;

	constructor(repo: string, rev: ?string, def: string, defObj: Def) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.defObj = defObj;
		this.eventName = "DefFetched";
	}
}

export class WantDefAuthors {
	repo: string;
	rev: ?string;
	def: string;

	constructor(repo: string, rev: ?string, def: string) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
	}
}

export class DefAuthorsFetched {
	repo: string;
	rev: ?string;
	def: string;
	authors: Object;

	constructor(repo: string, rev: ?string, def: string, authors: Object) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.authors = authors;
	}
}

export class WantDefs {
	repo: string;
	rev: ?string;
	query: string;
	filePathPrefix: ?string;

	constructor(repo: string, rev: ?string, query: string, filePathPrefix?: string) {
		this.repo = repo;
		this.rev = rev;
		this.query = query;
		this.filePathPrefix = filePathPrefix || null;
	}
}

export class DefsFetched {
	repo: string;
	rev: ?string;
	query: string;
	defs: Array<Def>;
	filePathPrefix: ?string;
	eventName: string;

	constructor(repo: string, rev: ?string, query: string, filePathPrefix: ?string, defs: Array<Def>) {
		this.repo = repo;
		this.rev = rev;
		this.query = query;
		this.filePathPrefix = filePathPrefix;
		this.defs = defs;
		this.eventName = "DefsFetched";
	}
}

export class SelectDef {
	repo: string;
	rev: ?string;
	def: string;
	eventName: string;

	constructor(repo: string, rev: ?string, def: string) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.eventName = "SelectDef";
	}
}

export class HighlightDef {
	url: ?string;
	eventName: string;

	constructor(url: ?string) {
		this.url = url;
		this.eventName = "HighlightDef";
	}
}

export class WantRefLocations {
	repo: string;
	rev: ?string;
	def: string;

	constructor(repo: string, rev: ?string, def: string) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
	}
}

export class RefLocationsFetched {
	repo: string;
	rev: ?string;
	def: string;
	locations: Array<Object>;

	constructor(repo: string, rev: ?string, def: string, locations: Array<Object>) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.locations = locations;
	}
}

export class WantRefs {
	repo: string;
	rev: ?string;
	def: string;
	refRepo: string; // return refs from files in this repo
	refFile: ?string; // only return refs in this file

	constructor(repo: string, rev: ?string, def: string, refRepo: string, refFile: ?string) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.refRepo = refRepo;
		this.refFile = refFile || null;
	}
}

export class RefsFetched {
	repo: string;
	rev: ?string;
	def: string;
	refRepo: string;
	refFile: ?string;
	refs: Array<Ref>;
	eventName: string;

	constructor(repo: string, rev: ?string, def: string, refRepo: string, refFile: ?string, refs: Array<Ref>) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.refRepo = refRepo;
		this.refFile = refFile || null;
		this.refs = refs;
		this.eventName = "RefsFetched";
	}
}
