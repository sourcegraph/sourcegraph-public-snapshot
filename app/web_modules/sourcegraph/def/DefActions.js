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

	constructor(repo: string, rev: ?string, def: string, defObj: Def) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.defObj = defObj;
	}
}

export class WantDefs {
	repo: string;
	rev: ?string;
	query: string;

	constructor(repo: string, rev: ?string, query: string) {
		this.repo = repo;
		this.rev = rev;
		this.query = query;
	}
}

export class DefsFetched {
	repo: string;
	rev: ?string;
	query: string;
	defs: Array<Def>;

	constructor(repo: string, rev: ?string, query: string, defs: Array<Def>) {
		this.repo = repo;
		this.rev = rev;
		this.query = query;
		this.defs = defs;
	}
}

export class SelectDef {
	repo: string;
	rev: ?string;
	def: string;

	constructor(repo: string, rev: ?string, def: string) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
	}
}

export class HighlightDef {
	url: ?string;

	constructor(url: ?string) {
		this.url = url;
	}
}

export class WantRefs {
	repo: string;
	rev: ?string;
	def: string;
	file: ?string; // only return refs in this file

	constructor(repo: string, rev: ?string, def: string, file: ?string) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.file = file || null;
	}
}

export class RefsFetched {
	repo: string;
	rev: ?string;
	def: string;
	file: ?string;
	refs: Array<Ref>;

	constructor(repo: string, rev: ?string, def: string, file: ?string, refs: Array<Ref>) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.file = file || null;
		this.refs = refs;
	}
}
